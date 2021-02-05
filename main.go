package main

import (
	"context"
	"github.com/oats87/rancher-agent/pkg/applyinator"
	"github.com/oats87/rancher-agent/pkg/config"
	"github.com/oats87/rancher-agent/pkg/localplan"
	"github.com/oats87/rancher-agent/pkg/remoteplan"
	"github.com/rancher/wrangler/pkg/signals"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

var (
	Version   = "v0.0.0-dev"
	GitCommit = "HEAD"
)

func main() {
	logrus.SetOutput(colorable.NewColorableStdout())
	if os.Getenv("CATTLE_DEBUG") == "true" || os.Getenv("RANCHER_DEBUG") == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}
	var err error
	err = run()

	if err != nil {
		logrus.Fatal(err)
	}

}

func run() error {
	topContext := signals.SetupSignalHandler(context.Background())

	logrus.Infof("Rancher System Agent version %s is starting", Version)

	configFile := os.Getenv("CATTLE_AGENT_CONFIG")

	if configFile == "" {
		configFile = "/etc/rancher/agent/config.yaml"
	}

	var cf config.AgentConfig

	err := config.Parse(configFile, &cf)
	if err != nil {
		logrus.Fatalf("Unable to parse config file %v", err)
	}

	logrus.Infof("Using directory %s for work", cf.WorkDir)

	var connInfo config.ConnectionInfo

	err = config.Parse(cf.ConnectionInfoFile, &connInfo)

	if err != nil {
		logrus.Fatalf("Unable to parse connection info file %v", err)
	}

	applyinator := applyinator.NewApplyinator(cf.WorkDir, connInfo.DockerConfig)

	if cf.RemoteEnabled {
		logrus.Infof("Starting remote watch of plans")
		remoteplan.Watch(topContext, *applyinator, connInfo)
	}

	logrus.Infof("Starting local watch of plans in %s", cf.LocalPlanDir)
	localplan.WatchFiles(topContext, *applyinator, cf.LocalPlanDir)

	<-topContext.Done()
	return nil
}
