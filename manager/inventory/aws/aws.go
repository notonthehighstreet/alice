package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"

	"errors"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/sirupsen/logrus"
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

func New(config *viper.Viper, log *logrus.Entry) *AWSInventory {
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
	switch a.Status() {
	case inventory.UPDATING:
		return errors.New("Won't scale servers while changes are in progress")
	case inventory.FAILED:
		return errors.New("Won't scale servers while something seems to be in a failed state")
	case inventory.OK:
		return a.Scale(1)
	}
	return errors.New("Unknown status")
}

func (a *AWSInventory) Decrease() error {
	switch a.Status() {
	case inventory.UPDATING:
		return errors.New("Won't scale servers while changes are in progress")
	case inventory.FAILED:
		return errors.New("Won't scale servers while something seems to be in a failed state")
	case inventory.OK:
		return a.Scale(-1)
	}
	return errors.New("Unknown status")
}

func (a *AWSInventory) Status() inventory.Status {
	// Use state of instances to determine health. Looks like this:
	//"Instances": [
	//    {
	//        "ProtectedFromScaleIn": false,
	//        "AvailabilityZone": "eu-west-1b",
	//        "InstanceId": "i-08599d089fe15af88",
	//        "HealthStatus": "Healthy",
	//        "LifecycleState": "Pending",
	//        "LaunchConfigurationName": "qa-full-lights-mesos-slave-LaunchConfig-QV7DAOP4GS35"
	//    },
	//    {
	//        "ProtectedFromScaleIn": false,
	//        "AvailabilityZone": "eu-west-1a",
	//        "InstanceId": "i-0dacc77123aca3074",
	//        "HealthStatus": "Healthy",
	//        "LifecycleState": "Pending",
	//        "LaunchConfigurationName": "qa-full-lights-mesos-slave-LaunchConfig-QV7DAOP4GS35"
	//    },
	//    {
	//        "ProtectedFromScaleIn": false,
	//        "AvailabilityZone": "eu-west-1c",
	//        "InstanceId": "i-0dda606d86e6fd163",
	//        "HealthStatus": "Healthy",
	//        "LifecycleState": "InService",
	//        "LaunchConfigurationName": "qa-full-lights-mesos-slave-LaunchConfig-QV7DAOP4GS35"
	//    }
	//],
	// Also there's "ScalingActivityInProgress: Scaling activity 15449fcf-acdb-4ca0-abc4-1ab111203267 is in progress and blocks this action"
	// describe-scaling-activities for that
	return inventory.OK
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
	group := a.describeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	// myGroup contains the autoscaling group we live in.
	groupName := *group.AutoScalingGroupName
	currentCapacity := *group.DesiredCapacity
	a.log.Infof("Current capacity is: %d", currentCapacity)
	newCapacity := currentCapacity + int64(amount)
	a.log.Infof("New desired capacity will be: %d", newCapacity)

	if newCapacity < *group.MinSize {
		return errors.New("aws: attempt to scale below minimum capacity denied")
	}

	// This will fail if there's an operation already in place.
	scalingParams := &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(groupName),
		DesiredCapacity:      aws.Int64(newCapacity),
		HonorCooldown:        aws.Bool(true),
	}

	// A successful response is one that doesn't return an error.
	if _, err := a.AutoscalingSvc.SetDesiredCapacity(scalingParams); err != nil {
		return err
	}

	a.log.Infof("Scaling %s to %f slaves", groupName, newCapacity)
	return nil
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
