package mesos

import (
	"net/url"
	"github.com/andygrunwald/megos"
)

type MesosClient interface {
	GetStateFromLeader() (*megos.State, error)
}

func NewMesosClient(u string) MesosClient {
	mesosNode, _ := url.Parse(u)
	mesos := megos.NewClient([]*url.URL{mesosNode}, nil)
	mesos.DetermineLeader()
	return mesos
}
