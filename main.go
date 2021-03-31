package main

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	core_v1 "gitlab.eng.vmware.com/landerr/k8s-object-code-generator/core/v1"
)

func main() {

	// read manifest file
	manifestFile, _ := filepath.Abs("sample-ns.yaml")
	yamlFile, err := ioutil.ReadFile(manifestFile)
	if err != nil {
		panic(err)
	}

	// determine group version kind of resource in manifest
	metaFactory := serializer.DefaultMetaFactory

	gvk, err := metaFactory.Interpret(yamlFile)
	if err != nil {
		panic(err)
	}

	// call code generation func for the GVK if supported
	switch gvk.String() {
	case "/v1, Kind=Namespace":
		if err = core_v1.GenNamespace(yamlFile); err != nil {
			panic(err)
		}
	default:
		errors.New("Unsupported resource kind")
		panic(err)
	}
}
