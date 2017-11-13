package main

import (
	"os"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	// Dummy config to develop
	b := balancer.NewBalancer("localhost:9999")
	b.Start()

	// Add local host as router host
	b.Scheduler.AddRouterHost("localhost:8080", []string{"localhost:8080"})
	b.Scheduler.AddRouterHost("localhost:8081", []string{"localhost:8081", "localhost:9999"})
	b.Scheduler.AddRouterHost("localhost:8082", []string{"localhost:8082", "3.ch"})

	// Sleep 4 ever
	select {}
}
