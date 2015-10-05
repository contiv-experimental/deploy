package nethooks

import (
	log "github.com/Sirupsen/logrus"
	"github.com/contiv/objmodel/contivModel"
	"github.com/docker/libcompose/project"
)

const (
	baseURL = "http://netmaster:9999/api/"
)

func getRulePath(tenantName, policyName, ruleID string) string {
	return baseURL + "rules/" + tenantName + ":" + policyName + ":" + ruleID + "/"
}
func getPolicyRulesPath(tenantName, policyName string) string {
	return baseURL + "rules/" + tenantName + ":" + policyName + "/"
}

func getPolicyPath(tenantName, policyName string) string {
	return baseURL + "policys/" + tenantName + ":" + policyName + "/"
}

func getEpgPath(tenantName, groupName string) string {
	return baseURL + "endpointGroups/" + tenantName + ":" + groupName + "/"
}

func getTenantName(labels map[string]string) string {
	tenantName := TENANT_DEFAULT
	if labels != nil {
		if value, ok := labels[TENANT_LABEL]; ok {
			tenantName = value
		}
	}
	return tenantName
}

func getNetworkName(labels map[string]string) string {
	networkName := NETWORK_DEFAULT
	if labels != nil {
		if value, ok := labels[NET_LABEL]; ok {
			networkName = value
		}
	}
	return networkName
}

func getSvcLinks(p *project.Project) (map[string][]string, error) {
	links := make(map[string][]string)

	for svcName, svc := range p.Configs {
		svcLinks := svc.Links.Slice()
		log.Debugf("found links for svc '%s' %#v ", svcName, svcLinks)
		links[svcName] = svcLinks
	}

	return links, nil
}

func clearSvcLinks(p *project.Project) error {
	for svcName, svc := range p.Configs {
		svc.Links = project.NewMaporColonSlice([]string{})
		log.Debugf("clearing links for svc '%s' %#v ", svcName, svc.Links)
	}
	return nil
}

func addDenyAllRule(tenantName, networkName, epgName, policyName string, ruleID int) error {
	// create rules and policy
	rule := &contivModel.Rule{
		Action:        "deny",
		Direction:     "in",
		EndpointGroup: epgName,
		Network:       networkName,
		PolicyName:    policyName,
		Priority:      1,
		RuleID:        string(ruleID),
		TenantName:    tenantName,
	}
	if err := httpPost(getRulePath(rule.TenantName, rule.PolicyName, rule.RuleID), rule); err != nil {
		log.Errorf("Unable to create policy rule. Error: %v", err)
		return err
	}

	return nil
}

func addPermitRule(tenantName, networkName, epgName, policyName string, ruleID int) error {
	if err := addDenyAllRule(tenantName, networkName, epgName, policyName, ruleID); err != nil {
		return err
	}

	ruleID++

	rule := &contivModel.Rule{
		Action:        "permit",
		Direction:     "in",
		EndpointGroup: epgName,
		Network:       networkName,
		PolicyName:    policyName,
		Priority:      2,
		Protocol:      "tcp",
		RuleID:        string(ruleID),
		TenantName:    tenantName,
	}
	if err := httpPost(getRulePath(rule.TenantName, rule.PolicyName, rule.RuleID), rule); err != nil {
		log.Errorf("Unable to create policy rule. Error: %v", err)
		return err
	}

	return nil
}

func addPolicy(tenantName, policyName string) error {
	policy := &contivModel.Policy{
		PolicyName: policyName,
		TenantName: tenantName,
	}
	if err := httpPost(getPolicyPath(policy.TenantName, policy.PolicyName), policy); err != nil {
		log.Errorf("Unable to create policy rule. Error: %v", err)
		return err
	}

	return nil
}

func addEpg(tenantName, networkName, epgName string, policies []string) error {
	// create epgs
	epg := &contivModel.EndpointGroup{
		EndpointGroupID: 1,
		GroupName:       epgName,
		NetworkName:     networkName,
		Policies:        policies,
		TenantName:      tenantName,
	}
	if err := httpPost(getEpgPath(epg.TenantName, epg.GroupName), epg); err != nil {
		log.Errorf("Unable to create endpoint group. Error: %v", err)
		return err
	}

	return nil
}

