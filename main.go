package main

import (
	"os"

	"os/signal"
	"syscall"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer"
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/api"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

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

	// Dummy config to develop
	b := balancer.NewBalancer(":8080", ":8443")
	b.Start()

	// Run web server
	go api.RunAPI(":8089", b)

	// Sleep 4 ever
	select {}
}
