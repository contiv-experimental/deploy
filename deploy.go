package main

import (
	"flag"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/deploy/hooks"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
)

type cliOpts struct {
	composeFile string
	policyFile  string
	projectName string
	labels      string
	debug       bool
}

func main() {
	var opts cliOpts
	var flagSet *flag.FlagSet

	flagSet = flag.NewFlagSet("netd", flag.ExitOnError)
	flagSet.StringVar(&opts.composeFile,
		"file",
		"docker-compose.yml",
		"File specifying application composition")
	flagSet.StringVar(&opts.projectName,
		"project",
		"example",
		"Keyword specifying project name")
	flagSet.StringVar(&opts.policyFile,
		"policyFile",
		"docker-policy.yml",
		"File specifying application deployment policy")
	flagSet.StringVar(&opts.labels,
		"labels",
		"",
		"Comma separated list of labels associated with this launch, e.g. \"io.contiv.epg:web1,io.contiv.env:prod\"")
	flagSet.BoolVar(&opts.debug,
		"debug",
		true,
		"Enable debugging")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to parse command. Error: %s", err)
	}

	if opts.debug {
		log.SetLevel(log.DebugLevel)
	}

	p, err := docker.NewProject(&docker.Context{
		Context: project.Context{
			ComposeFile: opts.composeFile,
			ProjectName: opts.projectName,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Deploying composition '%s'\n", p.Name)

	if opts.labels != "" {
		if err := hooks.PopulateEnvLabels(p, opts.labels); err != nil {
			log.Fatalf("Unable to insert environment labels. Error %v", err)
		}
	}

	if err := hooks.NetHooks(p, project.EventProjectUpStart); err != nil {
		log.Fatalf("Unable to generate network labels. Error %v", err)
	}

	p.Up()
}
