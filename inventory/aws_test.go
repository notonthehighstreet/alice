package inventory_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/notonthehighstreet/autoscaler/inventory"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var mockEc2MetadataClient inventory.MockEC2MetadataClient
var mockAutoscalingClient inventory.MockAutoScalingClient
var asg autoscaling.DescribeAutoScalingGroupsOutput
var asgScalingActivities autoscaling.DescribeScalingActivitiesOutput
var AWSInv *inventory.AWS

func setupAWSTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager":   "Mock",
		"inventory": "AWSInventory",
	})
	asg.AutoScalingGroups = []*autoscaling.Group{
		{
			Instances:            []*autoscaling.Instance{{InstanceId: aws.String("i-12345678")}},
			AutoScalingGroupName: aws.String("foo"),
			DesiredCapacity:      aws.Int64(10),
			MinSize:              aws.Int64(1),
		},
	}
	asg.NextToken = nil
	asgScalingActivities.Activities = []*autoscaling.Activity{
		{
			ActivityId: aws.String("activity-id-1"),
			StatusCode: aws.String(autoscaling.ScalingActivityStatusCodeSuccessful),
		},
		{
			ActivityId: aws.String("activity-id-2"),
			StatusCode: aws.String(autoscaling.ScalingActivityStatusCodeSuccessful),
		},
		{
			ActivityId: aws.String("activity-id-3"),
			StatusCode: aws.String(autoscaling.ScalingActivityStatusCodeSuccessful),
		},
	}
	asgScalingActivities.NextToken = nil
	mockEc2MetadataClient.On("GetMetadata", "instance-id").Return("i-12345678", nil)
	mockEc2MetadataClient.On("GetMetadata", "placement/availability-zone").Return("eu-west-1b", nil)
	mockAutoscalingClient.On("DescribeAutoScalingGroups").Return(asg, nil)
	mockAutoscalingClient.On("DescribeScalingActivities").Return(&asgScalingActivities, nil)
	mockAutoscalingClient.On("SetDesiredCapacity").Return(nil)
	log.Logger.Level = logrus.DebugLevel

	i, _ := inventory.NewAWS(viper.New(), log)
	AWSInv = i.(*inventory.AWS)
	AWSInv.AutoscalingSvc = &mockAutoscalingClient
	AWSInv.EC2metadataSvc = &mockEc2MetadataClient
}

func TestAWS_Scale(t *testing.T) {
	setupAWSTest()
	err := AWSInv.Scale(1)
	assert.Nil(t, err)
}

func TestAWS_GroupName(t *testing.T) {
	setupAWSTest()
	name := AWSInv.GroupName()
	assert.Equal(t, name, "foo")
}

func TestAWS_Total(t *testing.T) {
	setupAWSTest()
	total, _ := AWSInv.Total()
	assert.Equal(t, total, 1)
}

func TestAWS_Increase(t *testing.T) {
	setupAWSTest()
	assert.Nil(t, AWSInv.Increase())
	asgScalingActivities.Activities[0].StatusCode = aws.String(autoscaling.ScalingActivityStatusCodeInProgress)
	assert.Error(t, AWSInv.Increase())
}

func TestAWS_Decrease(t *testing.T) {
	setupAWSTest()
	assert.Nil(t, AWSInv.Decrease())
	asgScalingActivities.Activities[0].StatusCode = aws.String(autoscaling.ScalingActivityStatusCodeInProgress)
	assert.Error(t, AWSInv.Decrease())
}

func TestAWS_Status(t *testing.T) {
	failedActivity := &autoscaling.Activity{
		ActivityId: aws.String("failed-activity"),
		StatusCode: aws.String(autoscaling.ScalingActivityStatusCodeFailed),
	}
	updatingActivity := &autoscaling.Activity{
		ActivityId: aws.String("updating-activity"),
		StatusCode: aws.String(autoscaling.ScalingActivityStatusCodeInProgress),
	}
	setupAWSTest()
	status, _ := AWSInv.Status()
	assert.Equal(t, inventory.OK, status)
	asgScalingActivities.Activities = append(asgScalingActivities.Activities, updatingActivity)
	status, _ = AWSInv.Status()
	assert.Equal(t, inventory.UPDATING, status)
	asgScalingActivities.Activities = append(asgScalingActivities.Activities, failedActivity)
	status, _ = AWSInv.Status()
	assert.Equal(t, inventory.FAILED, status)
}

func TestAWS_SettleDownTime(t *testing.T) {
	setupAWSTest()
	AWSInv.Config.Set("settle_down_period", "5m")
	assert.Nil(t, AWSInv.Increase())
	status, _ := AWSInv.Status()
	assert.Equal(t, inventory.UPDATING, status)
	assert.Error(t, AWSInv.Decrease())
	assert.Error(t, AWSInv.Increase())
}
