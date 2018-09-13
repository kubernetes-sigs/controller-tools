/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manager

import (
	"path/filepath"

	"sigs.k8s.io/controller-tools/pkg/scaffold/input"
)

var _ input.File = &Cmd{}

// Cmd scaffolds a manager.go to run Controllers
type Cmd struct {
	input.Input
}

// GetInput implements input.File
func (a *Cmd) GetInput() (input.Input, error) {
	if a.Path == "" {
		a.Path = filepath.Join("cmd", "manager", "main.go")
	}
	a.TemplateBody = cmdTemplate
	return a.Input, nil
}

var cmdTemplate = `{{ .Boilerplate }}

package main

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"{{ .Repo }}/pkg/apis"
	"{{ .Repo }}/pkg/controller"
)

func main() {
	// Set logger
	log.SetLogger(log.ZapLogger(true))
	var logf = log.Log.WithName("main")

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		logf.Error(err, "Get config failed")
		return
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		logf.Error(err, "Create manager failed")
		return
	}

	logf.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		logf.Error(err, "Add to scheme failed")
		return
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		logf.Error(err, "Add to manager failed")
		return
	}

	logf.Info("Starting the cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logf.Error(err, "Start cmd failed")
		return
	}
}
`
