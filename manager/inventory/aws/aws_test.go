package aws_test

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/aws"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var log = logrus.WithFields(logrus.Fields{
	"manager":   "Mock",
	"inventory": "AWSInventory",
})
var mockEc2MetadataClient aws.MockEC2MetadataClient
var mockAutoscalingClient aws.MockAutoScalingClient
var asg autoscaling.DescribeAutoScalingGroupsOutput
var inv *aws.AWSInventory

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
	inv = aws.New(viper.New(), log).(*aws.AWSInventory)
	inv.AutoscalingSvc = &mockAutoscalingClient
	inv.EC2metadataSvc = &mockEc2MetadataClient
}

func TestAWS_Scale(t *testing.T) {
	setupTest()
	err := inv.Scale(1)
	assert.Nil(t, err)
}

func TestAWS_GroupName(t *testing.T) {
	setupTest()
	name := inv.GroupName()
	assert.Equal(t, name, "foo")

}

func TestAWS_Total(t *testing.T) {
	setupTest()
	total, _ := inv.Total()
	assert.Equal(t, total, 1)
}

func TestAWS_Increase(t *testing.T) {
	setupTest()
	err := inv.Increase()
	assert.Nil(t, err)
}

func TestAWS_Decrease(t *testing.T) {
	setupTest()
	err := inv.Decrease()
	assert.Nil(t, err)
}

func TestAWS_Status(t *testing.T) {
	setupTest()
	state := inv.Status()
	assert.Equal(t, state, inventory.OK)
}
