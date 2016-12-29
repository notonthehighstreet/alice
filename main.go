package main

import (
	"time"

	"github.com/notonthehighstreet/autoscaler/manager"
)

func main() {
	m := manager.New()
	for _ = range time.NewTicker(15 * time.Second).C {
		m.Run()
	}
}
