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
	cluster := balancer.NewCluster("OpenShift Local")
	cluster.Start()

	// Add local host as router host
	cluster.Scheduler.AddRouterHost("localhost:8080")
	cluster.Scheduler.AddRouterHost("localhost:8081")

	// Sleep 4 ever
	select {}
}
