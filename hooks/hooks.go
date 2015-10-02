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

func NetHooks(p *project.Project, e project.Event) error {
	switch e {
	case project.PROJECT_UP_START:
		if err := nethooks.AutoGenLabels(p); err != nil {
			return err
		}

		if err := nethooks.CreateNetConfig(p); err != nil {
			return err
		}
	case project.PROJECT_DOWN_START:
		if err := nethooks.DeleteNetConfig(p); err != nil {
			return err
		}
	}

	return nil
}
