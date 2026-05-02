/*
Copyright 2022 Gravitational, Inc.

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

package kubeserver

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"sort"

	"github.com/gravitational/trace"
	"github.com/julienschmidt/httprouter"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/kube/proxy/responsewriters"
)

var podList = corev1.PodList{
	TypeMeta: metav1.TypeMeta{
		Kind:       "PodList",
		APIVersion: "v1",
	},
	ListMeta: metav1.ListMeta{
		ResourceVersion: "1231415",
	},
	Items: []corev1.Pod{
		newPod("nginx-1", "default"),
		newPod("nginx-2", "default"),
		newPod("test", "default"),
		newPod("nginx-1", "dev"),
		newPod("nginx-2", "dev"),
	},
}

func newPod(name, namespace string) corev1.Pod {
	return corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{},
	}
}

func (s *KubeMockServer) listPods(w http.ResponseWriter, req *http.Request, p httprouter.Params) (any, error) {
	items := []corev1.Pod{}

	namespace := p.ByName("namespace")
	filter := func(pod corev1.Pod) bool {
		return len(namespace) == 0 || namespace == pod.Namespace
	}
	for _, pod := range podList.Items {
		if filter(pod) {
			items = append(items, pod)
		}
	}
	return &corev1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodList",
			APIVersion: "v1",
		},
		ListMeta: metav1.ListMeta{
			ResourceVersion: "1231415",
		},
		Items: items,
	}, nil
}

func (s *KubeMockServer) getPod(w http.ResponseWriter, req *http.Request, p httprouter.Params) (any, error) {
	if s.getPodError != nil {
		s.writeResponseError(w, nil, s.getPodError)
		return nil, nil
	}
	namespace := p.ByName("namespace")
	name := p.ByName("name")
	filter := func(pod corev1.Pod) bool {
		return pod.Name == name && namespace == pod.Namespace
	}
	for _, pod := range podList.Items {
		if filter(pod) {
			return pod, nil
		}
	}
	return nil, trace.NotFound("pod %q not found", filepath.Join(namespace, name))
}

func (s *KubeMockServer) deletePod(w http.ResponseWriter, req *http.Request, p httprouter.Params) (any, error) {
	namespace := p.ByName("namespace")
	name := p.ByName("name")
	deleteOpts, err := parseDeleteCollectionBody(req)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	reqID := ""
	if deleteOpts.Preconditions != nil && deleteOpts.Preconditions.UID != nil {
		reqID = string(*deleteOpts.Preconditions.UID)
	}
	filter := func(pod corev1.Pod) bool {
		return pod.Name == name && namespace == pod.Namespace
	}
	for _, pod := range podList.Items {
		if filter(pod) {
			s.mu.Lock()
			s.deletedResources[deletedResource{kind: types.KindKubePod, requestID: reqID}] = append(s.deletedResources[deletedResource{kind: types.KindKubePod, requestID: reqID}], filepath.Join(namespace, name))
			s.mu.Unlock()
			return pod, nil
		}
	}
	return nil, trace.NotFound("pod %q not found", filepath.Join(namespace, name))
}

func (s *KubeMockServer) DeletedPods(reqID string) []string {
	s.mu.Lock()
	key := deletedResource{kind: types.KindKubePod, requestID: reqID}
	deleted := make([]string, len(s.deletedResources[key]))
	copy(deleted, s.deletedResources[key])
	s.mu.Unlock()
	sort.Strings(deleted)
	return deleted
}

// parseDeleteCollectionBody parses the request body targeted to pod collection
// endpoints.
func parseDeleteCollectionBody(req *http.Request) (metav1.DeleteOptions, error) {
	into := metav1.DeleteOptions{}
	data, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return into, trace.Wrap(err)
	}
	if len(data) == 0 {
		return into, nil
	}

	decoder, err := newDecoderForContentType(responsewriters.GetContentTypeHeader(req.Header))
	if err != nil {
		return into, trace.Wrap(err)
	}
	obj, _, err := decoder.Decode(data, nil /* defaults */, &into)
	if err != nil {
		return into, trace.Wrap(err)
	}
	deleteOptions, ok := obj.(*metav1.DeleteOptions)
	if !ok {
		return into, trace.BadParameter("expected DeleteOptions, got %T", obj)
	}
	return *deleteOptions, nil
}

func newDecoderForContentType(contentType string) (runtime.Decoder, error) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, trace.WrapWithMessage(err, "unable to parse %q header %q", responsewriters.ContentTypeHeader, contentType)
	}
	decoder, err := newClientNegotiator().Decoder(mediaType, params)
	return decoder, trace.Wrap(err)
}

// newClientNegotiator creates a negotiator that can decode the content types
// used by Kubernetes clients, including protobuf DeleteOptions bodies.
func newClientNegotiator() runtime.ClientNegotiator {
	return runtime.NewClientNegotiator(
		kubeCodecs.WithoutConversion(),
		schema.GroupVersion{
			Version: "v1",
		},
	)
}
