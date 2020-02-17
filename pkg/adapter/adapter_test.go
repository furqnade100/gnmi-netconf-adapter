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
package gnmi

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/damianoneill/net/v2/netconf/ops/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/damianoneill/net/v2/netconf/ops"

	"github.com/golang/protobuf/proto"
	"github.com/openconfig/gnmi/value"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/openconfig/gnmi/proto/gnmi"

	"github.com/onosproject/gnmi-netconf-adapter/pkg/adapter/testdata/modeldata"
	"github.com/onosproject/gnmi-netconf-adapter/pkg/adapter/testdata/modeldata/gostruct"
)

var (
	// model is the model for test config server.
	model = NewModel(modeldata.ModelData, gostruct.SchemaTree["Device"])
)

func TestGet(t *testing.T) {

	jsonSystemRoot := `{
		"system": {
			"openflow": {
				"agent": {
					"config": {
						"failure-mode": "SECURE",
						"max-backoff": 10
					}
				}
			}
	  }
	}`

	tests := []struct {
		desc        string
		textPbPath  string
		modelData   []*pb.ModelData
		wantRetCode codes.Code
		wantRespVal interface{}
		ncFilter    interface{}
		ncResponse  error
		ncResult    string
	}{{
		desc: "get valid but non-existing node",
		textPbPath: `
			elem: <name: "system" >
			elem: <name: "clock" >
		`,
		ncFilter:    `<system><clock></clock></system>`,
		wantRetCode: codes.NotFound,
	}, {
		desc:        "root node",
		ncResult:    `<system><openflow><agent><config><failure-mode>SECURE</failure-mode><max-backoff>10</max-backoff></config></agent></openflow></system>`,
		wantRetCode: codes.OK,
		wantRespVal: jsonSystemRoot,
	}, {
		desc: "get non-enum type",
		textPbPath: `
					elem: <name: "system" >
					elem: <name: "openflow" >
					elem: <name: "agent" >
					elem: <name: "config" >
					elem: <name: "max-backoff" >
				`,
		ncFilter:    `<system><openflow><agent><config><max-backoff></max-backoff></config></agent></openflow></system>`,
		ncResult:    `<system><openflow><agent><config><max-backoff>10</max-backoff></config></agent></openflow></system>`,
		wantRetCode: codes.OK,
		wantRespVal: uint64(10),
	}, {
		desc: "get enum type",
		textPbPath: `
					elem: <name: "system" >
					elem: <name: "openflow" >
					elem: <name: "agent" >
					elem: <name: "config" >
					elem: <name: "failure-mode" >
				`,
		ncFilter:    `<system><openflow><agent><config><failure-mode></failure-mode></config></agent></openflow></system>`,
		ncResult:    `<system><openflow><agent><config><failure-mode>SECURE</failure-mode></config></agent></openflow></system>`,
		wantRetCode: codes.OK,
		wantRespVal: "SECURE",
	}, {
		desc:        "root child node",
		textPbPath:  `elem: <name: "components" >`,
		ncFilter:    `<components></components>`,
		ncResult:    `<components><component><name>swpri1-1-1</name><config><name>swpri1-1-1</name></config></component></components>`,
		wantRetCode: codes.OK,
		wantRespVal: `{
							"component": [{
								"config": {
						        	"name": "swpri1-1-1"
								},
						        "name": "swpri1-1-1"
							}]}`,
	}, {
		desc: "node with attribute",
		textPbPath: `
								elem: <name: "components" >
								elem: <
									name: "component"
									key: <key: "name" value: "swpri1-1-1" >
								>`,
		ncFilter:    `<components><component><name>swpri1-1-1</name></component></components>`,
		ncResult:    `<components><component><name>swpri1-1-1</name><config><name>swpri1-1-1</name></config></component></components>`,
		wantRetCode: codes.OK,
		wantRespVal: `{
								"config": {"name": "swpri1-1-1"},
								"name": "swpri1-1-1"
							}`,
	}, {
		desc: "node with attribute in its parent",
		textPbPath: `
								elem: <name: "components" >
								elem: <
									name: "component"
									key: <key: "name" value: "swpri1-1-1" >
								>
								elem: <name: "config" >`,
		ncFilter:    `<components><component><name>swpri1-1-1</name><config></config></component></components>`,
		ncResult:    `<components><component><name>swpri1-1-1</name><config><name>swpri1-1-1</name></config></component></components>`,
		wantRetCode: codes.OK,
		wantRespVal: `{"name": "swpri1-1-1"}`,
	}, {
		desc: "ref leaf node",
		textPbPath: `
								elem: <name: "components" >
								elem: <
									name: "component"
									key: <key: "name" value: "swpri1-1-1" >
								>
								elem: <name: "name" >`,
		ncFilter:    `<components><component><name>swpri1-1-1</name><name></name></component></components>`,
		ncResult:    `<components><component><name>swpri1-1-1</name><config><name>swpri1-1-1</name></config></component></components>`,
		wantRetCode: codes.OK,
		wantRespVal: "swpri1-1-1",
	}, {
		desc: "regular leaf node",
		textPbPath: `
								elem: <name: "components" >
								elem: <
									name: "component"
									key: <key: "name" value: "swpri1-1-1" >
								>
								elem: <name: "config" >
								elem: <name: "name" >`,
		ncFilter:    `<components><component><name>swpri1-1-1</name><config><name></name></config></component></components>`,
		ncResult:    `<components><component><name>swpri1-1-1</name><config><name>swpri1-1-1</name></config></component></components>`,
		wantRetCode: codes.OK,
		wantRespVal: "swpri1-1-1",
	}, {
		desc: "non-existing node: wrong path name",
		textPbPath: `
								elem: <name: "components" >
								elem: <
									name: "component"
									key: <key: "foo" value: "swpri1-1-1" >
								>
								elem: <name: "bar" >`,
		wantRetCode: codes.NotFound,
	}, {
		desc: "non-existing node: wrong path attribute",
		textPbPath: `
								elem: <name: "components" >
								elem: <
									name: "component"
									key: <key: "foo" value: "swpri2-2-2" >
								>
								elem: <name: "name" >`,
		ncFilter:    `<components><component><foo>swpri2-2-2</foo><name></name></component></components>`,
		wantRetCode: codes.NotFound,
	}, {
		desc:        "use of model data not supported",
		modelData:   []*pb.ModelData{&pb.ModelData{}},
		wantRetCode: codes.Unimplemented,
	}}

	for i := range tests {
		td := tests[i]
		t.Run(td.desc, func(t *testing.T) {
			runTestGet(t, td.textPbPath, td.wantRetCode, td.wantRespVal, td.modelData, td.ncFilter, td.ncResponse, td.ncResult)
		})
	}
}

