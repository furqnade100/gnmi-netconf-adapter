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
	"reflect"
	"strings"
	"time"

	"github.com/openconfig/gnmi/value"

	"github.com/onosproject/gnmi-netconf-adapter/pkg/adapter/modeldata/gostruct"
	"github.com/openconfig/goyang/pkg/yang"

	"github.com/damianoneill/net/v2/netconf/ops"

	log "github.com/golang/glog"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Get implements the Get RPC in gNMI spec.
func (s *Adapter) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {

	dataType := req.GetType()

	if err := s.checkEncodingAndModel(req.GetEncoding(), req.GetUseModels()); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	prefix := req.GetPrefix()
	paths := req.GetPath()
	notifications := make([]*pb.Notification, len(paths))

	s.mu.RLock()
	defer s.mu.RUnlock()

	//if paths == nil && dataType.String() != "" {
	//
	//	jsonType := "IETF"
	//	if req.GetEncoding() == pb.Encoding_JSON {
	//		jsonType = "Internal"
	//	}
	//	notifications := make([]*pb.Notification, 1)
	//	path := pb.Path{}
	//	// Gets the whole config data tree
	//	node, stat := ygotutils.GetNode(s.model.schemaTreeRoot, s.config, &path)
	//	if isNil(node) || stat.GetCode() != int32(cpb.Code_OK) {
	//		return nil, status.Errorf(codes.NotFound, "path %v not found", path)
	//	}
	//
	//	nodeStruct, _ := node.(ygot.GoStruct)
	//	jsonTree, _ := ygot.ConstructIETFJSON(nodeStruct, &ygot.RFC7951JSONConfig{AppendModuleName: true})
	//
	//	jsonTree = pruneConfigData(jsonTree, strings.ToLower(dataType.String()), &path).(map[string]interface{})
	//	jsonDump, err := json.Marshal(jsonTree)
	//
	//	if err != nil {
	//		msg := fmt.Sprintf("error in marshaling %s JSON tree to bytes: %v", jsonType, err)
	//		log.Error(msg)
	//		return nil, status.Error(codes.Internal, msg)
	//	}
	//	ts := time.Now().UnixNano()
	//
	//	update := buildUpdate(jsonDump, &path, jsonType)
	//	notifications[0] = &pb.Notification{
	//		Timestamp: ts,
	//		Prefix:    prefix,
	//		Update:    []*pb.Update{update},
	//	}
	//	resp := &pb.GetResponse{Notification: notifications}
	//	return resp, nil
	//}

	for i, path := range paths {
		// Get schema node for path from config struct.
		fullPath := path
		if prefix != nil {
			fullPath = gnmiFullPath(prefix, path)
		}

		if fullPath.GetElem() == nil && fullPath.GetElement() != nil {
			return nil, status.Error(codes.Unimplemented, "deprecated path element type is unsupported")
		}

		entry := getSchemaEntryForPath(fullPath)
		if entry == nil {
			return nil, status.Errorf(codes.NotFound, "path %v not found (Test)", fullPath)
		}

		filter := getSubtreeFilterForPath(fullPath)

		result := ""
		err := s.ncs.GetConfigSubtree(filter, ops.CandidateCfg, &result)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "Failed to get config for %v %v", fullPath, err)
		}
		ts := time.Now().UnixNano()

		netconfMap := netconfToJson(result)

		//=====================

		jsonTree, err := getTarget(netconfMap, fullPath)
		if err != nil {
			return nil, err
		}

		if entry.IsLeaf() {
			var err error
			var val *pb.TypedValue
			val, err = value.FromScalar(reflect.ValueOf(&jsonTree).Elem().Interface())
			if err != nil {
				msg := fmt.Sprintf("leaf node %v does not contain a scalar type value: %v", path, err)
				log.Error(msg)
				return nil, status.Error(codes.Internal, msg)
			}
			update := &pb.Update{Path: path, Val: val}
			notifications[i] = &pb.Notification{
				Timestamp: ts,
				Prefix:    prefix,
				Update:    []*pb.Update{update},
			}
			continue
		}
		if entry.IsContainer() {

			dataTypeString := strings.ToLower(dataType.String())
			jsonTree = pruneConfigData(jsonTree, strings.ToLower(dataTypeString), fullPath).(map[string]interface{})
			if err != nil {
				msg := fmt.Sprintf("error in constructing %s JSON tree from requested node: %v", "Internal", err)
				log.Error(msg)
				return nil, status.Error(codes.Internal, msg)
			}

			jsonDump, err := json.Marshal(jsonTree)
			if err != nil {
				msg := fmt.Sprintf("error in marshaling %s JSON tree to bytes: %v", "Internal", err)
				log.Error(msg)
				return nil, status.Error(codes.Internal, msg)
			}

			update := buildUpdate(jsonDump, path, "Internal")
			notifications[i] = &pb.Notification{
				Timestamp: ts,
				Prefix:    prefix,
				Update:    []*pb.Update{update},
			}
			continue
		}
	}
	resp := &pb.GetResponse{Notification: notifications}

	return resp, nil
}

