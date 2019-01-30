package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
)

var (
	gitCommit string

	clusterName   = flag.String("cluster-name", "", "Name of magnum cluster to clean nodes in")
	cloudConfig   = flag.String("cloud-config", "", "The path to the cloud provider configuration file.")
	watchInterval = flag.Duration("watch-interval", time.Second*10, "Interval between checking node statuses")
)

func main() {
	flag.Parse()
	glog.Infof("os-autoscaler-nodecleaner built from commit %s", gitCommit)

	if len(*cloudConfig) == 0 {
		glog.Fatalf("Must supply cloud-config location")
	}

	if len(*clusterName) == 0 {
		glog.Fatalf("Must supply cluster-name")
	}

	watcher, err := CreateWatcher()
	if err != nil {
		glog.Fatalf("could not create watcher: %v", err)
	}

	for range time.Tick(*watchInterval) {
		err := watcher.Tick()
		if err != nil {
			glog.Fatalf("watcher error: %v", err)
		}
	}

}
