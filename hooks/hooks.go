package hooks

import (
	"github.com/contiv/deploy/labels"
	"github.com/contiv/deploy/nethooks"
	"github.com/docker/libcompose/project"
)

type eventType int

const (
	noEvent = eventType(iota)
	startEvent = eventType(iota)
	stopEvent = eventType(iota)
	scaleEvent = eventType(iota)
)

// validation of events that are supported
func getEvent(event string) eventType {
	switch event {
	case "up", "start":
		return startEvent
	case "down", "delete", "kill", "rm", "stop":
		return stopEvent
    case "scale":
		return scaleEvent
	case "create", "build", "ps", "port", "pull", "log", "restart":	
		// unsupported
	}

	return noEvent
}

func PopulateEnvLabels(p *project.Project, csvLabels string) error {
	parts, err := labels.Parse(csvLabels)
	if err != nil {
		return err
	}

	if err := labels.Insert(p, parts); err != nil {
		return err
	}

	return nil
}

func PreHooks(p *project.Project, e string) error {
	event := getEvent(e)
	switch event {
	case startEvent, scaleEvent:
		if err := nethooks.AutoGenLabels(p); err != nil {
			return err
		}
		if err := nethooks.AutoGenParams(p); err != nil {
			return err
		}
	}

	switch event {
	case startEvent:
		if err := nethooks.CreateNetConfig(p); err != nil {
			return err
		}
	case scaleEvent:
		if err := nethooks.ScaleNetConfig(p); err != nil {
			return err
		}
	case stopEvent:
	}

	return nil
}

func PostHooks(p *project.Project, e string) error {
	event := getEvent(e)

	switch event {
	case startEvent:
		if err := nethooks.PopulateEtcHosts(p); err != nil {
			return err
		}
	case stopEvent:
		if err := nethooks.DeleteNetConfig(p); err != nil {
			return err
		}
	}

	return nil
}
