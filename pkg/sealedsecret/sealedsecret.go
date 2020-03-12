package sealedsecret

import (
	"time"

	"github.com/riser-platform/riser-server/pkg/sdk"

	"github.com/riser-platform/riser-server/api/v1/model"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1Client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const retryOnFailureSeconds = 5 * time.Second

type refresher struct {
	stageName           string
	controllerName      string
	controllerNamespace string
	log                 logr.Logger
	ticker              *time.Ticker
	kubeClient          *corev1Client.CoreV1Client
	riserClient         *sdk.Client
}

/*
	StartCertRefresher polls the sealed secret controller on an interval and updates the riser server with the latest cert (public key)
	This is not a controller because the only resources we can monitor are secrets which include the private key. Riser should never have access
	to the private key. The only way to safely get the public key is with access to the sealed secret API. Additionally, it is not critical
	that the key is refreshed at the same time that it is rotated within the sealed secret controller.  The goal is
	to rotate the key within N duration. At the time of writing the default sealed secret rotation is set to 30 days, and the refresh duration is set
	to 1 day. This means that for a period of a maximum of 1 day that new secrets stored via Riser will still use the old key. This is perfectly
	valid. The sealed secrets controller maintains a history of keys and will not delete them. If for some reason you need keys rotated within 30
	days, it's better to configure the sealed secrets controller to a shorter duration (e.g. 15 days) than to change the refresh frequency.

	If timely cert refreshing is critical it's important to setup monitoring. Errors other than on startup are logged but are not considered fatal.
	Other operations such	as reporting status should not be affected. The only case where this is truly is when there is a new stage that does not
	have any cert, in	which case no secrets can be saved.

	Read https://github.com/bitnami-labs/sealed-secrets#secret-rotation for more info.
*/
func StartCertRefresher(kubeConfig *rest.Config, riserClient *sdk.Client, stageName string, controllerName string, controllerNamespace string, refreshInterval time.Duration, log logr.Logger) error {
	client, err := corev1Client.NewForConfig(kubeConfig)
	if err != nil {
		return errors.Wrap(err, "Unable to create rest client for sealed secret cert refresher")
	}
	refresher := refresher{
		stageName:           stageName,
		controllerName:      controllerName,
		controllerNamespace: controllerNamespace,
		log:                 log,
		kubeClient:          client,
		ticker:              time.NewTicker(refreshInterval),
		riserClient:         riserClient,
	}

	refresher.refresh()
	refresher.start()
	return nil
}

func (r *refresher) start() {
	go func() {
		for {
			<-r.ticker.C
			r.refresh()
		}
	}()
}

func (r *refresher) refresh() {
	certBytes, err := r.kubeClient.Services(r.controllerNamespace).
		ProxyGet("http", r.controllerName, "", "/v1/cert.pem", nil).
		DoRaw()
	config := &model.StageConfig{
		SealedSecretCert: certBytes,
	}

	if err == nil {
		r.log.Info("Updating cert for sealed secrets")
		err = r.riserClient.Stages.SetConfig(r.stageName, config)
	}

	if err != nil {
		r.log.Error(err, "Error setting stage config. Retrying...")
		time.AfterFunc(retryOnFailureSeconds, r.refresh)
	}

}
