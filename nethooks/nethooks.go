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

func getFullSvcName(p *project.Project, svcName string) string {
	if p == nil {
		return svcName
	}

	return p.Name + "_" + svcName + "." + NETWORK_DEFAULT
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

func addDenyAllRule(tenantName, networkName, fromEpgName, policyName string, ruleID int) error {
	// create rules and policy
	rule := &contivModel.Rule{
		Action:        "deny",
		Direction:     "in",
		EndpointGroup: fromEpgName,
		Network:       networkName,
		PolicyName:    policyName,
		Priority:      1,
		RuleID:        string(ruleID + '0'),
		TenantName:    tenantName,
	}
	if err := httpPost(getRulePath(rule.TenantName, rule.PolicyName, rule.RuleID), rule); err != nil {
		log.Errorf("Unable to create deny all rule %#v. Error: %v", rule, err)
		return err
	}

	return nil
}

func addInAcceptRule(tenantName, networkName, fromEpgName, policyName string, portID, ruleID int) error {
	if err := addDenyAllRule(tenantName, networkName, fromEpgName, policyName, ruleID); err != nil {
		return err
	}

	ruleID++

	rule := &contivModel.Rule{
		Action:        "accept",
		Direction:     "in",
		EndpointGroup: fromEpgName,
		Network:       networkName,
		PolicyName:    policyName,
		Port:		   portID,
		Priority:      2,
		Protocol:      "tcp",
		RuleID:        string(ruleID + '0'),
		TenantName:    tenantName,
	}
	if err := httpPost(getRulePath(rule.TenantName, rule.PolicyName, rule.RuleID), rule); err != nil {
		log.Errorf("Unable to create accept rule %#v. Error: %v", rule, err)
		return err
	}

	return nil
}

func addOutAcceptAllRule(tenantName, networkName, fromEpgName, policyName string, ruleID int) error {
	rule := &contivModel.Rule{
		Action:        "accept",
		Direction:     "out",
		EndpointGroup: fromEpgName,
		Network:       networkName,
		PolicyName:    policyName,
		Priority:      2,
		Protocol:      "tcp",
		RuleID:        string(ruleID + '0'),
		TenantName:    tenantName,
	}
	if err := httpPost(getRulePath(rule.TenantName, rule.PolicyName, rule.RuleID), rule); err != nil {
		log.Errorf("Unable to create accept rule %#v. Error: %v", rule, err)
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

func applyUnspecifiedPolicy(p *project.Project, policyApplied map[string]bool) error {
	for svcName, _ := range p.Configs {
		svc := p.Configs[svcName]
		tenantName := getTenantName(svc.Labels.MapParts())
		networkName := getNetworkName(svc.Labels.MapParts())
		epgName := getFullSvcName(p, svcName)
		fromEpgName := ""

		if _, ok := policyApplied[svcName]; ok {
			continue
		}

		// add 'in' policy for the service tier
		ruleID := 1
		policyName := svcName + "-in"
		policies := []string{}

		log.Debugf("Applying deny all in policy for service '%s' ", svcName)
		if err := addPolicy(tenantName, policyName); err != nil {
			log.Errorf("Unable to add policy. Error %v ", err)
			return err
		}
		policies = append(policies, policyName)

		if err := addDenyAllRule(tenantName, networkName, fromEpgName, policyName, ruleID); err != nil {
			log.Errorf("Unable to add deny rule. Error %v ", err)
			return err
		}

		// add 'out' policy for the service tier
		ruleID = 1
		policyName = svcName + "-out"
		if err := addPolicy(tenantName, policyName); err != nil {
			log.Errorf("Unable to add policy. Error %v", err)
		}
		policies = append(policies, policyName)
		if err := addOutAcceptAllRule(tenantName, networkName, fromEpgName, policyName, ruleID); err != nil {
			log.Errorf("Unable to add deny rule. Error %v ", err)
			return err
		}


		// add epg with in and out policies
		if err := addEpg(tenantName, networkName, epgName, policies); err != nil {
			log.Errorf("Unable to add epg. Error %v", err)
			return err
		}
	}

	return nil
}

// CreateNetConfig creates the netmaster configuration
// It also updates the project with information if needed before project up
func CreateNetConfig(p *project.Project) error {
	skipUnspecified := false

	log.Debugf("Create network for the project '%s' ", p.Name)

	for _, svc := range p.Configs {
		log.Debugf("===> VJ service %s, publishService %s ", svc.Name, svc.PublishService)
	}

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
			epgName := getFullSvcName(p, svcName)
			fromEpgName := ""		// no contract rules for now
			ruleID := 1
			policies := []string{}

			log.Debugf("Creating network objects to service '%s': Tenant: %s Network %s", svcName, tenantName, networkName)

			if err := addPolicy(tenantName, policyName); err != nil {
				log.Errorf("Unable to add policy. Error %v ", err)
				return err
			}
			policies = append(policies, policyName)

			if err := addInAcceptRule(tenantName, networkName, fromEpgName, policyName, 6379, ruleID); err != nil {
				log.Errorf("Unable to add accept rule. Error %v ", err)
				return err
			}

			if err := addEpg(tenantName, networkName, epgName, policies); err != nil {
				log.Errorf("Unable to add epg. Error %v", err)
				return err
			}

			policyApplied[svcName] = true
		}
	}

	if skipUnspecified {
		err = applyUnspecifiedPolicy(p, policyApplied)
		if err != nil {
			log.Errorf("Unable to apply policies for unspecified tiers. Error %v", err)
			return err
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
