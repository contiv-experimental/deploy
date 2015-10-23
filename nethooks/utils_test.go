package nethooks

import (
	"testing"
)

const (
	imageName = "redis"
)

func TestGetImageInfo(t *testing.T) {
	imageInfo, err := getImageInfo(imageName)
	if err != nil {
		t.Errorf("Unable to get port id for image %s. Error %v \n", imageInfo, err)
	}
}
