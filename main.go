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
	b.Scheduler.AddRouterHost("localhost", 8180, 8143, []string{"no.ch"})
	b.Scheduler.AddRouterHost("localhost", 8280, 8243, []string{"no.ch"})

	// Https Backend
	b.Scheduler.AddRouterHost("localhost", 8380, 8343, []string{"tls.ch"})

	// Sleep 4 ever
	select {}
}
