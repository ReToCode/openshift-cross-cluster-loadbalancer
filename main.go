package main

import (
	"os"

	"os/signal"
	"runtime"
	"syscall"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer"
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/api"
	"github.com/sirupsen/logrus"

	"net/http"
	_ "net/http/pprof"
)

func init() {
	logrus.SetOutput(os.Stdout)
	//logrus.SetLevel(logrus.DebugLevel)
	logrus.SetLevel(logrus.InfoLevel)
}

const (
	OSE1 = "OpenShift_1"
	OSE2 = "OpenShift_2"
)

func main() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // Termination Request
			syscall.SIGSEGV, // FullDerp
			syscall.SIGABRT, // Abnormal termination
			syscall.SIGILL,  // illegal instruction
			syscall.SIGFPE)  // floating point
		sig := <-c
		logrus.Fatalf("Signal (%v) Detected, Shutting Down", sig)
	}()

	// Profiling
	runtime.SetMutexProfileFraction(5)

	go func() {
		logrus.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Dummy config to develop
	b := balancer.NewBalancer("localhost:8080", "localhost:8543")
	b.Start()

	// Run web server
	go api.RunAPI("localhost:8089", b)

	// Add dummy clusters
	//b.Scheduler.AddCluster(OSE1, []core.Route{{URL: "world.ch", Weight: 1}, {URL: "test.ch", Weight: 1}, {URL: "only.ch", Weight: 1}})
	//b.Scheduler.AddRouterHost(OSE1, "localhost", 8180, 8143)
	//
	//b.Scheduler.AddCluster(OSE2, []core.Route{{URL: "world.ch", Weight: 2}, {URL: "test.ch", Weight: 1}})
	//b.Scheduler.AddRouterHost(OSE2, "localhost", 8280, 8243)
	//b.Scheduler.AddRouterHost(OSE2, "localhost", 8380, 8343)

	// Sleep 4 ever
	select {}
}
