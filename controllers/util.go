package controllers

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// riserLabel returns a fully qualified riser label or annotation (e.g. riser.dev/your-label)
func riserLabel(labelName string) string {
	return fmt.Sprintf("riser.dev/%s", labelName)
}

func isRiserApp(obj metav1.Object) bool {
	_, b := obj.GetLabels()[riserLabel("app")]
	return b
}

func riserAppFilter(objectMeta metav1.ObjectMeta) client.MatchingLabels {
	labels := map[string]string{
		riserLabel("deployment"): objectMeta.Labels[riserLabel("deployment")],
	}
	return labels
}

func getRiserRevision(objectMeta metav1.ObjectMeta) (int64, error) {
	v, err := strconv.ParseInt(objectMeta.Annotations[riserLabel("revision")], 10, 64)
	if err != nil {
		return -1, errors.Wrap(err, fmt.Sprintf("Error parsing riser revision from annotation: %s", riserLabel("revision")))
	}
	return v, nil
}