func getTarget(mapin map[string]interface{}, path *pb.Path) (interface{}, error) {
	var value interface{} = mapin
	for i, elem := range path.Elem {
		ok := false
		nextmap := value.(map[string]interface{})
		value, ok = nextmap[elem.Name]
		if !ok {
			return nil, status.Errorf(codes.NotFound, "failed to find path: %v", path)
		}
		if i == len(path.Elem)-1 {
			break
		}

	}
	return value, nil
}

func getSchemaEntryForPath(path *pb.Path) *yang.Entry {
	rootEntry := gostruct.SchemaTree["Device"]
	if path.Elem == nil {
		return rootEntry
	}

	entry := rootEntry
	for _, elem := range path.Elem {
		next, ok := entry.Dir[elem.Name]
		if !ok {
			return nil
		}
		entry = next
	}
	return entry
}

func getSubtreeFilterForPath(path *pb.Path) interface{} {

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

	for i := len(path.Elem) - 1; i >= 0; i-- {
		_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: path.Elem[i].Name}})
	}

	_ = enc.Flush()
	filter := b2.String()
	if len(filter) == 0 {
		return nil
	}
	return filter
}

func netconfToJson(result string) map[string]interface{} {
	dec := xml.NewDecoder(strings.NewReader(result))

	type eldesc struct {
		schema   *yang.Entry
		tag      string
		value    string
		children map[string]interface{}
	}
	top := make(map[string]interface{})
	stack := []*eldesc{}
	var cureld *eldesc

	schema, _ := gostruct.SchemaTree["Device"]

	for {
		tk, _ := dec.Token()
		if tk != nil {
			switch v := tk.(type) {
			case xml.StartElement:
				var nschema *yang.Entry
				if cureld == nil {

					nschema = getChildSchema(v.Name.Local, schema)

				} else {
					stack = append(stack, cureld)
					if cureld.schema != nil {
						nschema = getChildSchema(v.Name.Local, cureld.schema)
					}
				}
				cureld = &eldesc{schema: nschema, tag: v.Name.Local, children: make(map[string]interface{})}
			case xml.EndElement:
				l := len(stack)
				if l > 0 {
					preveld := cureld
					cureld = stack[l-1]
					stack = stack[:l-1]

					if preveld.schema == nil {
						break
					}
					isList := preveld.schema.IsList()

					var value interface{}
					if len(preveld.children) > 0 {
						value = preveld.children
					} else {
						value = preveld.value
					}
					if isList {
						if _, ok := cureld.children[preveld.tag]; !ok {
							cureld.children[preveld.tag] = []interface{}{}
						}
						cureld.children[preveld.tag] = append(cureld.children[preveld.tag].([]interface{}), value)
					} else {
						cureld.children[preveld.tag] = value
					}
				} else {
					top[cureld.tag] = cureld.children
				}

			case xml.CharData:
				if cureld != nil {
					// TODO SJ Handle data according to leaf type.
					cureld.value = strings.TrimSpace(string(v))
				}
			default:
				fmt.Println("Got token", tk, reflect.TypeOf(tk))
			}
		} else {
			break
		}
	}

	return top
}

func getChildSchema(name string, parent *yang.Entry) *yang.Entry {
	var nschema *yang.Entry
	ok := false
	if nschema, ok = parent.Dir[name]; !ok {
		// fmt.Println("Failed to find schema for ", v.Name.Local)
	}
	return nschema
}
