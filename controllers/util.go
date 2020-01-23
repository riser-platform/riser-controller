package controllers

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// riserLabel returns a fully qualified riser label or annotation (e.g. riser.dev/your-label)
func riserLabel(labelName string) string {
	return fmt.Sprintf("riser.dev/%s", labelName)
}

func isRiserApp(objectMeta metav1.ObjectMeta) bool {
	_, b := objectMeta.Labels[riserLabel("app")]
	return b
}

func riserAppFilter(objectMeta metav1.ObjectMeta) client.MatchingLabels {
	labels := map[string]string{
		riserLabel("deployment"): objectMeta.Labels[riserLabel("deployment")],
	}
	return labels
}

func getRiserGeneration(objectMeta metav1.ObjectMeta) (int64, error) {
	v, err := strconv.ParseInt(objectMeta.Annotations[riserLabel("generation")], 10, 64)
	if err != nil {
		return -1, errors.Wrap(err, fmt.Sprintf("Error parsing riser generation from annotation: %s", riserLabel("generation")))
	}
	return v, nil
}
