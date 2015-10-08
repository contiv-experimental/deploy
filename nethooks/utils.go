
package nethooks

import (
	"os"
	"strings"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

type imageInfo struct {
	portID int
	protoName string
}

func getImageInfo(imageName string) ([]imageInfo, error) {
	imageInfoList := []imageInfo{}

	docker, err := dockerclient.NewDockerClient(os.Getenv("DOCKER_HOST"), nil)
	if err != nil {
		log.Errorf("Unable to connect to docker. Error %v", err)
		return imageInfoList, err
	}

	info, err := docker.InspectImage(imageName)
	log.Infof("Got the following info for the image %#v", info)

	if err != nil {
		log.Errorf("Unable to inspect image '%s'. Error %v", imageName, err)
		return imageInfoList, err
	}

	for exposedPort := range info.Config.ExposedPorts {
		if strings.Contains(exposedPort, "/") {
			imageInfo := imageInfo{}
			values := strings.Split(exposedPort, "/")
			imageInfo.portID, _ = strconv.Atoi(values[0])
			imageInfo.protoName = values[1]
			log.Infof("Extracted port info %v from image '%s'", imageInfo, imageName)
			imageInfoList = append(imageInfoList, imageInfo)
		}
	}

	return imageInfoList, nil
}
