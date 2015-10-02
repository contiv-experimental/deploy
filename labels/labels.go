package labels

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/netplugin/core"
	"github.com/docker/libcompose/project"
)

func Parse(csvLabels string) (map[string]string, error) {
	labels := map[string]string{}

	csvRecs := strings.Split(csvLabels, ",")
	if len(csvRecs) == 0 {
		return nil, core.Errorf("unable to parse labls '%s'", csvLabels)
	}

	for _, csvRec := range csvRecs {
		csvRec := strings.Trim(csvRec, ", ")
		csvValues := strings.Split(csvRec, ":")
		if len(csvValues) != 2 {
			log.Errorf("unable to parse labels: '%s'", csvRec)
			return nil, core.Errorf("unable to parse label record: '%s'", csvRec)
		}

		labels[csvValues[0]] = csvValues[1]
	}

	return labels, nil
}

func Insert(p *project.Project, parts map[string]string) error {
	for svcName, svc := range p.Configs {
		origParts := svc.Labels.MapParts()
		if origParts == nil {
			origParts = make(map[string]string)
		}

		for partKey, partValue := range parts {
			origParts[partKey] = partValue
		}
		log.Debugf("Updated composition labels for service %s: %#v", svcName, origParts)
		svc.Labels = project.NewSliceorMap(origParts)
	}

	return nil
}
