package scaler

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"

	"github.com/op/go-logging"
)

type Scaler struct {
	session *session.Session
	logger  *logging.Logger
}

type ScalerMetadata struct {
	region       string
	regionWithAZ string
	instanceID   string
}

// NewScaler initialises a new Scaler instance with an AWS session.
func NewScaler(logger *logging.Logger) (*Scaler, error) {
	session, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return &Scaler{session: session, logger: logger}, nil
}

func (s *Scaler) ScaleUp() error {
	metadata, err := s.metadata()
	if err != nil {
		return err
	}

	// create our session to AWS
	svc := autoscaling.New(s.session, aws.NewConfig().WithRegion(metadata.region))

	// Grab our Autoscaling Group
	// does this need to be an array? I *think* there will only ever be one
	var myGroup []*autoscaling.Group
	done := false
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{},
	}
	for !done {
		resp, errD := svc.DescribeAutoScalingGroups(params)
		if errD != nil {
			return err
		}

		for _, group := range resp.AutoScalingGroups {
			for _, server := range group.Instances {
				if *server.InstanceId == metadata.instanceID {
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
	s.logger.Infof("scaler: current capacity is: %d", currentCapacity)
	newCapacity := currentCapacity + 1
	s.logger.Infof("scaler: new desired capacity will be: %d", newCapacity)

	// TODO: we'll want to check for the status of the autoscaling group. Won't want to
	//    scale while a scaling operation is already in effect. Probably want to use
	//    a cooldown timer or something
	// Actually, the returned error codes might tell us if an operation is already in progress...

	scalingParams := &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(groupName),
		DesiredCapacity:      aws.Int64(newCapacity),
		HonorCooldown:        aws.Bool(true),
	}

	// A successful response is one that doesn't return an error.
	if _, err = svc.SetDesiredCapacity(scalingParams); err != nil {
		return err
	}

	s.logger.Infof("scaler: scaling %s to %d slaves", groupName, newCapacity)
	return nil
}

func (s *Scaler) metadata() (*ScalerMetadata, error) {
	svc := ec2metadata.New(s.session)

	instanceID, err := svc.GetMetadata("instance-id")
	if err != nil {
		return nil, err
	}
	regionWithAZ, err := svc.GetMetadata("placement/availability-zone")
	if err != nil {
		return nil, err
	}

	// Strip the AZ from the regionWithAZ to get the region
	region := regionWithAZ[:len(regionWithAZ)-1]

	return &ScalerMetadata{
		instanceID:   instanceID,
		regionWithAZ: regionWithAZ,
		region:       region,
	}, nil
}
