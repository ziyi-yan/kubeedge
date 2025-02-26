package cloudhub

import (
	"io/ioutil"
	"os"

	"github.com/kubeedge/beehive/pkg/common/config"
	"github.com/kubeedge/beehive/pkg/common/log"
	"github.com/kubeedge/beehive/pkg/core"
	"github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/kubeedge/cloud/pkg/cloudhub/channelq"
	"github.com/kubeedge/kubeedge/cloud/pkg/cloudhub/common/util"
	chconfig "github.com/kubeedge/kubeedge/cloud/pkg/cloudhub/config"
	"github.com/kubeedge/kubeedge/cloud/pkg/cloudhub/servers"
)

type cloudHub struct {
	context  *context.Context
	stopChan chan bool
}

func init() {
	core.Register(&cloudHub{})
}

func (a *cloudHub) Name() string {
	return "cloudhub"
}

func (a *cloudHub) Group() string {
	return "cloudhub"
}

func (a *cloudHub) Start(c *context.Context) {
	a.context = c
	a.stopChan = make(chan bool)

	initHubConfig()

	eventq := channelq.NewChannelEventQueue(c)

	// start dispatch message from the cloud to edge node
	go eventq.DispatchMessage()

	// start the cloudhub server
	if util.HubConfig.ProtocolWebsocket {
		go servers.StartCloudHub(servers.ProtocolWebsocket, eventq, c)
	}

	if util.HubConfig.ProtocolQuic {
		go servers.StartCloudHub(servers.ProtocolQuic, eventq, c)
	}

	<-a.stopChan
}

func (a *cloudHub) Cleanup() {
	a.stopChan <- true
	a.context.Cleanup(a.Name())
}

func initHubConfig() {
	cafile, err := config.CONFIG.GetValue("cloudhub.ca").ToString()
	if err != nil {
		log.LOGGER.Info("missing cloudhub.ca configuration key, loading default path and filename ./" + chconfig.DefaultCAFile)
		cafile = chconfig.DefaultCAFile
	}

	certfile, err := config.CONFIG.GetValue("cloudhub.cert").ToString()
	if err != nil {
		log.LOGGER.Info("missing cloudhub.cert configuration key, loading default path and filename ./" + chconfig.DefaultCertFile)
		certfile = chconfig.DefaultCertFile
	}

	keyfile, err := config.CONFIG.GetValue("cloudhub.key").ToString()
	if err != nil {
		log.LOGGER.Info("missing cloudhub.key configuration key, loading default path and filename ./" + chconfig.DefaultKeyFile)
		keyfile = chconfig.DefaultKeyFile
	}

	errs := make([]string, 0)

	util.HubConfig.Ca, err = ioutil.ReadFile(cafile)
	if err != nil {
		errs = append(errs, err.Error())
	}
	util.HubConfig.Cert, err = ioutil.ReadFile(certfile)
	if err != nil {
		errs = append(errs, err.Error())
	}
	util.HubConfig.Key, err = ioutil.ReadFile(keyfile)
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		log.LOGGER.Errorf("cloudhub failed with errors : %v", errs)
		os.Exit(1)
	}
}
