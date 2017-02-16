package alice_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/notonthehighstreet/alice"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockAutoScalingClient struct {
	mock.Mock
	autoscalingiface.AutoScalingAPI
}

func (m *MockAutoScalingClient) DescribeAutoScalingGroups(p *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	args := m.Mock.Called()
	output := args.Get(0).(autoscaling.DescribeAutoScalingGroupsOutput)
	return &output, args.Error(1)
}

func (m *MockAutoScalingClient) SetDesiredCapacity(p *autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error) {
	args := m.Mock.Called()
	output := autoscaling.SetDesiredCapacityOutput{}
	return &output, args.Error(0)
}

func (m *MockAutoScalingClient) DescribeScalingActivities(p *autoscaling.DescribeScalingActivitiesInput) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	args := m.Mock.Called()
	output := args.Get(0).(*autoscaling.DescribeScalingActivitiesOutput)
	return output, args.Error(1)
}

type MockEC2MetadataClient struct {
	mock.Mock
}

func (m *MockEC2MetadataClient) GetMetadata(p string) (string, error) {
	args := m.Mock.Called(p)
	return args.Get(0).(string), args.Error(1)
}

var mockEc2MetadataClient MockEC2MetadataClient
var mockAutoscalingClient MockAutoScalingClient
var asg autoscaling.DescribeAutoScalingGroupsOutput
var asgScalingActivities autoscaling.DescribeScalingActivitiesOutput
var AWSInv *alice.AWSInventory

func setupAWSInventoryTest() {
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

	i, _ := alice.NewAWSInventory(viper.New(), log)
	AWSInv = i.(*alice.AWSInventory)
	AWSInv.AutoscalingSvc = &mockAutoscalingClient
	AWSInv.EC2metadataSvc = &mockEc2MetadataClient
}

func TestAWSInventory_Scale(t *testing.T) {
	setupAWSInventoryTest()
	err := AWSInv.Scale(1)
	assert.Nil(t, err)
}

func TestAWSInventory_GroupName(t *testing.T) {
	setupAWSInventoryTest()
	name := AWSInv.GroupName()
	assert.Equal(t, name, "foo")
}

func TestAWSInventory_Total(t *testing.T) {
	setupAWSInventoryTest()
	total, _ := AWSInv.Total()
	assert.Equal(t, total, 1)
}

func TestAWSInventory_Increase(t *testing.T) {
	setupAWSInventoryTest()
	assert.Nil(t, AWSInv.Increase())
	asgScalingActivities.Activities[0].StatusCode = aws.String(autoscaling.ScalingActivityStatusCodeInProgress)
	assert.Error(t, AWSInv.Increase())
}

func TestAWSInventory_Decrease(t *testing.T) {
	setupAWSInventoryTest()
	assert.Nil(t, AWSInv.Decrease())
	asgScalingActivities.Activities[0].StatusCode = aws.String(autoscaling.ScalingActivityStatusCodeInProgress)
	assert.Error(t, AWSInv.Decrease())
}

func TestAWSInventory_Status(t *testing.T) {
	failedActivity := &autoscaling.Activity{
		ActivityId: aws.String("failed-activity"),
		StatusCode: aws.String(autoscaling.ScalingActivityStatusCodeFailed),
	}
	updatingActivity := &autoscaling.Activity{
		ActivityId: aws.String("updating-activity"),
		StatusCode: aws.String(autoscaling.ScalingActivityStatusCodeInProgress),
	}
	setupAWSInventoryTest()
	status, _ := AWSInv.Status()
	assert.Equal(t, alice.OK, status)
	asgScalingActivities.Activities = append(asgScalingActivities.Activities, updatingActivity)
	status, _ = AWSInv.Status()
	assert.Equal(t, alice.UPDATING, status)
	asgScalingActivities.Activities = append(asgScalingActivities.Activities, failedActivity)
	status, _ = AWSInv.Status()
	assert.Equal(t, alice.FAILED, status)
}

func TestAWSInventory_SettleDownTime(t *testing.T) {
	setupAWSInventoryTest()
	AWSInv.Config.Set("settle_down_period", "5m")
	assert.Nil(t, AWSInv.Increase())
	status, _ := AWSInv.Status()
	assert.Equal(t, alice.UPDATING, status)
	assert.Error(t, AWSInv.Decrease())
	assert.Error(t, AWSInv.Increase())
}
