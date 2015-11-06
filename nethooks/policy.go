package nethooks

import (
	log "github.com/Sirupsen/logrus"
	"github.com/contiv/objmodel/contivModel"
	"github.com/contiv/objmodel/objdb/modeldb"
	"github.com/docker/libcompose/project"
)

const (
	baseURL = "http://netmaster:9999/api/"
)

func getRulePath(tenantName, policyName, ruleID string) string {
	return baseURL + "rules/" + tenantName + ":" + policyName + ":" + ruleID + "/"
}

func getRulePathFromName(ruleName string) string {
	return baseURL + "rules/" + ruleName
}

func getPolicyRulesPath(tenantName, policyName string) string {
	return baseURL + "rules/" + tenantName + ":" + policyName + "/"
}

func getPolicyPath(tenantName, policyName string) string {
	return baseURL + "policys/" + tenantName + ":" + policyName + "/"
}

func getAppPath(tenantName, appName string) string {
	return baseURL + "apps/" + tenantName + ":" + appName + "/"
}

func getEpgPath(tenantName, networkName, groupName string) string {
	return baseURL + "endpointGroups/" + tenantName + ":" + networkName + ":" + groupName + "/"
}

func getRuleStr(ruleID int) string {
	return string(ruleID + '0')
}

func getInPolicyStr(projectName, svcName string) string {
	return projectName + "_" + svcName + "-in"
}

func getOutPolicyStr(projectName, svcName string) string {
	return projectName + "_" + svcName + "-out"
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

func getSvcName(p *project.Project, svcName string) string {
	if p == nil {
		return svcName
	}

	return p.Name + "_" + svcName
}

func getFromEpgName(p *project.Project, fromSvcName string) string {
	if applyContractPolicyFlag {
		return getSvcName(p, fromSvcName)
	}

	return ""
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
		// if len(svc.Links.Slice()) > 0 {
		svc.Links = project.NewMaporColonSlice([]string{})
		log.Debugf("clearing links for svc '%s' %#v ", svcName, svc.Links)
		// }
	}
	return nil
}

func clearExposedPorts(p *project.Project) error {
	for svcName, svc := range p.Configs {
		if len(svc.Expose) > 0 {
			svc.Expose = []string{}
			log.Debugf("clearing exposed ports for svc '%s' %#v ", svcName, svc.Links)
		}
		svc.Ports = []string{}
	}
	return nil
}

func addDenyAllRule(tenantName, networkName, fromEpgName, policyName string, ruleID int) error {
	rule := &contivModel.Rule{
		Action:        "deny",
		Direction:     "in",
		EndpointGroup: fromEpgName,
		Network:       networkName,
		PolicyName:    policyName,
		Priority:      ruleID,
		Protocol:      "tcp",
		RuleID:        getRuleStr(ruleID),
		TenantName:    tenantName,
	}
	if err := httpPost(getRulePath(rule.TenantName, rule.PolicyName, rule.RuleID), rule); err != nil {
		log.Errorf("Unable to create deny all rule %#v. Error: %v", rule, err)
		return err
	}

	return nil
}

func addInAcceptRule(tenantName, networkName, fromEpgName, policyName, protoName string, portID, ruleID int) error {
	rule := &contivModel.Rule{
		Action:        "accept",
		Direction:     "in",
		EndpointGroup: fromEpgName,
		Network:       networkName,
		PolicyName:    policyName,
		Port:          portID,
		Priority:      ruleID,
		Protocol:      protoName,
		RuleID:        getRuleStr(ruleID),
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
		Priority:      ruleID,
		Protocol:      "tcp",
		RuleID:        getRuleStr(ruleID),
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
		log.Debugf("Unable to create policy rule. Error: %v", err)
		return err
	}

	return nil
}

func addApp(tenantName string, p *project.Project) error {

	log.Debugf("Entered addApp '%s':'%s' ", tenantName, p.Name)
	app := &contivModel.App{
		AppName:    p.Name,
		TenantName: tenantName,
	}

	// Add services
	for svcName := range p.Configs {
		epgKey := TENANT_DEFAULT + ":" + NETWORK_DEFAULT + ":" + getSvcName(p, svcName)
		epg := &contivModel.EndpointGroup{
			Key: epgKey,
		}

		if err := modeldb.AddLinkSet(&app.LinkSets.Services, epg); err != nil {
			log.Errorf("addApp:Unable to add link for service '%s'. Error %v", svcName, err)
			return err
		}
		log.Debugf("addApp add link for:'%s' ", epgKey)
	}

	if err := httpPost(getAppPath(tenantName, p.Name), app); err != nil {
		log.Errorf("Unable to post app to netmaster. Error: %v", err)
		return err
	}

	return nil
}

func addEpg(tenantName, networkName, epgName string, policies []string) error {
	epg := &contivModel.EndpointGroup{
		EndpointGroupID: 1,
		GroupName:       epgName,
		NetworkName:     networkName,
		Policies:        policies,
		TenantName:      tenantName,
	}
	if err := httpPost(getEpgPath(epg.TenantName, epg.NetworkName, epg.GroupName), epg); err != nil {
		log.Errorf("Unable to create endpoint group. Tenant '%s' Network '%s' Epg '%s'. Error %v",
			tenantName, networkName, epgName, err)
		return err
	}

	return nil
}

