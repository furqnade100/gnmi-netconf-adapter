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
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/openconfig/goyang/pkg/yang"

	"github.com/damianoneill/net/v2/netconf/ops"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmi/value"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *Adapter) executeRequest(op pb.UpdateResult_Operation, prefix, path *pb.Path, val *pb.TypedValue) (*pb.UpdateResult, error) {

	request, err := a.mapRequest(op, prefix, path, val)
	if err != nil {
		return nil, err
	}

	err = a.ncs.EditConfig(ops.CandidateCfg, ops.Cfg(request))
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "edit failed %s", err)
	}

	return &pb.UpdateResult{
		Path: path,
		Op:   op,
	}, nil
}

// Set implements the Set RPC in gNMI spec.
func (a *Adapter) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {

	prefix := req.GetPrefix()
	var results []*pb.UpdateResult

	for _, path := range req.GetDelete() {
		res, grpcStatusError := a.executeRequest(pb.UpdateResult_DELETE, prefix, path, nil)
		if grpcStatusError != nil {
			return nil, grpcStatusError
		}
		results = append(results, res)
	}
	for _, upd := range req.GetReplace() {
		res, grpcStatusError := a.executeRequest(pb.UpdateResult_REPLACE, prefix, upd.GetPath(), upd.GetVal())
		if grpcStatusError != nil {
			return nil, grpcStatusError
		}
		results = append(results, res)
	}
	for _, upd := range req.GetUpdate() {
		res, grpcStatusError := a.executeRequest(pb.UpdateResult_UPDATE, prefix, upd.GetPath(), upd.GetVal())
		if grpcStatusError != nil {
			return nil, grpcStatusError
		}
		results = append(results, res)
	}

	setResponse := &pb.SetResponse{
		Prefix:   req.GetPrefix(),
		Response: results,
	}

	return setResponse, nil
}

func (a *Adapter) mapRequest(op pb.UpdateResult_Operation, prefix, path *pb.Path, inval *pb.TypedValue) (interface{}, error) {

	fullPath := path
	if prefix != nil {
		fullPath = gnmiFullPath(prefix, path)
	}

	var b2 bytes.Buffer
	enc := xml.NewEncoder(&b2)
	for i, elem := range fullPath.Elem {
		startEl := xml.StartElement{Name: xml.Name{Local: elem.Name}}
		if i == len(fullPath.Elem)-1 {
			startEl.Attr = []xml.Attr{{Name: xml.Name{Local: "operation"}, Value: mapOperation(op)}}
		}
		_ = enc.EncodeToken(startEl)
		for k, v := range elem.Key {
			_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: k}})
			_ = enc.EncodeToken(xml.CharData(v))
			_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: k}})
		}
	}

	if op != pb.UpdateResult_DELETE {
		entry := a.getSchemaEntryForPath(fullPath)
		if entry == nil {
			return nil, status.Errorf(codes.NotFound, "path %v not found (Test)", fullPath)
		}

		editValue, err := mapValue(entry, inval)
		if err != nil {
			return nil, err
		}
		if entry.IsDir() {
			_ = jsonToXML(editValue.(map[string]interface{}), enc)
		} else {
			_ = enc.EncodeToken(xml.CharData(fmt.Sprintf("%v", editValue)))
		}
	}

	for i := len(fullPath.Elem) - 1; i >= 0; i-- {
		_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: fullPath.Elem[i].Name}})
	}

	_ = enc.Flush()
	filter := b2.String()
	if len(filter) == 0 {
		return nil, nil
	}
	return filter, nil
}

func mapValue(entry *yang.Entry, inval *pb.TypedValue) (interface{}, error) {
	var editValue interface{}
	if entry.IsDir() {
		editValue = make(map[string]interface{})
		err := json.Unmarshal(inval.GetJsonVal(), &editValue)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "invalid value %s", err)
		}
	} else {
		var err error
		editValue, err = value.ToScalar(inval)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "invalid value %s", err)
		}
	}
	return editValue, nil
}

func mapOperation(op pb.UpdateResult_Operation) string {
	opdesc := ""
	switch op {
	case pb.UpdateResult_REPLACE:
		opdesc = "replace"
	case pb.UpdateResult_UPDATE:
		opdesc = "merge"
	case pb.UpdateResult_DELETE:
		opdesc = "delete"
	default:
		panic(fmt.Sprintf("unexpected operation %s", op))
	}
	return opdesc
}

func jsonToXML(input map[string]interface{}, enc *xml.Encoder) error {
	for k, v := range input {
		err := enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: k}})
		if err != nil {
			return err
		}
		switch val := v.(type) {
		case map[string]interface{}:
			err = jsonToXML(val, enc)
			if err != nil {
				return err
			}
		default:
			_ = enc.EncodeToken(xml.CharData(fmt.Sprintf("%v", val)))
		}
		err = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: k}})
		if err != nil {
			return err
		}
	}
	return nil
}
