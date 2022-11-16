package spcmutate

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	kapi "k8s.io/kubernetes/pkg/apis/core"
)

// TestPodAdmission verifies various scenarios involving pod/project/global node label selectors
func TestPodAdmission(t *testing.T) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testProject",
			Namespace: "",
		},
	}

	handler := &podMutateSPC{}
	pod := &kapi.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "testPod"},
	}

	tests := []struct {
		podNodeSelector           map[string]string
		mergedNodeSelector        map[string]string
		ignoreProjectNodeSelector bool
		admit                     bool
		testName                  string
	}{
		{
			podNodeSelector:           map[string]string{},
			mergedNodeSelector:        map[string]string{},
			ignoreProjectNodeSelector: true,
			admit:                     true,
			testName:                  "No node selectors",
		},
	}
	for _, test := range tests {
		indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
		indexer.Add(namespace)
		handler.nsLister = corev1listers.NewNamespaceLister(indexer)
		handler.nsListerSynced = func() bool { return true }

		if !test.ignoreProjectNodeSelector {
			//namespace.ObjectMeta.Annotations = map[string]string{projectv1.ProjectNodeSelector: test.projectNodeSelector}
		}
		pod.Spec = kapi.PodSpec{NodeSelector: test.podNodeSelector}

		attrs := admission.NewAttributesRecord(pod, nil, kapi.Kind("Pod").WithVersion("version"), "testProject", namespace.ObjectMeta.Name, kapi.Resource("pods").WithVersion("version"), "", admission.Create, nil, false, nil)
		err := handler.Admit(context.TODO(), attrs, nil)
		if test.admit && err != nil {
			t.Errorf("Test: %s, expected no error but got: %s", test.testName, err)
		} else if !test.admit && err == nil {
			t.Errorf("Test: %s, expected an error", test.testName)
		} else if err == nil {
			if err := handler.Validate(context.TODO(), attrs, nil); err != nil {
				t.Errorf("Test: %s, unexpected Validate error after Admit succeeded: %v", test.testName, err)
			}
		}
	}
}

func TestHandles(t *testing.T) {
	for op, shouldHandle := range map[admission.Operation]bool{
		admission.Create:  true,
		admission.Update:  false,
		admission.Connect: false,
		admission.Delete:  false,
	} {
		mutate, err := NewPodMutateSPC()
		if err != nil {
			t.Errorf("%v: error getting node environment: %v", op, err)
			continue
		}

		if e, a := shouldHandle, mutate.Handles(op); e != a {
			t.Errorf("%v: shouldHandle=%t, handles=%t", op, e, a)
		}
	}
}
