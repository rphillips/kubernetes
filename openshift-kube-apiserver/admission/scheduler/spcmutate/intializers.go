package spcmutate

import (
	"k8s.io/apiserver/pkg/admission"
)

func NewInitializer() admission.PluginInitializer {
	return &localInitializer{}
}

type WantsDefaultNodeSelector interface {
	admission.InitializationValidator
}

type localInitializer struct{}

// Initialize will check the initialization interfaces implemented by each plugin
// and provide the appropriate initialization data
func (i *localInitializer) Initialize(plugin admission.Interface) {
}
