package alice

import (
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/spf13/viper"
)

// EC2MetadataAPI is an interface allowing mocks of the AWS client
type EC2MetadataAPI interface {
	GetMetadata(string) (string, error)
}

// AWSInventory is an inventory of AWS EC2 instances in an autoscaling group
type AWSInventory struct {
	log            *logrus.Entry
	Config         *viper.Viper
	AutoscalingSvc autoscalingiface.AutoScalingAPI
	EC2metadataSvc EC2MetadataAPI
	groupName      string
	metadata       AWSMetadata
	lastModified   time.Time
}

// AWSMetadata provides region and instance id metadata
type AWSMetadata struct {
	region       string
	regionWithAZ string
	instanceID   string
}

const (
	defaultAWSRegion        = "eu-west-1"
	defaultSettleDownPeriod = "0s"
)

// NewAWSInventory creates a new AWSInventory
func NewAWSInventory(config *viper.Viper, log *logrus.Entry) (Inventory, error) {
	config.SetDefault("region", defaultAWSRegion)
	config.SetDefault("settle_down_period", defaultSettleDownPeriod)
	s, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	region := config.GetString("region")
	s.Config.Region = &region
	inv := AWSInventory{
		AutoscalingSvc: autoscaling.New(s),
		EC2metadataSvc: ec2metadata.New(s),
		log:            log,
		Config:         config,
	}
	return &inv, nil
}

// Total returns the current total number of resources
func (a *AWSInventory) Total() (int, error) {
	name := a.GroupName()
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&name},
	}
	group := a.describeAutoScalingGroups(params)
	return len(group.Instances), nil
}

// Increase (scale up) the number of resources in the inventory
func (a *AWSInventory) Increase() error {
	return a.Scale(+1)
}

// Decrease (scale down) the number of resources in the inventory
func (a *AWSInventory) Decrease() error {
	return a.Scale(-1)
}

// Status returns OK if the inventory is ready to be scaled, UPDATING if an update is in progress, or FAILED
func (a *AWSInventory) Status() (Status, error) {
	params := &autoscaling.DescribeScalingActivitiesInput{AutoScalingGroupName: aws.String(a.GroupName())}
	status := OK
	done := false
	for !done {
		resp, err := a.AutoscalingSvc.DescribeScalingActivities(params)
		if err != nil {
			return status, err
		}
		a.log.Debugf("Checking %d pre-existing scaling activites", len(resp.Activities))
		for _, activity := range resp.Activities {
			switch *activity.StatusCode {
			case autoscaling.ScalingActivityStatusCodeSuccessful:
				continue
			case autoscaling.ScalingActivityStatusCodeCancelled:
				a.log.Debugln("Ignoring a cancelled activity")
				continue
			case autoscaling.ScalingActivityStatusCodeFailed:
				a.log.Debugln("Found a failed activity")
				return FAILED, nil
			default:
				a.log.Debugln("Found an in-progress activity")
				status = UPDATING
			}
		}
		if resp.NextToken == nil {
			done = true
		} else {
			params.NextToken = resp.NextToken
		}
	}
	if status == OK && time.Now().Before(a.lastModified.Add(a.Config.GetDuration("settle_down_period"))) {
		a.log.Debugln("Still within settle down period")
		status = UPDATING
	}
	return status, nil
}

func (a *AWSInventory) describeAutoScalingGroups(params *autoscaling.DescribeAutoScalingGroupsInput) *autoscaling.Group {
	a.RefreshMetadata()
	var group *autoscaling.Group
	done := false
	for !done {
		resp, err := a.AutoscalingSvc.DescribeAutoScalingGroups(params)
		if err != nil {
			a.log.Fatalf("describeAutoScalingGroups: %v", err)
		}

		for _, scaleGroup := range resp.AutoScalingGroups {
			for _, server := range scaleGroup.Instances {
				if *server.InstanceId == a.metadata.instanceID {
					group = scaleGroup
				}
			}
		}
		if resp.NextToken == nil {
			done = true
		} else {
			params.NextToken = resp.NextToken
		}
	}

	if group == nil {
		a.log.Fatal("No auto scaling group available")
	}
	return group
}

// GroupName returns the autoscaling group for this inventory
func (a *AWSInventory) GroupName() string {
	if a.groupName == "" {
		group := a.describeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
		a.groupName = *group.AutoScalingGroupName
	}
	return a.groupName
}

// Scale attempts to increase the number of instances by the amount specified
func (a *AWSInventory) Scale(amount int) error {
	// Check inventory status before trying to scale anything
	status, err := a.Status()
	if err != nil {
		return err
	}

	switch status {
	case UPDATING:
		err = errors.New("Won't scale servers while changes are in progress")
	case FAILED:
		err = errors.New("Won't scale servers while something seems to be in a failed state")
	case OK:
		group := a.describeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
		currentCapacity := *group.DesiredCapacity
		a.log.Infof("Current capacity is: %d", currentCapacity)
		newCapacity := currentCapacity + int64(amount)
		a.log.Infof("New desired capacity will be: %d", newCapacity)

		if newCapacity < *group.MinSize {
			err = errors.New("Attempt to scale below minimum capacity denied")
			break
		}
		scalingParams := &autoscaling.SetDesiredCapacityInput{
			AutoScalingGroupName: aws.String(a.GroupName()),
			DesiredCapacity:      aws.Int64(newCapacity),
			HonorCooldown:        aws.Bool(false),
		}
		_, err = a.AutoscalingSvc.SetDesiredCapacity(scalingParams)
	default:
		err = errors.New("Unknown status")
	}
	if err == nil {
		a.log.Infof("Scaling %v by %v", a.GroupName(), amount)
		a.lastModified = time.Now()
	}
	return err
}

// RefreshMetadata pulls updated metadata
func (a *AWSInventory) RefreshMetadata() {
	instanceID, err := a.EC2metadataSvc.GetMetadata("instance-id")
	if err != nil {
		a.log.Fatal(err)
	}
	regionWithAZ, err := a.EC2metadataSvc.GetMetadata("placement/availability-zone")
	if err != nil {
		a.log.Fatal(err)
	}

	// Strip the AZ from the regionWithAZ to get the region
	region := regionWithAZ[:len(regionWithAZ)-1]

	a.metadata = AWSMetadata{
		instanceID:   instanceID,
		regionWithAZ: regionWithAZ,
		region:       region,
	}
}
