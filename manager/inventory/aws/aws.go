package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"

	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/spf13/viper"
)

type EC2MetadataAPI interface {
	GetMetadata(string) (string, error)
}

type AWSInventory struct {
	log            *logrus.Entry
	config         *viper.Viper
	AutoscalingSvc autoscalingiface.AutoScalingAPI
	EC2metadataSvc EC2MetadataAPI
	groupName      string
	metadata       AWSMetadata
}

type AWSMetadata struct {
	region       string
	regionWithAZ string
	instanceID   string
}

func New(config *viper.Viper, log *logrus.Entry) inventory.Inventory {
	config.SetDefault("region", "eu-west-1")
	s, err := session.NewSession()
	if err != nil {
		log.Errorf("%s", err.Error())
	}
	region := config.GetString("region")
	s.Config.Region = &region
	a := AWSInventory{AutoscalingSvc: autoscaling.New(s), EC2metadataSvc: ec2metadata.New(s), log: log, config: config}
	return &a
}

func (a *AWSInventory) Total() (int, error) {
	name := a.GroupName()
	params := &autoscaling.DescribeAutoScalingGroupsInput{AutoScalingGroupNames: []*string{&name}}
	group := a.describeAutoScalingGroups(params)
	return len(group.Instances), nil

}

func (a *AWSInventory) Increase() error {
	return a.Scale(+1)
}

func (a *AWSInventory) Decrease() error {
	return a.Scale(-1)
}

func (a *AWSInventory) Status() inventory.Status {
	params := &autoscaling.DescribeScalingActivitiesInput{AutoScalingGroupName: aws.String(a.GroupName())}
	status := inventory.OK
	done := false
	for !done {
		resp, err := a.AutoscalingSvc.DescribeScalingActivities(params)
		if err != nil {
			a.log.Fatalf("%v", err)
		}
		a.log.Debugf("Checking %v pre-existing scaling activites", len(resp.Activities))
		for _, activity := range resp.Activities {
			switch *activity.StatusCode {
			case autoscaling.ScalingActivityStatusCodeSuccessful:
				a.log.Debugln("Ignoring a successful activity")
				continue
			case autoscaling.ScalingActivityStatusCodeCancelled:
				a.log.Debugln("Ignoring a cancelled activity")
				continue
			case autoscaling.ScalingActivityStatusCodeFailed:
				a.log.Debugln("Found a failed activity")
				return inventory.FAILED
			default:
				a.log.Debugln("Found an in-progress activity")
				status = inventory.UPDATING
			}
		}
		if resp.NextToken == nil {
			done = true
		} else {
			params.NextToken = resp.NextToken
		}
	}
	return status
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

func (a *AWSInventory) GroupName() string {
	if a.groupName == "" {
		group := a.describeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
		a.groupName = *group.AutoScalingGroupName
	}
	return a.groupName
}

func (a *AWSInventory) Scale(amount int) error {
	// Check inventory status before trying to scale anything
	var e error
	switch a.Status() {
	case inventory.UPDATING:
		e = errors.New("Won't scale servers while changes are in progress")
	case inventory.FAILED:
		e = errors.New("Won't scale servers while something seems to be in a failed state")
	case inventory.OK:
		group := a.describeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
		currentCapacity := *group.DesiredCapacity
		a.log.Infof("Current capacity is: %d", currentCapacity)
		newCapacity := currentCapacity + int64(amount)
		a.log.Infof("New desired capacity will be: %d", newCapacity)

		if newCapacity < *group.MinSize {
			e = errors.New("Attempt to scale below minimum capacity denied")
			break
		}
		scalingParams := &autoscaling.SetDesiredCapacityInput{
			AutoScalingGroupName: aws.String(a.GroupName()),
			DesiredCapacity:      aws.Int64(newCapacity),
			HonorCooldown:        aws.Bool(false),
		}
		_, e = a.AutoscalingSvc.SetDesiredCapacity(scalingParams)
	default:
		e = errors.New("Unknown status")
	}
	if e == nil {
		a.log.Infof("Scaling %v by %v", a.GroupName(), amount)
	} else {
		a.log.Errorln(e.Error())
	}
	return e
}

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
