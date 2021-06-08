// +build linux

package main

import (
	"github.com/sirupsen/logrus"
	"github.com/tktip/cfger"
	"github.com/tktip/flyvo-api/internal/api"
	"github.com/tktip/flyvo-api/internal/version"
	"github.com/tktip/flyvo-api/pkg/healthcheck"
)

func main() {
	// it is good practice to log version on startup
	logrus.Infof("Running version: %s", version.VERSION)

	apiSrv := api.Server{}
	_, err := cfger.ReadStructuredCfgRecursive("env::CONFIG", &apiSrv)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	if apiSrv.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	//Starting health check
	go healthcheck.StartHealthService()

	logrus.Fatalf("Api server failed: %s", apiSrv.Run())
}
