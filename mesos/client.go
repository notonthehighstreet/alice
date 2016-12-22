package mesos

import (
	"github.com/andygrunwald/megos"
	"net/url"
)

type MesosClient interface {
	GetStateFromLeader() (*megos.State, error)
}

func NewMesosClient(URL string) (MesosClient, error) {
	mesosNode, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	mesos := megos.NewClient([]*url.URL{mesosNode}, nil)
	mesos.DetermineLeader()
	return mesos, nil
}
