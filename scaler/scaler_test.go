package scaler_test

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/notonthehighstreet/autoscaler/scaler"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"testing"
)

var log = logging.MustGetLogger("autoscaler")
var mockEc2MetadataClient scaler.MockEC2MetadataClient
var mockAutoscalingClient scaler.MockAutoScalingClient
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
	mockAutoscalingClient.On("DescribeAutoScalingGroups", &autoscaling.DescribeAutoScalingGroupsInput{}).Return(asg, nil)
	mockAutoscalingClient.On("SetDesiredCapacity").Return(nil)
}
func TestNewScaler(t *testing.T) {
	_, err := scaler.NewScaler(log, &mockAutoscalingClient, &mockEc2MetadataClient)
	assert.Equal(t, nil, err)
}

func TestScaler_Scale(t *testing.T) {
	setupTest()
	s, _ := scaler.NewScaler(log, &mockAutoscalingClient, &mockEc2MetadataClient)
	err := s.Scale(1)
	assert.Equal(t, nil, err)
}
