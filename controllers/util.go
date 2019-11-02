package controllers

import (
	"fmt"

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
