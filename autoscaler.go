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
)

var (
	// Info loglevel
	Info *log.Logger
	// Error loglevel
	Error *log.Logger
)

// MesosClusterStats for storing cluster info
type MesosClusterStats struct {
	clusterCPUAvailable, clusterCPUUsed float64
	clusterMemAvailable, clusterMemUsed float64
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

func mesosStats(mesosURL string, clusterStats MesosClusterStats) {
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
	clusterStats.clusterCPUAvailable = totalCPUAvailable
	clusterStats.clusterCPUUsed = totalCPUUsed
	clusterStats.clusterMemAvailable = totalMemAvailable
	clusterStats.clusterMemUsed = totalMemUsed
	fmt.Println("CPUs used:", totalCPUUsed, "of", totalCPUAvailable)
	fmt.Println("Memory used:", totalMemUsed, "of", totalMemAvailable)
}

func autoscaler() {
	fmt.Println("Pretend I'm an autoscaler running sometimes...")
}

func webUI(w http.ResponseWriter, r *http.Request, clusterStats MesosClusterStats) {
	io.WriteString(w, "Hello world!")
	fmt.Println(clusterStats)
}

func webHealth(w http.ResponseWriter, r *http.Request) {
	// for now, if the process is up, we'll consider it healthy
	io.WriteString(w, "200 OK")
	// look into using Runner's task.Running() as a health check.
}

func main() {
	// Initialize logging
	Init(os.Stdout, os.Stderr)

	// Global instance of our MesosClusterStats so multiple functions can access the info
	var clusterStats MesosClusterStats

	// Get the Mesos URL from an ENV var, or from Consul
	mesosURL := getMesosURL()

	// Run goroutine to grab stats every 5 seconds
	for _ = range time.NewTicker(5 * time.Second).C {
		mesosStats(mesosURL, clusterStats)
	}

	// run autoscaler function
	for _ = range time.NewTicker(5 * time.Second).C {
		autoscaler()
	}

	// start web server just because
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		webUI(w, r, clusterStats)
	})
	http.ListenAndServe(":8000", nil)

}
