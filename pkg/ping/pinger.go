package ping

import (
	"fmt"
	"time"

	"github.com/riser-platform/riser-server/pkg/sdk"

	"github.com/go-logr/logr"
)

type pinger struct {
	riserClient *sdk.Client
	log         logr.Logger
	stageName   string
	ticker      *time.Ticker
}

/*
StartNewPinger creates and starts a new pinger. The server maintains the "last ping" from a controller to track when the last communication was received
from the controller. A ping happens automatically when a status update is received, however, since there can be periods with few status updates to
the server, a "pinger" is needed to inform the server of connectivity.
*/
func StartNewPinger(riserClient *sdk.Client, log logr.Logger, stageName string, pingFrequency time.Duration) {
	ping := &pinger{riserClient, log, stageName, time.NewTicker(pingFrequency)}
	// We block on the first ping since this bootstraps a new stage. We probably want to remove this in favor of stage config endpoints
	// on the server creating the stage if it doesn't exist.
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
	err := ping.riserClient.Stages.Ping(ping.stageName)
	if err != nil {
		ping.log.Error(err, fmt.Sprintf("Error pinging stage %q", ping.stageName))
	}
}
