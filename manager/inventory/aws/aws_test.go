package aws_test

import (
	"github.com/Sirupsen/logrus"
	amz "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/aws"
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
var asgScalingActivities autoscaling.DescribeScalingActivitiesOutput
var inv *aws.AWSInventory

func setupTest() {
	asg.AutoScalingGroups = []*autoscaling.Group{
		{
			Instances:            []*autoscaling.Instance{{InstanceId: amz.String("i-12345678")}},
			AutoScalingGroupName: amz.String("foo"),
			DesiredCapacity:      amz.Int64(10),
			MinSize:              amz.Int64(1),
		},
	}
	asg.NextToken = nil
	asgScalingActivities.Activities = []*autoscaling.Activity{
		{
			ActivityId: amz.String("activity-id-1"),
			StatusCode: amz.String(autoscaling.ScalingActivityStatusCodeSuccessful),
		},
		{
			ActivityId: amz.String("activity-id-2"),
			StatusCode: amz.String(autoscaling.ScalingActivityStatusCodeSuccessful),
		},
		{
			ActivityId: amz.String("activity-id-3"),
			StatusCode: amz.String(autoscaling.ScalingActivityStatusCodeSuccessful),
		},
	}
	asgScalingActivities.NextToken = nil
	mockEc2MetadataClient.On("GetMetadata", "instance-id").Return("i-12345678", nil)
	mockEc2MetadataClient.On("GetMetadata", "placement/availability-zone").Return("eu-west-1b", nil)
	mockAutoscalingClient.On("DescribeAutoScalingGroups").Return(asg, nil)
	mockAutoscalingClient.On("DescribeScalingActivities").Return(&asgScalingActivities, nil)
	mockAutoscalingClient.On("SetDesiredCapacity").Return(nil)
	log.Logger.Level = logrus.DebugLevel

	i, _ := aws.New(viper.New(), log)
	inv = i.(*aws.AWSInventory)
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
	assert.Nil(t, inv.Increase())
	asgScalingActivities.Activities[0].StatusCode = amz.String(autoscaling.ScalingActivityStatusCodeInProgress)
	assert.Error(t, inv.Increase())
}

func TestAWS_Decrease(t *testing.T) {
	setupTest()
	assert.Nil(t, inv.Decrease())
	asgScalingActivities.Activities[0].StatusCode = amz.String(autoscaling.ScalingActivityStatusCodeInProgress)
	assert.Error(t, inv.Decrease())
}

func TestAWS_Status(t *testing.T) {
	failedActivity := &autoscaling.Activity{
		ActivityId: amz.String("failed-activity"),
		StatusCode: amz.String(autoscaling.ScalingActivityStatusCodeFailed),
	}
	updatingActivity := &autoscaling.Activity{
		ActivityId: amz.String("updating-activity"),
		StatusCode: amz.String(autoscaling.ScalingActivityStatusCodeInProgress),
	}
	setupTest()
	status, _ := inv.Status()
	assert.Equal(t, inventory.OK, status)
	asgScalingActivities.Activities = append(asgScalingActivities.Activities, updatingActivity)
	status, _ = inv.Status()
	assert.Equal(t, inventory.UPDATING, status)
	asgScalingActivities.Activities = append(asgScalingActivities.Activities, failedActivity)
	status, _ = inv.Status()
	assert.Equal(t, inventory.FAILED, status)
}

func TestSettleDownTime(t *testing.T) {
	setupTest()
	inv.Config.Set("settle_down_period", "5m")
	assert.Nil(t, inv.Increase())
	status, _ := inv.Status()
	assert.Equal(t, inventory.UPDATING, status)
	assert.Error(t, inv.Decrease())
	assert.Error(t, inv.Increase())
}
