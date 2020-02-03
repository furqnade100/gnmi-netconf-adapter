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

// doDelete deletes the path from the json tree if the path exists. If success,
// it calls the callback function to apply the change to the device hardware.
func (s *Adapter) doDelete(prefix, path *pb.Path) (*pb.UpdateResult, error) {
	//// Update json tree of the device config
	//var curNode interface{} = jsonTree
	//pathDeleted := false
	//fullPath := gnmiFullPath(prefix, path)
	//schema := s.model.schemaTreeRoot
	//for i, elem := range fullPath.Elem { // Delete sub-tree or leaf node.
	//	node, ok := curNode.(map[string]interface{})
	//	if !ok {
	//		break
	//	}
	//
	//	// Delete node
	//	if i == len(fullPath.Elem)-1 {
	//		if elem.GetKey() == nil {
	//			delete(node, elem.Name)
	//			pathDeleted = true
	//			break
	//		}
	//		pathDeleted = deleteKeyedListEntry(node, elem)
	//		break
	//	}
	//
	//	if curNode, schema = getChildNode(node, schema, elem, false); curNode == nil {
	//		break
	//	}
	//}
	//if reflect.DeepEqual(fullPath, pbRootPath) { // Delete root
	//	for k := range jsonTree {
	//		delete(jsonTree, k)
	//	}
	//}
	//
	//// Apply the validated operation to the config tree and device.
	//if pathDeleted {
	//	newConfig, err := s.toGoStruct(jsonTree)
	//	if err != nil {
	//		return nil, status.Error(codes.Internal, err.Error())
	//	}
	//	if s.callback != nil {
	//		if applyErr := s.callback(newConfig); applyErr != nil {
	//			if rollbackErr := s.callback(s.config); rollbackErr != nil {
	//				return nil, status.Errorf(codes.Internal, "error in rollback the failed operation (%v): %v", applyErr, rollbackErr)
	//			}
	//			return nil, status.Errorf(codes.Aborted, "error in applying operation to device: %v", applyErr)
	//		}
	//	}
	//}
	//return &pb.UpdateResult{
	//	Path: path,
	//	Op:   pb.UpdateResult_DELETE,
	//}, nil
	return nil, nil
}

// doReplaceOrUpdate validates the replace or update operation to be applied to
// the device, modifies the json tree of the config struct, then calls the
// callback function to apply the operation to the device hardware.
func (s *Adapter) doReplaceOrUpdate(op pb.UpdateResult_Operation, prefix, path *pb.Path, val *pb.TypedValue) (*pb.UpdateResult, error) {

	fullPath := path
	if prefix != nil {
		fullPath = gnmiFullPath(prefix, path)
	}

	entry := getSchemaEntryForPath(fullPath)
	if entry == nil {
		return nil, status.Errorf(codes.NotFound, "path %v not found (Test)", fullPath)
	}

	request, _ := editConfigRequest(fullPath, val, entry)
	fmt.Println("request", request)

	err := s.ncs.EditConfig(ops.CandidateCfg, ops.Cfg(request))
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "edit failed %s", err)
	}
	fmt.Println("edit-config ok")

	//// Validate the operation.
	//fullPath := gnmiFullPath(prefix, path)
	//emptyNode, stat := ygotutils.NewNode(s.model.structRootType, fullPath)
	//if stat.GetCode() != int32(cpb.Code_OK) {
	//	return nil, status.Errorf(codes.NotFound, "path %v is not found in the config structure: %v", fullPath, stat)
	//}
	//var nodeVal interface{}
	//nodeStruct, ok := emptyNode.(ygot.ValidatedGoStruct)
	//if ok {
	//	if err := s.model.jsonUnmarshaler(val.GetJsonIetfVal(), nodeStruct); err != nil {
	//		return nil, status.Errorf(codes.InvalidArgument, "unmarshaling json data to config struct fails: %v", err)
	//	}
	//	if err := nodeStruct.Validate(); err != nil {
	//		return nil, status.Errorf(codes.InvalidArgument, "config data validation fails: %v", err)
	//	}
	//	var err error
	//	if nodeVal, err = ygot.ConstructIETFJSON(nodeStruct, &ygot.RFC7951JSONConfig{}); err != nil {
	//		msg := fmt.Sprintf("error in constructing IETF JSON tree from config struct: %v", err)
	//		log.Error(msg)
	//		return nil, status.Error(codes.Internal, msg)
	//	}
	//} else {
	//	var err error
	//	if nodeVal, err = value.ToScalar(val); err != nil {
	//		return nil, status.Errorf(codes.Internal, "cannot convert leaf node to scalar type: %v", err)
	//	}
	//}
	//
	//// Update json tree of the device config.
	//var curNode interface{} = jsonTree
	//schema := s.model.schemaTreeRoot
	//for i, elem := range fullPath.Elem {
	//	switch node := curNode.(type) {
	//	case map[string]interface{}:
	//		// Set node value.
	//		if i == len(fullPath.Elem)-1 {
	//			if elem.GetKey() == nil {
	//				if grpcStatusError := setPathWithoutAttribute(op, node, elem, nodeVal); grpcStatusError != nil {
	//					return nil, grpcStatusError
	//				}
	//				break
	//			}
	//			if grpcStatusError := setPathWithAttribute(op, node, elem, nodeVal); grpcStatusError != nil {
	//				return nil, grpcStatusError
	//			}
	//			break
	//		}
	//
	//		if curNode, schema = getChildNode(node, schema, elem, true); curNode == nil {
	//			return nil, status.Errorf(codes.NotFound, "path elem not found: %v", elem)
	//		}
	//	case []interface{}:
	//		return nil, status.Errorf(codes.NotFound, "uncompatible path elem: %v", elem)
	//	default:
	//		return nil, status.Errorf(codes.Internal, "wrong node type: %T", curNode)
	//	}
	//}
	//if reflect.DeepEqual(fullPath, pbRootPath) { // Replace/Update root.
	//	if op == pb.UpdateResult_UPDATE {
	//		return nil, status.Error(codes.Unimplemented, "update the root of config tree is unsupported")
	//	}
	//	nodeValAsTree, ok := nodeVal.(map[string]interface{})
	//	if !ok {
	//		return nil, status.Errorf(codes.InvalidArgument, "expect a tree to replace the root, got a scalar value: %T", nodeVal)
	//	}
	//	for k := range jsonTree {
	//		delete(jsonTree, k)
	//	}
	//	for k, v := range nodeValAsTree {
	//		jsonTree[k] = v
	//	}
	//}
	//newConfig, err := s.toGoStruct(jsonTree)
	//if err != nil {
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//// Apply the validated operation to the device.
	//if s.callback != nil {
	//	if applyErr := s.callback(newConfig); applyErr != nil {
	//		if rollbackErr := s.callback(s.config); rollbackErr != nil {
	//			return nil, status.Errorf(codes.Internal, "error in rollback the failed operation (%v): %v", applyErr, rollbackErr)
	//		}
	//		return nil, status.Errorf(codes.Aborted, "error in applying operation to device: %v", applyErr)
	//	}
	//}
	//return &pb.UpdateResult{
	//	Path: path,
	//	Op:   op,
	//}, nil
	return nil, nil
}

