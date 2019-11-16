package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_getRiserGeneration(t *testing.T) {
	objectMeta := metav1.ObjectMeta{
		Annotations: map[string]string{
			riserLabel("generation"): "2",
		},
	}
	result, err := getRiserGeneration(objectMeta)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), result)
}

func Test_getRiserGeneration_WhenBadValue(t *testing.T) {
	objectMeta := metav1.ObjectMeta{
		Annotations: map[string]string{
			riserLabel("generation"): "lol",
		},
	}
	result, err := getRiserGeneration(objectMeta)

	assert.Equal(t, `Error parsing riser generation from annotation: riser.dev/generation: strconv.ParseInt: parsing "lol": invalid syntax`, err.Error())
	assert.Equal(t, int64(-1), result)
}
