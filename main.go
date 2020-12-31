package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/packer/packer-plugin-sdk/plugin"
	builder "github.com/veertuinc/packer-builder-veertu-anka/builder/anka"
	postprocessor "github.com/veertuinc/packer-builder-veertu-anka/post-processor/anka"
)

var version = "SNAPSHOT"
var commit = ""

func main() {
	if commit == "" {
		log.Printf("packer-builder-veertu-anka version: %s", version)
	} else {
		log.Printf("packer-builder-veertu-anka version: %s+%s", version, commit)
	}
	pps := plugin.NewSet()
	pps.RegisterBuilder("vm", new(builder.Builder))
	pps.RegisterPostProcessor("registry", new(postprocessor.PostProcessor))
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
