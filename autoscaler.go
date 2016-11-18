package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/andygrunwald/megos"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

var (
	// Info loglevel
	Info *log.Logger
	// Error loglevel
	Error *log.Logger
)

// MesosSlaveStats stores the individual slave stats
type MesosSlaveStats struct {
	// Figure out how exactly to have an array of slave data in here
}

// Init sets logging settings when the app starts
func Init(
	infoHandle io.Writer,
	errorHandle io.Writer) {

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// getMesosURL returns a live, valid URL to the Mesos API
func getMesosURL() string {
	hostvar := os.Getenv("MESOS_HOST")
	var url string
	if len(hostvar) > 0 {
		url = strings.Join([]string{"http://", hostvar, ":5050/"}, "")
	} else {
		Info.Println("MESOS_HOST environment variable not set, so trying default: 'mesos.service.consul'")
		url = "http://mesos.service.consul:5050/"
	}
	response, err := http.Head(url)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		_, err := io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
	Info.Println("Mesos URL is:", url)
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
	fmt.Println("")
	fmt.Println("CPUs used:", totalCPUUsed, "of", totalCPUAvailable)
	fmt.Println("Memory used:", totalMemUsed, "of", totalMemAvailable)
}

func autoscaler(mesosURL string) {
	for _ = range time.NewTicker(5 * time.Second).C {
		fmt.Println("\nI'm an autoscaler, and here's some mesos info:")
		mesos := mesosStats(mesosURL)
		fmt.Println(mesosURL)

		sess, err := session.NewSession()
		if err != nil {
			fmt.Println("failed to create session,", err)
			return
		}

		svc := autoscaling.New(sess, &aws.Config{Region: aws.String("eu-west-1")})

		params := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{
				aws.String("ResourceName"), // Required
				// More values...
			},
			MaxRecords: aws.Int64(1),
			NextToken:  aws.String("XmlString"),
		}
		resp, err := svc.DescribeAutoScalingGroups(params)

		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			return
		}

		// Pretty-print the response data.
		fmt.Println(resp)
	}
}

func webUI(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "200 OK")
}

func main() {
	// Initialize logging
	Init(os.Stdout, os.Stderr)

	// Get the Mesos URL from an ENV var, or from Consul
	mesosURL := getMesosURL()

	// Run goroutine to grab stats every 5 seconds
	//go mesosStats(mesosURL)
	go autoscaler(mesosURL)

	// start web server just because
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	webUI(w, r)
	//})
	//http.ListenAndServe(":8000", nil)

}
