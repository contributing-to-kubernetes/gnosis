/*
  Entrypoint into our command-line app.

  Inspired by k/k/cmd/kube-apiserver/apiserver.go
*/
package main

import (
	"os"

	"github.com/contributing-to-kubernetes/gnosis/stories/kk-pr-85968/example-cobra/cmd/app"
)

func main() {
	command := app.NewAPIServerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