// Set implements the Set RPC in gNMI spec.
func (s *Adapter) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {

	prefix := req.GetPrefix()
	var results []*pb.UpdateResult

	for _, path := range req.GetDelete() {
		res, grpcStatusError := s.doDelete(prefix, path)
		if grpcStatusError != nil {
			return nil, grpcStatusError
		}
		results = append(results, res)
	}
	for _, upd := range req.GetReplace() {
		res, grpcStatusError := s.doReplaceOrUpdate(pb.UpdateResult_REPLACE, prefix, upd.GetPath(), upd.GetVal())
		if grpcStatusError != nil {
			return nil, grpcStatusError
		}
		results = append(results, res)
	}
	for _, upd := range req.GetUpdate() {
		res, grpcStatusError := s.doReplaceOrUpdate(pb.UpdateResult_UPDATE, prefix, upd.GetPath(), upd.GetVal())
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
func editConfigRequest(path *pb.Path, inval *pb.TypedValue, entry *yang.Entry) (interface{}, error) {

	var editValue interface{}
	if entry.IsDir() {
		editValue = make(map[string]interface{})
		err := json.Unmarshal(inval.GetJsonIetfVal(), &editValue)
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

	var b2 bytes.Buffer
	enc := xml.NewEncoder(&b2)
	for _, elem := range path.Elem {
		_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: elem.Name}})
		for k, v := range elem.Key {
			_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: k}})
			_ = enc.EncodeToken(xml.CharData(v))
			_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: k}})
		}
	}

	if entry.IsDir() {
		_ = jsonToXml(editValue.(map[string]interface{}), enc)
	} else {
		_ = enc.EncodeToken(xml.CharData(fmt.Sprintf("%v", editValue)))
	}

	for i := len(path.Elem) - 1; i >= 0; i-- {
		_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: path.Elem[i].Name}})
	}

	_ = enc.Flush()
	filter := b2.String()
	if len(filter) == 0 {
		return nil, nil
	}
	return filter, nil
}

func jsonToXml(input map[string]interface{}, enc *xml.Encoder) error {
	for k, v := range input {
		err := enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: k}})
		if err != nil {
			return err
		}
		switch val := v.(type) {
		case map[string]interface{}:
			err = jsonToXml(val, enc)
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
