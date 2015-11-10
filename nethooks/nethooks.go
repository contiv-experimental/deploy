package nethooks

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project"
)

const (
	applyLinksBasedPolicyFlag  = true
	applyLabelsBasedPolicyFlag = true
	applyDefaultPolicyFlag     = false
	applyContractPolicyFlag    = true
)

func applyLinksBasedPolicy(p *project.Project) error {
	//TODO allow tenant name to be specified
	name := "default"
	links, err := getSvcLinks(p)
	if err != nil {
		log.Debugf("Unable to find links from service chains. Error %v", err)
		return err
	}
	
	if err := addEpgs(p); err != nil {
		log.Errorf("Unable to apply policies for unspecified tiers. Error %v", err)
		return err
	}

	policyRecs := make(map[string]policyCreateRec)
	for fromSvcName, toSvcNames := range links {
		for _, toSvcName := range toSvcNames {
			log.Infof("Creating policy contract from service '%s' to services '%s'", fromSvcName, toSvcName)
			if err := applyInPolicy(p, fromSvcName, toSvcName); err != nil {
				log.Errorf("Unable to apply in-policy for service '%s'. Error %v", toSvcName, err)
				return err
			}

		}
	}

	spMap, err := getSvcPorts(p)
	if err != nil {
		log.Debugf("Unable to find exposed ports from service chains. Error %v", err)
		return err
	}
	if err := applyExposePolicy(p, spMap, policyRecs); err != nil {
		log.Errorf("Unable to apply expose-policy %v", err)
		return err
	}

	if err := addApp(name, p); err != nil {
		log.Errorf("Unable to create app with unspecified tiers. Error %v", err)
		return err
	}

	if applyDefaultPolicyFlag {
		if err := applyDefaultPolicy(p, policyRecs); err != nil {
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
		if err := clearExposedPorts(p); err != nil {
			log.Errorf("Unable to clear exposed ports. Error: %s", err)
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
	//TODO allow tenant name to be specified
	name := "default"

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

	if err := deleteApp(name, p); err != nil {
		log.Errorf("Unable to delete app. Error %v", err)
	}

	return nil
}

func AutoGenParams(p *project.Project) error {
	for svcName, svc := range p.Configs {
		if svc.Net == "" {
			svc.Net = getFullSvcName(p, svcName) + "." + getTenantName(nil)
		}
		if svc.Hostname == "" {
			svc.Hostname = p.Name + "_" + svcName + "_1"
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

func getContName(p *project.Project, svcName string) string {
	return p.Name + "_" + svcName + "_1"
}

func getSvcNameWithProject(p *project.Project, svcName string) string {
	return p.Name + "_" + svcName
}

func PopulateEtcHosts(p *project.Project) error {
	for dnsSvcName, _ := range p.Configs {
		dnsContName := getContName(p, dnsSvcName)
		dnsSvcIpAddress := getContainerIP(dnsContName)
		dnsSvcEntryName := getSvcNameWithProject(p, dnsSvcName)

		// TODO: need to populate all instances not just first instance
		for contSvcName, _ := range p.Configs {
			if contSvcName == dnsSvcName {
				continue
			}
			contName := getContName(p, contSvcName)
			if err := populateEtcHosts(contName, dnsSvcEntryName, dnsSvcIpAddress); err != nil {
				log.Errorf("Unable to populate /etc/hosts entry into container '%s' entry '%s %s'. Error %v",
					contName, dnsSvcEntryName, dnsSvcIpAddress)
			}

			if err := populateEtcHosts(contName, dnsSvcName, dnsSvcIpAddress); err != nil {
				log.Errorf("Unable to populate /etc/hosts entry into container '%s' entry '%s %s'. Error %v",
					contName, dnsSvcEntryName, dnsSvcIpAddress)
			}
			log.Debugf("populated dns: container '%s' svc '%s' with ip '%s'", contName, dnsSvcEntryName, dnsSvcIpAddress)
		}
	}
	return nil
}
