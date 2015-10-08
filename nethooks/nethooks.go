package nethooks

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project"
)

const (
	applyLinksBasedPolicyFlag = true
	applyLabelsBasedPolicyFlag = true
	applyDefaultPolicyFlag = false
)

func applyLinksBasedPolicy(p *project.Project) error {
	links, err := getSvcLinks(p)
	if err != nil {
		log.Debugf("Unable to find links from service chains. Error %v", err)
		return err
	}

	policyApplied := make(map[string]bool)
	for fromSvcName, toSvcNames := range links {
		log.Debugf("Initiating contracts from service '%s' to services %s", fromSvcName, toSvcNames)
		for _, toSvcName := range toSvcNames {
			if err := applyInPolicy(p, toSvcName); err != nil {
				log.Errorf("Unable to apply in-policy for service '%s'. Error %v", toSvcName, err)
				return err
			}

			policyApplied[toSvcName] = true
		}
	}

	if applyDefaultPolicyFlag {
		if err := applyDefaultPolicy(p, policyApplied); err != nil {
			log.Errorf("Unable to apply policies for unspecified tiers. Error %v", err)
			return err
		}
	} else {
		if err := addEpgs(p, policyApplied); err != nil {
			log.Errorf("Unable to apply policies for unspecified tiers. Error %v", err)
			return err
		}
	}

	return nil
}

// CreateNetConfig creates the netmaster configuration
// It also updates the project with information if needed before project up
func CreateNetConfig(p *project.Project) error {
	log.Debugf("Create network for the project '%s' ", p.Name)

	if applyLinksBasedPolicyFlag {
		if err := applyLinksBasedPolicy(p); err != nil {
			log.Errorf("Unable to apply links based policy. Error: %s", err)
		}
		if err := clearSvcLinks(p); err != nil {
			log.Errorf("Unable to clear service links. Error: %s", err)
		}
	}
	if applyLabelsBasedPolicyFlag {
		log.Infof("Applying labels based policies")
	}

	return nil
}

// DeleteNetConfig removes the netmaster configuraton
func DeleteNetConfig(p *project.Project) error {
	log.Debugf("Delete network for the project '%s' ", p.Name)

	for svcName, _ := range p.Configs {
		if err := removeEpg(p, svcName); err != nil {
			log.Errorf("Unable to remove out-policy for service '%s'. Error %v", svcName, err)
		}

		if err := removePolicy(p, svcName, "in"); err != nil {
			log.Errorf("Unable to remove in-policy for service '%s'. Error %v", svcName, err)
		}

		if err := removePolicy(p, svcName, "out"); err != nil {
			log.Errorf("Unable to remove out-policy for service '%s'. Error %v", svcName, err)
		}

		if err := clearSvcLinks(p); err != nil {
			log.Errorf("Unable to clear service links. Error: %s", err)
		}
	}

	return nil
}

func AutoGenParams(p *project.Project) error {
	for svcName, svc := range p.Configs {
		if svc.PublishService == "" {
			svc.PublishService = getFullSvcName(p, svcName)
		}
	}

	return nil
}

func AutoGenLabels(p *project.Project) error {
	for svcName, svc := range p.Configs {
		labels := svc.Labels.MapParts()
		if labels == nil {
			labels = make(map[string]string)
		}

		labels[EPG_LABEL] = svcName

		svc.Labels = project.NewSliceorMap(labels)
	}

	return nil
}
