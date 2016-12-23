package aws_test

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/aws"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"testing"
)

var log = logging.MustGetLogger("autoscaler")
var mockEc2MetadataClient aws.MockEC2MetadataClient
var mockAutoscalingClient aws.MockAutoScalingClient
var asg autoscaling.DescribeAutoScalingGroupsOutput

func setupTest() {
	instanceId := "i-12345678"
	autoScalingGroupName := "foo"
	desiredCapacity := int64(10)
	minSize := int64(1)
	asg.AutoScalingGroups = []*autoscaling.Group{
		{
			Instances:            []*autoscaling.Instance{{InstanceId: &instanceId}},
			AutoScalingGroupName: &autoScalingGroupName,
			DesiredCapacity:      &desiredCapacity,
			MinSize:              &minSize,
		},
	}
	asg.NextToken = nil
	mockEc2MetadataClient.On("GetMetadata", "instance-id").Return("i-12345678", nil)
	mockEc2MetadataClient.On("GetMetadata", "placement/availability-zone").Return("eu-west-1b", nil)
	mockAutoscalingClient.On("DescribeAutoScalingGroups").Return(asg, nil)
	mockAutoscalingClient.On("SetDesiredCapacity").Return(nil)
}

func TestAWS_Scale(t *testing.T) {
	setupTest()
	s := aws.New(log, &mockAutoscalingClient, &mockEc2MetadataClient)
	err := s.Scale(1)
	assert.Nil(t, err)
}

func TestAWS_GroupName(t *testing.T) {
	setupTest()
	s := aws.New(log, &mockAutoscalingClient, &mockEc2MetadataClient)
	name := s.GroupName()
	assert.Equal(t, name, "foo")

}

func TestAWS_Total(t *testing.T) {
	setupTest()
	s := aws.New(log, &mockAutoscalingClient, &mockEc2MetadataClient)
	total, _ := s.Total()
	assert.Equal(t, total, 1)
}

func TestAWS_Increase(t *testing.T) {
	setupTest()
	s := aws.New(log, &mockAutoscalingClient, &mockEc2MetadataClient)
	err := s.Increase()
	assert.Nil(t, err)
}

func TestAWS_Decrease(t *testing.T) {
	setupTest()
	s := aws.New(log, &mockAutoscalingClient, &mockEc2MetadataClient)
	err := s.Decrease()
	assert.Nil(t, err)
}

func TestAWS_Status(t *testing.T) {
	setupTest()
	s := aws.New(log, &mockAutoscalingClient, &mockEc2MetadataClient)
	state := s.Status()
	assert.Equal(t, state, inventory.OK)
}