// CreateNetConfig creates the netmaster configuration
// It also updates the project with information if needed before project up
func CreateNetConfig(p *project.Project) error {
	log.Debugf("Create network for the project '%s' ", p.Name)

	links, err := getSvcLinks(p)
	if err != nil {
		log.Debugf("Unable to find links from service chains. Error %v", err)
		return err
	}

	policyApplied := make(map[string]bool)
	for fromSvcName, toSvcNames := range links {
		log.Debugf("Initiating contracts from service '%s' to services %s", fromSvcName, toSvcNames)
		for _, svcName := range toSvcNames {
			svc := p.Configs[svcName]

			tenantName := getTenantName(svc.Labels.MapParts())
			networkName := getNetworkName(svc.Labels.MapParts())
			policyName := svcName + "-in"
			epgName := svcName
			ruleID := 1
			policies := []string{}

			log.Debugf("Creating network objects to service '%s': Tenant: %s Network %s", svcName, tenantName, networkName)

			if err := addPermitRule(tenantName, networkName, epgName, policyName, ruleID); err != nil {
				log.Errorf("Unable to add permit rule. Error %v ", err)
				return err
			}

			if err := addPolicy(tenantName, policyName); err != nil {
				log.Errorf("Unable to add policy. Error %v ", err)
				return err
			}
			policies = append(policies, policyName)

			if err := addEpg(tenantName, networkName, epgName, policies); err != nil {
				log.Errorf("Unable to add epg. Error %v", err)
				return err
			}

			policyApplied[svcName] = true
		}
	}

	// apply the policy on the services on the elements that we didn't get to so far
	for svcName, _ := range p.Configs {
		if _, ok := policyApplied[svcName]; !ok {
			svc := p.Configs[svcName]

			log.Debugf("Applying deny all in policy for service '%s' ", svcName)
			tenantName := getTenantName(svc.Labels.MapParts())
			networkName := getNetworkName(svc.Labels.MapParts())
			policyName := svcName + "-in"
			epgName := svcName
			ruleID := 1

			if err := addDenyAllRule(tenantName, networkName, epgName, policyName, ruleID); err != nil {
				return err
			}
		}
	}

	if err := clearSvcLinks(p); err != nil {
		log.Errorf("Unable to clear service links. Error: %s", err)
	}

	return nil
}

// DeleteNetConfig removes the netmaster configuraton
func DeleteNetConfig(p *project.Project) error {
	log.Debugf("Delete network for the project '%s' ", p.Name)

	for svcName, _ := range p.Configs {
		svc := p.Configs[svcName]

		log.Debugf("Deleting policies for service '%s' ", svcName)
		tenantName := getTenantName(svc.Labels.MapParts())
		networkName := getNetworkName(svc.Labels.MapParts())
		policyName := svcName + "-in"
		epgName := svcName

		log.Debugf("Deleting network objects to service '%s': Tenant: %s Network %s", svcName, tenantName, networkName)

		for ruleID := 1; ruleID <= 2; ruleID++ {
			rulePath := getRulePath(tenantName, policyName, string(ruleID))
			if err := httpDelete(rulePath); err != nil {
				log.Errorf("Unable to delete '%s' rule. Error: %v", rulePath, err)
			}
		}

		policyPath := getPolicyPath(tenantName, policyName)
		if err := httpDelete(policyPath); err != nil {
			log.Errorf("Unable to delete '%s' policy. Error: %v", policyPath, err)
		}

		epgPath := getEpgPath(tenantName, epgName)
		if err := httpDelete(epgPath); err != nil {
			log.Errorf("Unable to delete '%s' epg. Error: %v", epgPath, err)
		}

	}

	return nil
}

func AutoGenParams(p *project.Project) error {
	for svcName, svc := range p.Configs {
		if svc.PublishService == "" {
			svc.PublishService = "default:" + svcName
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