// runTestGet requests a path from the server by Get grpc call, and compares if
// the return code and response value are expected.
func runTestGet(t *testing.T, textPbPath string, wantRetCode codes.Code, wantRespVal interface{}, useModels []*pb.ModelData,
	ncFilter interface{}, ncResponse error, ncResult string) {

	mockNc := &mocks.OpSession{}
	mockNc.On("GetConfigSubtree", ncFilter, ops.CandidateCfg, mock.Anything).Return(
		func(filter interface{}, source string, result interface{}) error {
			*result.(*string) = ncResult
			return ncResponse
		})

	s, err := NewAdapter(model, mockNc)
	if err != nil {
		t.Fatalf("error in creating server: %v", err)
	}

	// Send request
	var pbPath pb.Path
	if err := proto.UnmarshalText(textPbPath, &pbPath); err != nil {
		t.Fatalf("error in unmarshaling path: %v", err)
	}
	req := &pb.GetRequest{
		Path:      []*pb.Path{&pbPath},
		Encoding:  pb.Encoding_JSON,
		UseModels: useModels,
	}
	resp, err := s.Get(context.TODO(), req)

	// Check return code
	gotRetStatus, ok := status.FromError(err)
	if !ok {
		t.Fatal("got a non-grpc error from grpc call")
	}
	if gotRetStatus.Code() != wantRetCode {
		t.Fatalf("got return code %v, want %v", gotRetStatus.Code(), wantRetCode)
	}

	// Check response value
	var gotVal interface{}
	if resp != nil {
		notifs := resp.GetNotification()
		if len(notifs) != 1 {
			t.Fatalf("got %d notifications, want 1", len(notifs))
		}
		updates := notifs[0].GetUpdate()
		if len(updates) != 1 {
			t.Fatalf("got %d updates in the notification, want 1", len(updates))
		}
		val := updates[0].GetVal()
		if val.GetJsonVal() == nil {
			gotVal, err = value.ToScalar(val)
			if err != nil {
				t.Errorf("got: %v, want a scalar value", gotVal)
			}
		} else {
			// Unmarshal json data to gotVal container for comparison
			if err := json.Unmarshal(val.GetJsonVal(), &gotVal); err != nil {
				t.Fatalf("error in unmarshaling JSON data to json container: %v", err)
			}
			var wantJSONStruct interface{}
			if err := json.Unmarshal([]byte(wantRespVal.(string)), &wantJSONStruct); err != nil {
				t.Fatalf("error in unmarshaling IETF JSON data to json container: %v", err)
			}
			wantRespVal = wantJSONStruct
		}
	}

	if !reflect.DeepEqual(gotVal, wantRespVal) {
		t.Errorf("got: %v (%T),\nwant %v (%T)", gotVal, gotVal, wantRespVal, wantRespVal)
	}
}

