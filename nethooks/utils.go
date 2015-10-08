
package nethooks

import (
	"errors"
	"os"
	"strings"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

const (
	invalidPort = 0
	invalidProto = ""
)

func getImageInfo(imageName string) (int, string, error) {
	docker, err := dockerclient.NewDockerClient(os.Getenv("DOCKER_HOST"), nil)
	if err != nil {
		log.Errorf("Unable to connect to docker. Error %v", err)
		return invalidPort, invalidProto, err
	}

	info, err := docker.InspectImage(imageName)
	log.Infof("Got the following info for the image %#v", info)

	if err != nil {
		log.Errorf("Unable to inspect image '%s'. Error %v", imageName, err)
		return invalidPort, invalidProto, err
	}

	// TODO: support for multiple exposed ports
	for exposedPort := range info.Config.ExposedPorts {
		if strings.Contains(exposedPort, "/") {
			values := strings.Split(exposedPort, "/")
			portID, _ := strconv.Atoi(values[0])
			protoID := values[1]
			log.Infof("Extracted the port '%d' proto %s for '%s'", portID, protoID, imageName)
			return portID, protoID, nil
		}
	}

	return invalidPort, invalidProto, errors.New("unable to parse exposed ports")
}
