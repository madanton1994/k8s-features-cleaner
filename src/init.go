package main

import (
	"log"

	"github.com/caarlos0/env"
)

type config struct {
	Pids		[]string	`env:"PIDS" envSeparator:":" envDefault:"0"`
	GitlabToken	  string	`env:"GITLAB_TOKEN"`
	Namespace	  string	`env:"NAMESPASE" envDefault:"features"`
	GitUrl		  string	`env:"GIT_URL" envDefault:"https://gitlab.com/api/v4"`
}

func getEnvVars () *config{
	cfg := config{}
	err:= env.Parse(&cfg)
	if err != nil {
		log.Fatal("Unable to parse envs: ", err)
	}
	return &cfg
}