type gnmiSetTestCase struct {
	desc        string                    // description of test case.
	op          pb.UpdateResult_Operation // operation type.
	textPbPath  string                    // text format of gnmi Path proto.
	val         *pb.TypedValue            // value for UPDATE/REPLACE operations. always nil for DELETE.
	wantRetCode codes.Code                // grpc return code.
	ncFilter    interface{}
	ncResponse  error
}

func TestUpdate(t *testing.T) {
	tests := []gnmiSetTestCase{{
		desc: "update leaf node",
		op:   pb.UpdateResult_UPDATE,
		textPbPath: `
			elem: <name: "system" >
			elem: <name: "config" >
			elem: <name: "domain-name" >
		`,
		val: &pb.TypedValue{
			Value: &pb.TypedValue_StringVal{StringVal: "foo.bar.com"},
		},
		ncFilter:    `<system><config><domain-name operation="merge">foo.bar.com</domain-name></config></system>`,
		wantRetCode: codes.OK,
	}, {
		desc: "update subtree",
		op:   pb.UpdateResult_UPDATE,
		textPbPath: `
			elem: <name: "system" >
			elem: <name: "config" >
		`,
		val: &pb.TypedValue{
			Value: &pb.TypedValue_JsonVal{
				JsonVal: []byte(`{"domain-name": "foo.bar.com", "hostname": "switch_a"}`),
			},
		},
		ncFilter:    `<system><config operation="merge"><domain-name>foo.bar.com</domain-name><hostname>switch_a</hostname></config></system>`,
		wantRetCode: codes.OK,
	}}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.desc, func(t *testing.T) {
			runTestSet(t, model, tc)
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []gnmiSetTestCase{{
		desc: "delete leaf node",
		op:   pb.UpdateResult_DELETE,
		textPbPath: `
			elem: <name: "system" >
			elem: <name: "config" >
			elem: <name: "login-banner" >
		`,
		ncFilter:    `<system><config><login-banner operation="delete"></login-banner></config></system>`,
		wantRetCode: codes.OK,
	}, {
		desc: "delete sub-tree",
		op:   pb.UpdateResult_DELETE,
		textPbPath: `
			elem: <name: "system" >
			elem: <name: "clock" >
		`,
		ncFilter:    `<system><clock operation="delete"></clock></system>`,
		wantRetCode: codes.OK,
	}, {
		desc: "delete root",

		op:          pb.UpdateResult_DELETE,
		wantRetCode: codes.OK,
	}, {
		desc: "delete leaf node with attribute in its precedent path",
		op:   pb.UpdateResult_DELETE,
		textPbPath: `
			elem: <name: "components" >
			elem: <
				name: "component"
				key: <key: "name" value: "swpri1-1-1" >
			>
			elem: <name: "state" >
			elem: <name: "mfg-name" >`,
		ncFilter:    `<components><component><name>swpri1-1-1</name><state><mfg-name operation="delete"></mfg-name></state></component></components>`,
		wantRetCode: codes.OK,
	}, {
		desc: "delete sub-tree with attribute in its precedent path",
		op:   pb.UpdateResult_DELETE,
		textPbPath: `
			elem: <name: "components" >
			elem: <
				name: "component"
				key: <key: "name" value: "swpri1-1-1" >
			>
			elem: <name: "state" >`,
		ncFilter:    `<components><component><name>swpri1-1-1</name><state operation="delete"></state></component></components>`,
		wantRetCode: codes.OK,
	}}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.desc, func(t *testing.T) {
			runTestSet(t, model, tc)
		})
	}
}

