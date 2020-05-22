package ping

import (
	"fmt"
	"time"

	"github.com/riser-platform/riser-server/pkg/sdk"

	"github.com/go-logr/logr"
)

type pinger struct {
	riserClient     *sdk.Client
	log             logr.Logger
	environmentName string
	ticker          *time.Ticker
}

/*
StartNewPinger creates and starts a new pinger. The server maintains the "last ping" from a controller to track when the last communication was received
from the controller. A ping happens automatically when a status update is received, however, since there can be periods with few status updates to
the server, a "pinger" is needed to inform the server of connectivity.
*/
func StartNewPinger(riserClient *sdk.Client, log logr.Logger, environmentName string, pingFrequency time.Duration) {
	ping := &pinger{riserClient, log, environmentName, time.NewTicker(pingFrequency)}
	// We block on the first ping since this bootstraps a new environment. We probably want to remove this in favor of environment config endpoints
	// on the server creating the environment if it doesn't exist.
	ping.ping()
	ping.start()
}

func (ping *pinger) start() {
	go func() {
		for {
			<-ping.ticker.C
			ping.ping()
		}
	}()
}

func (ping *pinger) ping() {
	err := ping.riserClient.Environments.Ping(ping.environmentName)
	if err != nil {
		ping.log.Error(err, fmt.Sprintf("Error pinging environment %q", ping.environmentName))
	}
}
