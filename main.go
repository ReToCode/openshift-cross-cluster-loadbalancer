package main

import (
	"os"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer"
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/api"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
	// Dummy config to develop
	b := balancer.NewBalancer("localhost:8080", "localhost:8443")
	b.Start()

	// Run webserver
	go api.RunAPI("localhost:8089", b)

	// Add local host as router host
	// Http backends
	b.Scheduler.AddRouterHost("localhost:8001", []string{"localhost:8001", "no.ch"})
	b.Scheduler.AddRouterHost("localhost:8002", []string{"localhost:8002", "no.ch"})

	// Https Backend
	b.Scheduler.AddRouterHost("localhost:8003", []string{"localhost:8003", "tls.ch"})

	// Sleep 4 ever
	select {}
}
