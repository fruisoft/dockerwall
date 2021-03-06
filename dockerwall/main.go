package main

import (
	"flag"
	"os"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
)

const VERSION = "1.0.0-beta"

func main() {
	versionFlag := flag.Bool("version", false, "Print version")
	logLevel := flag.String("loglevel", "debug", "debug, info, warning, error")
	gatewayNetworks := flag.String("gateway-networks", "", "Docker networks whose gateway access will be managed by DockerWall. If empty, all bridge networks will be used")
	flag.Parse()

	switch *logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	if *versionFlag {
		logrus.Infof("%s\n", VERSION)
		return
	}

	logrus.Infof("====Starting Dockerwall %s====", VERSION)

	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		logrus.Errorf("Error creating Docker client instance. err=%s", err)
		return
	}

	gatewayNets := make([]string, 0)
	skipNets := make([]string, 0)
	if *gatewayNetworks != "" {
		gn := strings.Split(*gatewayNetworks, ",")
		for _, v := range gn {
			if len(v) > 1 {
				if v[0] == '!' {
					skipNets = append(skipNets, v[1:])
				} else {
					gatewayNets = append(gatewayNets, v)
				}
			}
		}
	}

	swarmWaller := Waller{
		dockerClient:       cli,
		useDefaultNetworks: (len(gatewayNets) == 0),
		gatewayNetworks:    gatewayNets,
		skipNetworks:       skipNets,
		currentMetrics:     "",
		m:                  &sync.Mutex{},
	}

	swarmWaller.init()
	err = swarmWaller.startup()
	if err != nil {
		logrus.Errorf("Startup error. Exiting. err=%s", err)
		os.Exit(1)
	}

}
