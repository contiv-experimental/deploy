package nethooks

import (
	"testing"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
)

func TestAutoGenLabel(t *testing.T) {

	p, err := docker.NewProject(&docker.Context{
		Context: project.Context{
			ComposeFile: "./docker-compose.yml",
			ProjectName: "example",
		},
	})

	if err != nil {
		t.Fatalf("Unable to create a project. Error %v\n", err)
	}

	labelCount := map[string]int{}
	for svcName, svc := range p.Configs {
		labelCount[svcName] = len(svc.Labels.MapParts())
	}

	if err := AutoGenLabels(p); err != nil {
		t.Fatalf("Unable to auto insert labels to a project. Error %v\n", err)
	}

	for svcName, svc := range p.Configs {
		if labelCount[svcName] == len(svc.Labels.MapParts()) {
			t.Fatalf("service '%s' did not insert any labels", svcName)
		}
	}
}
