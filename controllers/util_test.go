package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_getRiserRevision(t *testing.T) {
	objectMeta := metav1.ObjectMeta{
		Annotations: map[string]string{
			riserLabel("revision"): "2",
		},
	}
	result, err := getRiserRevision(objectMeta)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), result)
}

func Test_getRiserRevision_WhenBadValue(t *testing.T) {
	objectMeta := metav1.ObjectMeta{
		Annotations: map[string]string{
			riserLabel("revision"): "lol",
		},
	}
	result, err := getRiserRevision(objectMeta)

	assert.Equal(t, `Error parsing riser revision from annotation: riser.dev/revision: strconv.ParseInt: parsing "lol": invalid syntax`, err.Error())
	assert.Equal(t, int64(-1), result)
}
