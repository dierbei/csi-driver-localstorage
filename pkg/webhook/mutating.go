/*
Copyright 2021 The Caoyingjunz Authors.

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

package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	localstoragev1 "github.com/caoyingjunz/csi-driver-localstorage/pkg/apis/localstorage/v1"
	"github.com/caoyingjunz/csi-driver-localstorage/pkg/util"
)

type LocalstorageMutate struct {
	Client  client.Client
	decoder *admission.Decoder
}

var _ admission.Handler = &LocalstorageMutate{}
var _ admission.DecoderInjector = &LocalstorageMutate{}

func (s *LocalstorageMutate) Handle(ctx context.Context, req admission.Request) admission.Response {
	ls := &localstoragev1.LocalStorage{}
	if err := s.decoder.Decode(req, ls); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	klog.Infof("Mutating localstorage %s for %s", ls.Name, req.Operation)

	// AddFinalizer
	if !util.ContainsFinalizer(ls, util.LsProtectionFinalizer) {
		util.AddFinalizer(ls, util.LsProtectionFinalizer)
	}
	// Set localstorage phase when created
	if len(ls.Status.Phase) == 0 {
		ls.Status.Phase = localstoragev1.LocalStoragePending
	}

	// TODO: set the other spec

	data, err := json.Marshal(ls)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	klog.Infof("Mutated localstorage %+v for %s", ls, req.Operation)
	return admission.PatchResponseFromRaw(req.Object.Raw, data)
}

// InjectDecoder implements admission.DecoderInjector interface.
// A decoder will be automatically injected by InjectDecoderInto.
func (s *LocalstorageMutate) InjectDecoder(d *admission.Decoder) error {
	s.decoder = d
	return nil
}
