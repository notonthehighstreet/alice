package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/andygrunwald/megos"
)

// getMesosURL returns the URL to the Mesos API when passed a Mesos Master IP/Hostname
func getMesosURL(host string) string {
	var url string
	if len(host) > 0 {
		url = strings.Join([]string{"http://", host, ":5050/"}, "")
	} else {
		fmt.Println("Please set the MESOS_HOST environment variable.")
		os.Exit(1)
	}
	return url
}

func main() {
	// Read the MESOS_HOST environment variable and then build the corect URL
	// probably this should default to the Consul address by default
	host := os.Getenv("MESOS_HOST")
	// Get the Mesos API URL
	mesosURL := getMesosURL(host)
	// Determine who is and connect to the Leader
	mesosNode, _ := url.Parse(mesosURL)
	mesos := megos.NewClient([]*url.URL{mesosNode}, nil)
	mesos.DetermineLeader()
	state, _ := mesos.GetStateFromLeader()

	var totalCPUAvailable, totalMemAvailable, totalCPUUsed, totalMemUsed float64

	for _, slave := range state.Slaves {
		cpuAvailable := slave.UnreservedResources.CPUs - slave.UsedResources.CPUs
		memAvailable := slave.UnreservedResources.Mem - slave.UsedResources.Mem
		totalCPUAvailable += slave.UnreservedResources.CPUs
		totalMemAvailable += slave.UnreservedResources.Mem
		totalCPUUsed += slave.UsedResources.CPUs
		totalMemUsed += slave.UsedResources.Mem

		fmt.Println(cpuAvailable, "CPUs and", memAvailable, "MBs are available on", slave.ID, "(", slave.Hostname, ")")
	}
	fmt.Println("CPUs used:", totalCPUUsed, "of", totalCPUAvailable)
	fmt.Println("Memory used:", totalMemUsed, "of", totalMemAvailable)

}
