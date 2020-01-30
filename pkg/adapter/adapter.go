// Copyright 2019-present Open Networking Foundation.
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

// Package gnmi implements a gnmi server to mock a device with YANG models.
package gnmi

import (
	"github.com/damianoneill/net/v2/netconf/ops"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// NewAdapter creates an instance of Adapter with given json config.
func NewAdapter(model *Model, ncs ops.OpSession) (*Adapter, error) {

	s := &Adapter{
		model: model,
		ncs:   ncs,
	}

	// Initialize readOnlyUpdateValue variable

	val := &pb.TypedValue{
		Value: &pb.TypedValue_StringVal{
			StringVal: "INIT_STATE",
		},
	}
	s.readOnlyUpdateValue = &pb.Update{Path: nil, Val: val}
	s.subscribers = make(map[string]*streamClient)
	s.ConfigUpdate = make(chan *pb.Update, 100)

	return s, nil
}