func addEpgs(p *project.Project) error {
	for svcName, svc := range p.Configs {
		tenantName := getTenantName(svc.Labels.MapParts())
		networkName := getNetworkName(svc.Labels.MapParts())
		epgName := getSvcName(p, svcName)

		if err := addEpg(tenantName, networkName, epgName, []string{}); err != nil {
			log.Errorf("Unable to add epg for service '%s'. Error %v", svcName, err)
			return err
		}
	}
	return nil
}

func applyDefaultPolicy(p *project.Project, policyApplied map[string]bool) error {
	for svcName, svc := range p.Configs {
		tenantName := getTenantName(svc.Labels.MapParts())
		networkName := getNetworkName(svc.Labels.MapParts())
		toEpgName := getSvcName(p, svcName)

		if _, ok := policyApplied[svcName]; ok {
			continue
		}

		// add 'in' policy for the service tier
		ruleID := 1
		policyName := getInPolicyStr(p.Name, svcName)
		policies := []string{}

		log.Debugf("Applying deny all in policy for service '%s' ", svcName)
		if err := addPolicy(tenantName, policyName); err != nil {
			log.Errorf("Unable to add policy. Error %v ", err)
			return err
		}
		policies = append(policies, policyName)

		if err := addDenyAllRule(tenantName, networkName, "", policyName, ruleID); err != nil {
			log.Errorf("Unable to add deny rule. Error %v ", err)
			return err
		}

		// add 'out' policy for the service tier
		ruleID = 1
		policyName = getOutPolicyStr(p.Name, svcName)
		if err := addPolicy(tenantName, policyName); err != nil {
			log.Errorf("Unable to add policy. Error %v", err)
		}
		policies = append(policies, policyName)
		if err := addOutAcceptAllRule(tenantName, networkName, "", policyName, ruleID); err != nil {
			log.Errorf("Unable to add deny rule. Error %v ", err)
			return err
		}

		// add epg with in and out policies
		if err := addEpg(tenantName, networkName, toEpgName, policies); err != nil {
			log.Errorf("Unable to add epg. Error %v", err)
			return err
		}
	}

	return nil
}

func applyInPolicy(p *project.Project, fromSvcName, toSvcName string) error {
	svc := p.Configs[toSvcName]

	tenantName := getTenantName(svc.Labels.MapParts())
	networkName := getNetworkName(svc.Labels.MapParts())
	toEpgName := getSvcName(p, toSvcName)

	policyName := getInPolicyStr(p.Name, toSvcName)
	fromEpgName := getFromEpgName(p, fromSvcName)

	ruleID := 1
	policies := []string{}

	imageInfoList, err := getImageInfo(toSvcName)
	if err != nil {
		log.Infof("Unable to auto fetch port/protocol information. Error %v", err)
	}

	log.Debugf("Creating network objects to service '%s': Tenant: %s Network %s", toSvcName, tenantName, networkName)

	if err := addPolicy(tenantName, policyName); err != nil {
		log.Errorf("Unable to add policy. Error %v ", err)
		return err
	}
	policies = append(policies, policyName)

	if err := addDenyAllRule(tenantName, networkName, "", policyName, ruleID); err != nil {
		return err
	}
	ruleID++

	for _, imageInfo := range imageInfoList {
		if err := addInAcceptRule(tenantName, networkName, fromEpgName, policyName, imageInfo.protoName, imageInfo.portID, ruleID); err != nil {
			log.Errorf("Unable to add accept rule. Error %v ", err)
			return err
		}
		ruleID++
	}

	if err := addEpg(tenantName, networkName, toEpgName, policies); err != nil {
		log.Errorf("Unable to add epg. Error %v", err)
		return err
	}

	return nil
}

func removePolicy(p *project.Project, svcName, dir string) error {
	svc := p.Configs[svcName]

	log.Debugf("Deleting policies for service '%s' ", svcName)
	tenantName := getTenantName(svc.Labels.MapParts())
	networkName := getNetworkName(svc.Labels.MapParts())
	policyName := getInPolicyStr(p.Name, svcName)
	if dir == "out" {
		policyName = getOutPolicyStr(p.Name, svcName)
	}

	log.Debugf("Deleting network objects to service '%s': Tenant: %s Network %s", svcName, tenantName, networkName)

	policyPath := getPolicyPath(tenantName, policyName)

	policy := contivModel.Policy{}
	if err := httpGet(policyPath, &policy); err != nil {
		log.Debugf("Unable to delete policy for service '%s' policy %s", svcName, policyName)
		return nil
	}

	for ruleName, _ := range policy.LinkSets.Rules {
		rulePath := getRulePathFromName(ruleName)
		if err := httpDelete(rulePath); err != nil {
			log.Errorf("Unable to delete '%s' rule. Error: %v", rulePath, err)
		}
	}

	if err := httpDelete(policyPath); err != nil {
		log.Errorf("Unable to delete '%s' policy. Error: %v", policyPath, err)
	}

	return nil
}

func removeEpg(p *project.Project, svcName string) error {
	svc := p.Configs[svcName]

	log.Debugf("Deleting Epg for service '%s' ", svcName)
	tenantName := getTenantName(svc.Labels.MapParts())
	networkName := getNetworkName(svc.Labels.MapParts())
	epgName := getSvcName(p, svcName)

	epgPath := getEpgPath(tenantName, networkName, epgName)
	if err := httpDelete(epgPath); err != nil {
		log.Errorf("Unable to delete '%s' epg. Error: %v", epgPath, err)
	}

	return nil
}
