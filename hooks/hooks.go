package hooks

import (
	"github.com/contiv/deploy/labels"
	"github.com/contiv/deploy/nethooks"
	"github.com/docker/libcompose/project"
)

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

func NetHooks(p *project.Project, e project.EventType) error {
	switch e {
	case project.EventProjectUpStart:
		if err := nethooks.AutoGenLabels(p); err != nil {
			return err
		}

		if err := nethooks.CreateNetConfig(p); err != nil {
			return err
		}
	case project.EventProjectDownStart:
		if err := nethooks.DeleteNetConfig(p); err != nil {
			return err
		}
	}

	return nil
}