func TestReplace(t *testing.T) {

	//systemConfig := `{
	//	"system": {
	//		"clock": {
	//			"config": {
	//				"timezone-name": "Europe/Stockholm"
	//			}
	//		},
	//		"config": {
	//			"hostname": "switch_a",
	//			"login-banner": "Hello!"
	//		}
	//	}
	//}`

	tests := []gnmiSetTestCase{{
		//	desc: "replace root",
		//	op:   pb.UpdateResult_REPLACE,
		//	val: &pb.TypedValue{
		//		Value: &pb.TypedValue_JsonVal{
		//			JsonVal: []byte(systemConfig),
		//		}},
		//	wantRetCode: codes.OK,
		//}, {
		desc: "replace a subtree",
		op:   pb.UpdateResult_REPLACE,
		textPbPath: `
			elem: <name: "system" >
			elem: <name: "clock" >
		`,
		val: &pb.TypedValue{
			Value: &pb.TypedValue_JsonVal{
				JsonVal: []byte(`{"config": {"timezone-name": "US/New York"}}`),
			},
		},
		ncFilter:    `<system><clock operation="replace"><config><timezone-name>US/New York</timezone-name></config></clock></system>`,
		wantRetCode: codes.OK,
	}, {
		desc: "replace a keyed list subtree",
		op:   pb.UpdateResult_REPLACE,
		textPbPath: `
			elem: <name: "components" >
			elem: <
				name: "component"
				key: <key: "name" value: "swpri1-1-1" >
			>`,
		val: &pb.TypedValue{
			Value: &pb.TypedValue_JsonVal{
				JsonVal: []byte(`{"config": {"name": "swpri1-1-1"}}`),
			},
		},
		ncFilter:    `<components><component operation="replace"><name>swpri1-1-1</name><config><name>swpri1-1-1</name></config></component></components>`,
		wantRetCode: codes.OK,
	}, {
		desc: "replace node with int type attribute in its precedent path",
		op:   pb.UpdateResult_REPLACE,
		textPbPath: `
			elem: <name: "system" >
			elem: <name: "openflow" >
			elem: <name: "controllers" >
			elem: <
				name: "controller"
				key: <key: "name" value: "main" >
			>
			elem: <name: "connections" >
			elem: <
				name: "connection"
				key: <key: "aux-id" value: "0" >
			>
			elem: <name: "config" >
		`,
		val: &pb.TypedValue{
			Value: &pb.TypedValue_JsonVal{
				JsonVal: []byte(`{"address": "192.0.2.10", "aux-id": 0}`),
			},
		},
		ncFilter:    `<system><openflow><controllers><controller><name>main</name><connections><connection><aux-id>0</aux-id><config operation="replace"><address>192.0.2.10</address><aux-id>0</aux-id></config></connection></connections></controller></controllers></openflow></system>`,
		wantRetCode: codes.OK,
	}, {
		desc: "replace a leaf node of int type",
		op:   pb.UpdateResult_REPLACE,
		textPbPath: `
			elem: <name: "system" >
			elem: <name: "openflow" >
			elem: <name: "agent" >
			elem: <name: "config" >
			elem: <name: "backoff-interval" >
		`,
		val: &pb.TypedValue{
			Value: &pb.TypedValue_IntVal{IntVal: 5},
		},
		ncFilter:    `<system><openflow><agent><config><backoff-interval operation="replace">5</backoff-interval></config></agent></openflow></system>`,
		wantRetCode: codes.OK,
	}, {
		desc: "replace a leaf node of string type",
		op:   pb.UpdateResult_REPLACE,
		textPbPath: `
			elem: <name: "system" >
			elem: <name: "openflow" >
			elem: <name: "agent" >
			elem: <name: "config" >
			elem: <name: "datapath-id" >
		`,
		val: &pb.TypedValue{
			Value: &pb.TypedValue_StringVal{StringVal: "00:16:3e:00:00:00:00:00"},
		},
		ncFilter:    `<system><openflow><agent><config><datapath-id operation="replace">00:16:3e:00:00:00:00:00</datapath-id></config></agent></openflow></system>`,
		wantRetCode: codes.OK,
	}, {
		desc: "replace an non-existing leaf node",
		op:   pb.UpdateResult_REPLACE,
		textPbPath: `
			elem: <name: "system" >
			elem: <name: "openflow" >
			elem: <name: "agent" >
			elem: <name: "config" >
			elem: <name: "foo-bar" >
		`,
		val: &pb.TypedValue{
			Value: &pb.TypedValue_StringVal{StringVal: "SECURE"},
		},
		wantRetCode: codes.NotFound,
	}}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.desc, func(t *testing.T) {
			runTestSet(t, model, tc)
		})
	}
}

