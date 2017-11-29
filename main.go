package main

import (
	"os"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer"
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/api"
	"github.com/sirupsen/logrus"
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

const (
	OSE1 = "OpenShift_1"
	OSE2 = "OpenShift_2"
)

func main() {
	// Dummy config to develop
	b := balancer.NewBalancer("localhost:8080", "localhost:8443")
	b.Start()

	// Run web server
	go api.RunAPI("localhost:8089", b)

	// Add dummy clusters
	b.Scheduler.AddCluster(OSE1, []core.Route{{URL: "world.ch", Weight: 1},{URL: "test.ch", Weight: 1},{URL: "only.ch", Weight: 1}})
	b.Scheduler.AddRouterHost(OSE1, "localhost", 8180, 8143)

	b.Scheduler.AddCluster(OSE2, []core.Route{{URL: "world.ch", Weight: 2},{URL: "test.ch", Weight: 1}})
	b.Scheduler.AddRouterHost(OSE2, "localhost", 8280, 8243)
	b.Scheduler.AddRouterHost(OSE2, "localhost", 8380, 8343)

	// Sleep 4 ever
	select {}
}
