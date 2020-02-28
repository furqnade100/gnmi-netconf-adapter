// Copyright 2020-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package gnmi

import (
	"context"
	"reflect"
	"testing"

	pb "github.com/openconfig/gnmi/proto/gnmi"

	"github.com/onosproject/gnmi-netconf-adapter/pkg/adapter/modeldata"
	"github.com/onosproject/gnmi-netconf-adapter/pkg/adapter/modeldata/gostruct"
)

func TestCapabilities(t *testing.T) {
	s, _ := NewAdapter(NewModel(modeldata.ModelData, gostruct.SchemaTree["Device"]), nil)

	resp, err := s.Capabilities(context.Background(), &pb.CapabilityRequest{})
	if err != nil {
		t.Fatalf("got error %v, want nil", err)
	}
	if !reflect.DeepEqual(resp.GetSupportedModels(), model.modelData) {
		t.Errorf("got supported models %v\nare not the same as\nmodel supported by the server %v", resp.GetSupportedModels(), model.modelData)
	}
	if !reflect.DeepEqual(resp.GetSupportedEncodings(), supportedEncodings) {
		t.Errorf("got supported encodings %v\nare not the same as\nencodings supported by the server %v", resp.GetSupportedEncodings(), supportedEncodings)
	}
}
