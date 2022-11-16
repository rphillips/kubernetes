package spcmutate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/initializer"
	"k8s.io/client-go/informers"
	corev1listers "k8s.io/client-go/listers/core/v1"
	coreapi "k8s.io/kubernetes/pkg/apis/core"
)

const (
	kubeMutateSPC          = "node.openshift.io/mutate-spc"
	timeToWaitForCacheSync = 10 * time.Second
)

func Register(plugins *admission.Plugins) {
	plugins.Register(kubeMutateSPC,
		func(config io.Reader) (admission.Interface, error) {
			return NewPodMutateSPC()
		})
}

// podMutateSPC is an implementation of admission.MutationInterface.
type podMutateSPC struct {
	*admission.Handler
	nsLister       corev1listers.NamespaceLister
	nsListerSynced func() bool
	// TODO this should become a piece of config passed to the admission plugin
	defaultNodeSelector string
}

var _ = initializer.WantsExternalKubeInformerFactory(&podMutateSPC{})
var _ = WantsDefaultNodeSelector(&podMutateSPC{})
var _ = admission.ValidationInterface(&podMutateSPC{})
var _ = admission.MutationInterface(&podMutateSPC{})

// Admit enforces that pod and its project node label selectors matches at least a node in the cluster.
func (p *podMutateSPC) admit(ctx context.Context, a admission.Attributes, mutationAllowed bool) (err error) {
	resource := a.GetResource().GroupResource()
	if resource != corev1.Resource("pods") {
		return nil
	}
	if a.GetSubresource() != "" {
		// only run the checks below on pods proper and not subresources
		return nil
	}

	obj := a.GetObject()
	pod, ok := obj.(*coreapi.Pod)
	if !ok {
		return nil
	}

	name := pod.Name

	if !p.waitForSyncedStore(time.After(timeToWaitForCacheSync)) {
		return admission.NewForbidden(a, errors.New(fmt.Sprintf("%v: caches not synchronized", kubeMutateSPC)))
	}
	namespace, err := p.nsLister.Get(a.GetNamespace())
	if err != nil {
		return apierrors.NewForbidden(resource, name, err)
	}

	// Validate annotation and skip if it doesn't exist
	if _, ok := namespace.ObjectMeta.Annotations[kubeMutateSPC]; ok {
		return nil
	}

	// mutate pod
	//pod.Spec.Label

	return nil
}

func (p *podMutateSPC) Admit(ctx context.Context, a admission.Attributes, _ admission.ObjectInterfaces) (err error) {
	return p.admit(ctx, a, true)
}

func (p *podMutateSPC) Validate(ctx context.Context, a admission.Attributes, _ admission.ObjectInterfaces) (err error) {
	return p.admit(ctx, a, false)
}

func (p *podMutateSPC) SetExternalKubeInformerFactory(kubeInformers informers.SharedInformerFactory) {
	p.nsLister = kubeInformers.Core().V1().Namespaces().Lister()
	p.nsListerSynced = kubeInformers.Core().V1().Namespaces().Informer().HasSynced
}

func (p *podMutateSPC) waitForSyncedStore(timeout <-chan time.Time) bool {
	for !p.nsListerSynced() {
		select {
		case <-time.After(100 * time.Millisecond):
		case <-timeout:
			return p.nsListerSynced()
		}
	}

	return true
}

func (p *podMutateSPC) ValidateInitialization() error {
	if p.nsLister == nil {
		return fmt.Errorf("mutate spc plugin needs a namespace lister")
	}
	if p.nsListerSynced == nil {
		return fmt.Errorf("mutate spc plugin needs a namespace lister synced")
	}
	return nil
}

func NewPodMutateSPC() (admission.Interface, error) {
	return &podMutateSPC{
		Handler: admission.NewHandler(admission.Create),
	}, nil
}
