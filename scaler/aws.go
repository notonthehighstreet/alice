package scaler

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

func UpByOne() {

	// Create aws session
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}

	// Retrieve instance ID from metadata
	ec2svc := ec2metadata.New(sess)
	instanceID, _ := ec2svc.GetMetadata("instance-id")
	regionWithAZ, _ := ec2svc.GetMetadata("placement/availability-zone")
	var region string
	//strip the AZ from the regionWithAZ to get the region
	region = regionWithAZ[:len(regionWithAZ)-1]

	// create our session to AWS
	svc := autoscaling.New(sess, aws.NewConfig().WithRegion(region))

	// Grab our Autoscaling Group
	// does this need to be an array? I *think* there will only ever be one
	var myGroup []*autoscaling.Group
	done := false
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{},
	}
	for !done {
		resp, _ := svc.DescribeAutoScalingGroups(params)

		for _, group := range resp.AutoScalingGroups {
			for _, server := range group.Instances {
				if *server.InstanceId == instanceID {
					myGroup = append(myGroup, group)
				}
			}
		}
		if resp.NextToken == nil {
			done = true
		} else {
			params.NextToken = resp.NextToken
		}
	}
	// myGroup contains the autoscaling group we live in.
	groupName := *myGroup[0].AutoScalingGroupName
	currentCapacity := *myGroup[0].DesiredCapacity
	fmt.Println("Current capacity is:", currentCapacity)
	newCapacity := int64(currentCapacity) + 1
	fmt.Println("New desired capacity will be:", newCapacity)

	// TODO: we'll want to check for the status of the autoscaling group. Won't want to
	//    scale while a scaling operation is already in effect. Probably want to use
	//    a cooldown timer or something
	// Actually, the returned error codes should tell us if an operation is already in progress...

	// TODO: we'll want to run the upscaler next.
	scalingParams := &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(groupName),  // Required
		DesiredCapacity:      aws.Int64(newCapacity), // Required
		HonorCooldown:        aws.Bool(true),
	}
	resp, err := svc.SetDesiredCapacity(scalingParams)
	if err != nil {
		fmt.Println("Failed:", err)
		return
	}
	fmt.Println(resp)
	fmt.Println("Success! Scaling", groupName, "to", newCapacity, "slaves.")

}