func runTestSet(t *testing.T, m *Model, tc gnmiSetTestCase) {

	mockNc := &mocks.OpSession{}
	mockNc.On("EditConfigCfg", ops.CandidateCfg, tc.ncFilter).Return(tc.ncResponse)

	s, err := NewAdapter(model, mockNc)
	if err != nil {
		t.Fatalf("error in creating adapter: %v", err)
	}

	// Send request
	var pbPath pb.Path
	if err := proto.UnmarshalText(tc.textPbPath, &pbPath); err != nil {
		t.Fatalf("error in unmarshaling path: %v", err)
	}
	var req *pb.SetRequest
	switch tc.op {
	case pb.UpdateResult_DELETE:
		req = &pb.SetRequest{Delete: []*pb.Path{&pbPath}}
	case pb.UpdateResult_REPLACE:
		req = &pb.SetRequest{Replace: []*pb.Update{{Path: &pbPath, Val: tc.val}}}
	case pb.UpdateResult_UPDATE:
		req = &pb.SetRequest{Update: []*pb.Update{{Path: &pbPath, Val: tc.val}}}
	default:
		t.Fatalf("invalid op type: %v", tc.op)
	}
	_, err = s.Set(context.TODO(), req)

	// Check return code
	gotRetStatus, ok := status.FromError(err)
	if !ok {
		t.Fatal("got a non-grpc error from grpc call")
	}
	if gotRetStatus.Code() != tc.wantRetCode {
		t.Fatalf("got return code %v, want %v\nerror message: %v", gotRetStatus.Code(), tc.wantRetCode, err)
	}
}

func testServer(t *testing.T) (ops.OpSession, error) {
	sshConfig := &ssh.ClientConfig{
		User:            "regress",
		Auth:            []ssh.AuthMethod{ssh.Password("MaRtInI")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	ncs, err := ops.NewSession(context.Background(), sshConfig, fmt.Sprintf("10.228.63.5:%d", 830))
	if err != nil {
		t.Fatalf("failed in creating server: %v", err)
	}
	return ncs, err
}
