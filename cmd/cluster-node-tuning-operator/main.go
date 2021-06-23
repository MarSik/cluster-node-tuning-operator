package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"

	"k8s.io/klog/v2"

	tunedv1 "github.com/openshift/cluster-node-tuning-operator/pkg/apis/tuned/v1"
	"github.com/openshift/cluster-node-tuning-operator/pkg/metrics"
	"github.com/openshift/cluster-node-tuning-operator/pkg/operator"
	"github.com/openshift/cluster-node-tuning-operator/pkg/signals"
	"github.com/openshift/cluster-node-tuning-operator/pkg/tuned"
	"github.com/openshift/cluster-node-tuning-operator/version"
)

var (
	boolVersion = flag.Bool("version", false, "show program version and exit")
	boolLocal   = flag.Bool("local", false, "local run outside a pod")
)

func printVersion() {
	klog.Infof("Go Version: %s", runtime.Version())
	klog.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	klog.Infof("%s Version: %s", tunedv1.TunedClusterOperatorResourceName, version.Version)
}

func main() {
	const (
		operandFilename  string = "openshift-tuned"
		operatorFilename string = "cluster-node-tuning-operator"
	)

	runAs := filepath.Base(os.Args[0])

	stopCh := signals.SetupSignalHandler()

	switch runAs {
	case operatorFilename:
		klog.InitFlags(nil)
		flag.Parse()

		printVersion()

		if *boolVersion {
			os.Exit(0)
		}

		go metrics.RunServer(metrics.MetricsPort, stopCh)
		metrics.RegisterVersion(version.Version)
		controller, err := operator.NewController()
		if err != nil {
			klog.Fatal(err)
		}

		controller.BecomeLeader(stopCh)
	case operandFilename:
		tuned.Run(stopCh, boolVersion, version.Version)
	default:
		klog.Fatalf("application should be run as \"%s\" or \"%s\"", operatorFilename, operandFilename)
	}
}
