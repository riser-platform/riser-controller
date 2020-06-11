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
func StartNewPinger(riserClient *sdk.Client, log logr.Logger, environmentName string, pingFrequency time.Duration) error {
	ping := &pinger{riserClient, log, environmentName, time.NewTicker(pingFrequency)}
	// We block on the first ping since this bootstraps a new environment. We probably want to remove this in favor of environment config endpoints
	// on the server creating the environment if it doesn't exist.
	err := ping.ping()
	if err != nil {
		return err
	}
	ping.start()
	return nil
}

func (ping *pinger) start() {
	go func() {
		for {
			<-ping.ticker.C
			err := ping.ping()
			if err != nil {
				ping.log.Error(err, fmt.Sprintf("Error pinging environment %q", ping.environmentName))
			}
		}
	}()
}

func (ping *pinger) ping() error {
	return ping.riserClient.Environments.Ping(ping.environmentName)
}
