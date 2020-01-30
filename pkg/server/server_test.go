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

	"github.com/damianoneill/net/v2/netconf/ops"

	"github.com/golang/protobuf/proto"
	"github.com/openconfig/gnmi/value"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/openconfig/gnmi/proto/gnmi"

	"github.com/onosproject/gnmi-netconf-adapter/pkg/server/modeldata"
	"github.com/onosproject/gnmi-netconf-adapter/pkg/server/modeldata/gostruct"
)

var (
	// model is the model for test config server.
	model = &Model{
		modelData:       modeldata.ModelData,
		schemaTreeRoot:  gostruct.SchemaTree["Device"],
		jsonUnmarshaler: gostruct.Unmarshal,
		enumData:        gostruct.ΛEnum,
	}
)

func TestGet(t *testing.T) {

	//ts := testserver.NewTestNetconfServer(nil).WithRequestHandler(testserver.SmartRequesttHandler)
	//defer ts.Close()

	sshConfig := &ssh.ClientConfig{
		User:            "regress",
		Auth:            []ssh.AuthMethod{ssh.Password("MaRtInI")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	ncs, err := ops.NewSession(context.Background(), sshConfig, fmt.Sprintf("10.228.63.5:%d", 830))
	if err != nil {
		t.Fatalf("failed in creating server: %v", err)
	}
	defer ncs.Close()

	s, err := NewServer(model, ncs)
	if err != nil {
		t.Fatalf("error in creating server: %v", err)
	}

	tds := []struct {
		desc        string
		textPbPath  string
		modelData   []*pb.ModelData
		wantRetCode codes.Code
		wantRespVal interface{}
	}{{
		desc: "get valid but non-existing node",
		textPbPath: `
			elem: <name: "system" >
			elem: <name: "clock" >
		`,
		wantRetCode: codes.NotFound,
	}, {
		desc:        "root node",
		wantRetCode: codes.OK,
		wantRespVal: ``,
	}, {
		desc: "get leaf",
		textPbPath: `
					elem: <name: "configuration" >
					elem: <name: "system" >
					elem: <name: "services" >
					elem: <name: "ssh" >
					elem: <name: "max-sessions-per-connection" >
				`,
		wantRetCode: codes.OK,
		wantRespVal: uint64(10),
	}, {
		desc: "get container",
		textPbPath: `
					elem: <name: "configuration" >
					elem: <name: "system" >
					elem: <name: "services" >
					elem: <name: "ssh" >
				`,
		wantRetCode: codes.OK,
		wantRespVal: uint64(10),
	}, {
		desc: "get keyed container",
		textPbPath: `
					elem: <name: "configuration" >
					elem: <
						name: "groups"
						key: <key: "name" value: "re1" >
					>
	`,
		wantRetCode: codes.OK,
		wantRespVal: "SECURE",
	}, {
		desc:        "root child node",
		textPbPath:  `elem: <name: "components" >`,
		wantRetCode: codes.OK,
		wantRespVal: `{
							"openconfig-platform:component": [{
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
		wantRetCode: codes.OK,
		wantRespVal: `{
								"openconfig-platform:config": {"name": "swpri1-1-1"},
								"openconfig-platform:name": "swpri1-1-1"
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
		wantRetCode: codes.OK,
		wantRespVal: `{"openconfig-platform:name": "swpri1-1-1"}`,
	}, {
		desc: "ref leaf node",
		textPbPath: `
								elem: <name: "components" >
								elem: <
									name: "component"
									key: <key: "name" value: "swpri1-1-1" >
								>
								elem: <name: "name" >`,
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
		wantRetCode: codes.NotFound,
	}, {
		desc:        "use of model data not supported",
		modelData:   []*pb.ModelData{&pb.ModelData{}},
		wantRetCode: codes.Unimplemented,
	}}

	for _, td := range tds {
		t.Run(td.desc, func(t *testing.T) {
			runTestGet(t, s, td.textPbPath, td.wantRetCode, td.wantRespVal, td.modelData)
		})
	}
}

// runTestGet requests a path from the server by Get grpc call, and compares if
// the return code and response value are expected.
func runTestGet(t *testing.T, s *Server, textPbPath string, wantRetCode codes.Code, wantRespVal interface{}, useModels []*pb.ModelData) {
	// Send request
	var pbPath pb.Path
	if err := proto.UnmarshalText(textPbPath, &pbPath); err != nil {
		t.Fatalf("error in unmarshaling path: %v", err)
	}
	req := &pb.GetRequest{
		Path:      []*pb.Path{&pbPath},
		Encoding:  pb.Encoding_JSON_IETF,
		UseModels: useModels,
	}
	resp, err := s.Get(nil, req)

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
		if val.GetJsonIetfVal() == nil {
			gotVal, err = value.ToScalar(val)
			if err != nil {
				t.Errorf("got: %v, want a scalar value", gotVal)
			}
		} else {
			// Unmarshal json data to gotVal container for comparison
			if err := json.Unmarshal(val.GetJsonIetfVal(), &gotVal); err != nil {
				t.Fatalf("error in unmarshaling IETF JSON data to json container: %v", err)
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
