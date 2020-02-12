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
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// getGNMIServiceVersion returns a pointer to the gNMI service version string.
// The method is non-trivial because of the way it is defined in the proto file.
func getGNMIServiceVersion() (*string, error) {
	gzB, _ := (&pb.Update{}).Descriptor()
	r, err := gzip.NewReader(bytes.NewReader(gzB))
	if err != nil {
		return nil, fmt.Errorf("error in initializing gzip reader: %v", err)
	}
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error in reading gzip data: %v", err)
	}
	desc := &dpb.FileDescriptorProto{}
	if err := proto.Unmarshal(b, desc); err != nil {
		return nil, fmt.Errorf("error in unmarshaling proto: %v", err)
	}
	ver, err := proto.GetExtension(desc.Options, pb.E_GnmiService)
	if err != nil {
		return nil, fmt.Errorf("error in getting version from proto extension: %v", err)
	}
	return ver.(*string), nil
}

// gnmiFullPath builds the full path from the prefix and path.
func gnmiFullPath(prefix, path *pb.Path) *pb.Path {
	fullPath := &pb.Path{Origin: path.Origin}
	if path.GetElement() != nil {
		fullPath.Element = append(prefix.GetElement(), path.GetElement()...)
	}
	if path.GetElem() != nil {
		fullPath.Elem = append(prefix.GetElem(), path.GetElem()...)
	}
	return fullPath
}

// checkEncodingAndModel checks whether encoding and models are supported by the server. Return error if anything is unsupported.
func (a *Adapter) checkEncodingAndModel(encoding pb.Encoding, models []*pb.ModelData) error {
	hasSupportedEncoding := false
	for _, supportedEncoding := range supportedEncodings {
		if encoding == supportedEncoding {
			hasSupportedEncoding = true
			break
		}
	}
	if !hasSupportedEncoding {
		return fmt.Errorf("unsupported encoding: %s", pb.Encoding_name[int32(encoding)])
	}
	for _, m := range models {
		isSupported := false
		for _, supportedModel := range a.model.modelData {
			if reflect.DeepEqual(m, supportedModel) {
				isSupported = true
				break
			}
		}
		if !isSupported {
			return fmt.Errorf("unsupported model: %v", m)
		}
	}
	return nil
}

// Contains checks the existence of a given string in an array of strings.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// pruneConfigData prunes the given JSON subtree based on the given data type and path info.
func pruneConfigData(data interface{}, dataType string, fullPath *pb.Path) interface{} {

	if reflect.ValueOf(data).Kind() == reflect.Slice {
		d := reflect.ValueOf(data)
		tmpData := make([]interface{}, d.Len())
		returnSlice := make([]interface{}, d.Len())
		for i := 0; i < d.Len(); i++ {
			tmpData[i] = d.Index(i).Interface()
		}
		for i, v := range tmpData {
			returnSlice[i] = pruneConfigData(v, dataType, fullPath)
		}
		return returnSlice
	} else if reflect.ValueOf(data).Kind() == reflect.Map {
		d := reflect.ValueOf(data)
		tmpData := make(map[string]interface{})
		for _, k := range d.MapKeys() {
			match, _ := regexp.MatchString(dataType, k.String())
			matchAll := strings.Compare(dataType, "all")
			typeOfValue := reflect.TypeOf(d.MapIndex(k).Interface()).Kind()

			if match || matchAll == 0 {
				newKey := k.String()
				if typeOfValue == reflect.Map || typeOfValue == reflect.Slice {
					tmpData[newKey] = pruneConfigData(d.MapIndex(k).Interface(), dataType, fullPath)

				} else {
					tmpData[newKey] = d.MapIndex(k).Interface()
				}
			} else {
				tmpIteration := pruneConfigData(d.MapIndex(k).Interface(), dataType, fullPath)
				if typeOfValue == reflect.Map {
					tmpMap := tmpIteration.(map[string]interface{})
					if len(tmpMap) != 0 {
						tmpData[k.String()] = tmpIteration
						if Contains(dataTypes, k.String()) {
							delete(tmpData, k.String())
						}
					}
				} else if typeOfValue == reflect.Slice {
					tmpMap := tmpIteration.([]interface{})
					if len(tmpMap) != 0 {
						tmpData[k.String()] = tmpIteration
						if Contains(dataTypes, k.String()) {
							delete(tmpData, k.String())

						}
					}
				} else {
					tmpData[k.String()] = d.MapIndex(k).Interface()

				}
			}

		}

		return tmpData
	}
	return data
}

func buildUpdate(b []byte, path *pb.Path, valType string) *pb.Update {
	var update *pb.Update

	if strings.Compare(valType, "Internal") == 0 {
		update = &pb.Update{Path: path, Val: &pb.TypedValue{Value: &pb.TypedValue_JsonVal{JsonVal: b}}}
		return update
	}
	update = &pb.Update{Path: path, Val: &pb.TypedValue{Value: &pb.TypedValue_JsonIetfVal{JsonIetfVal: b}}}

	return update
}
