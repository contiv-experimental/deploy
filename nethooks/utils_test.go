package nethooks

import (
	"testing"
)

const (
	imageName = "redis"
)

func TestPortID(t *testing.T) {
	portID, protoID, err := getPortID(imageName)
	if err != nil {
		t.Errorf("Unable to get port id for image %s. Error %v \n", imageName, err)
	}
}
