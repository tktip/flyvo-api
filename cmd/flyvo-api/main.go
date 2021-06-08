// +build windows

package main

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tktip/cfger"
	"github.com/tktip/flyvo-api/internal/api"
	"github.com/tktip/flyvo-api/internal/version"
	"github.com/tktip/flyvo-api/pkg/healthcheck"
)

func main() {
	cfgFile := ""
	if len(os.Args) > 1 {
		for _, v := range os.Args {
			vals := strings.Split(v, "=")
			if len(vals) == 1 {
				continue
			}
			if vals[0] == "configFile" {
				cfgFile = vals[1]
			}
		}
	}
	if cfgFile == "" {
		logrus.Fatal("Missing parameter 'configFile'")
	}

	// it is good practice to log version on startup
	logrus.Infof("Running version: %s", version.VERSION)

	//Starting health check
	go healthcheck.StartHealthService()

	apiSrv := api.Server{}
	_, err := cfger.ReadStructuredCfg("file::"+cfgFile, &apiSrv)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	logrus.Errorf("Api server failed: %s", apiSrv.Run())
	if apiSrv.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	apiSrv.Run()
}
