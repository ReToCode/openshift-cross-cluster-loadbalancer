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
	go api.RunAPI(":8089", b)

	// Sleep 4 ever
	select {}
}
