package main

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/andygrunwald/megos"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("autoscaler")

// getMesosURL returns a valid URL to the Mesos API
func getMesosURL() string {
	hostvar := os.Getenv("MESOS_HOST")
	var url string
	if len(hostvar) > 0 {
		url = strings.Join([]string{"http://", hostvar, ":5050/state"}, "")
	} else {
		// If the environment variable isn't set, then use the Consul address
		url = "http://mesos.service.consul:5050/state"
	}
	log.Info("Mesos URL is:", url)
	return url
}

func mesosStats(mesosURL string) {
	mesosNode, _ := url.Parse(mesosURL)
	mesos := megos.NewClient([]*url.URL{mesosNode}, nil)
	mesos.DetermineLeader()
	state, stateErr := mesos.GetStateFromLeader()
	if stateErr != nil {
		log.Fatal(stateErr)
	}

	var totalCPUAvailable, totalMemAvailable, totalCPUUsed, totalMemUsed float64

	for _, slave := range state.Slaves {
		//cpuAvailable := slave.UnreservedResources.CPUs - slave.UsedResources.CPUs
		//memAvailable := slave.UnreservedResources.Mem - slave.UsedResources.Mem
		totalCPUAvailable += slave.UnreservedResources.CPUs
		totalMemAvailable += slave.UnreservedResources.Mem
		totalCPUUsed += slave.UsedResources.CPUs
		totalMemUsed += slave.UsedResources.Mem
		//fmt.Println(cpuAvailable, "CPUs and", memAvailable, "MBs are available on", slave.ID, "(", slave.Hostname, ")")
	}
	log.Info("")
	log.Info("CPUs used:", totalCPUUsed, "of", totalCPUAvailable)
	log.Info("Memory used:", totalMemUsed, "of", totalMemAvailable)

	cpuPercent := totalCPUUsed / totalCPUAvailable
	memPercent := totalMemUsed / totalMemAvailable
	log.Info("CPU usage:", cpuPercent)
	log.Info("Memory usage:", memPercent)

	if cpuPercent >= 0.8 {
		upByOne()
		log.Info("Scaled up by one due to cpu pressure!")
	} else if memPercent >= 0.8 {
		upByOne()
		log.Info("Scaled up by one due to memory pressure!")
		// quick hack for proof-of-concept. Preferably the scaler will not act if a scaling operation is currently in progress.
		time.Sleep(5 * time.Minute)
		log.Info("Suspending further actions for 5 minutes as the new server joins the cluster")
	}

}

func autoscaler(mesosURL string) {
	for _ = range time.NewTicker(2 * time.Second).C {
		mesosStats(mesosURL)
	}
}

// upByOne scales increases the number of hosts in the autoscaling group by one
func upByOne() {

	// Create aws session
	sess, err := session.NewSession()
	if err != nil {
		log.Error("failed to create session,", err)
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
	log.Info("Current capacity is:", currentCapacity)
	newCapacity := int64(currentCapacity) + 1
	log.Info("New desired capacity will be:", newCapacity)

	// TODO: we'll want to check for the status of the autoscaling group. Won't want to
	//    scale while a scaling operation is already in effect. Probably want to use
	//    a cooldown timer or something
	// Actually, the returned error codes might tell us if an operation is already in progress...

	// TODO: we'll want to run the upscaler next.
	scalingParams := &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(groupName),  // Required
		DesiredCapacity:      aws.Int64(newCapacity), // Required
		HonorCooldown:        aws.Bool(true),
	}
	// A successful response appears to be empty...
	resp, err := svc.SetDesiredCapacity(scalingParams)
	if err != nil {
		log.Error("Failed scaling action:", err)
		return
	}
	log.Info("AWS Response:", resp)
	log.Info("Success! Scaling", groupName, "to", newCapacity, "slaves.")

}

func main() {
	// Get the Mesos URL from an ENV var, or from Consul
	mesosURL := getMesosURL()

	// Run goroutine to grab stats every 5 seconds
	log.Info("Running autoscaler.")
	autoscaler(mesosURL)

}
