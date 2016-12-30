package aws

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/stretchr/testify/mock"
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
