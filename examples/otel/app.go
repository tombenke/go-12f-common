package main

import (
	"context"

	"github.com/tombenke/go-12f-common/apprun"
)

type Application struct {
	config *Config
}

func (a *Application) Components(ctx context.Context) []apprun.ComponentLifecycleManager {
	return nil
}

func NewApplication(config *Config) (apprun.Application, error) {
	return &Application{config: config}, nil
}
