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

package main

import (
	//"encoding/json"

	"strconv"

	"github.com/google/gnxi/utils/credentials"
	//dataConv "github.com/onosproject/gnmi-netconf-adapter/pkg/dataConversion"
	"fmt"
	//"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/gnmi/proto/gnmi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// type set_go ygot.GoStruct
// Name string `yang:"Name,nomerge"`

// Set overrides the Set func of gnmi.Target to provide user auth.
func (s *server) Set(ctx context.Context, req *gnmi.SetRequest) (*gnmi.SetResponse, error) {
	//checking pull behavior
	msg, ok := credentials.AuthorizeUser(ctx)
	if !ok {
		log.Infof("denied a Set request: %v", msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}
	log.Infof("allowed a Set request: %v", msg)
	//log.Infof("Allowed set req..")
	// var args *ygot.RFC7951JSONConfig
	// args.AppendModuleName = true

	// var set_go ygot.GoStruct

	// ygot.ConstructIETFJSON(req, args)

	log.Infof(req.String())
	global_counter := -1
	var xmlPath string
	for _, upd := range req.GetUpdate() {
		for i, e := range upd.GetPath().Elem {
			fmt.Println(i, e.GetName())
			fmt.Println(i, e.GetKey())
		}

		calculateXmlPath(&xmlPath, &global_counter, upd, upd.GetPath().Elem)

	}
	fmt.Println(xmlPath)

	// fmt.Println("ext number = ", len(req.GetExtension()))
	// for i, e := range req.GetExtension() {
	// 	fmt.Println(i, e.String())
	// }
	//log.Print(path)

	//dataConv.Convert(req)
	// log.Infof(req.String())
	// path := upd.GetPath()
	// fullPath := path
	// if prefix != nil {
	// 	fmt.Println("prefix exists")
	// 	fullPath = gnmiFullPath(prefix, path)
	// }
	//log.Infof(upd.getva)
	setResponse, err := s.Server.Set(ctx, req)
	return setResponse, err
	//	return nil, nil
}

func GetValue(upd *gnmi.Update) string {

	fmt.Println(upd.GetVal().String())
	bool_val := upd.GetVal().GetBoolVal()
	fmt.Println(bool_val)
	// log.Infof(string(upd.GetVal().GetJsonIetfVal()))
	// log.Infof(upd.GetVal().GetStringVal())
	// var editValue interface{}
	// editValue = make(map[string]interface{})
	// err := json.Unmarshal(upd.GetVal().GetJsonVal(), &editValue)
	// if err != nil {
	// 	status.Errorf(codes.Unknown, "invalid value %s", err)
	// }

	//return upd.GetVal().String()
	return strconv.FormatBool(bool_val)
}

func addMapValues(count int, path *string, elem []*gnmi.PathElem) {

	for key, value := range elem[count].GetKey() {
		*path += `<` + key + `>` + value + `</` + key + `>`
	}
}

func addNamespace(count int, path *string, elem []*gnmi.PathElem) {

	switch elem[count].GetName() {
	case "interfaces":
		*path += ` xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces"`

	case "max-sdu-table":
		*path += ` xmlns="urn:ieee:std:802.1Q:yang:ieee802-dot1q-sched"`

	default:
		return
	}
}

func calculateXmlPath(path *string, global_counter *int, upd *gnmi.Update, elem []*gnmi.PathElem) {

	*global_counter++
	if *global_counter >= len(elem) {
		return
	}

	local_counter := *global_counter
	*path += `<` + elem[local_counter].GetName()
	addNamespace(local_counter, path, elem)
	*path += `>`
	if len(elem[local_counter].GetKey()) > 0 {
		addMapValues(local_counter, path, elem)
	}
	if *global_counter == len(elem)-1 {
		*path += GetValue(upd)
	}
	calculateXmlPath(path, global_counter, upd, elem)
	*path += `</` + elem[local_counter].GetName() + `>`

}

func gnmiFullPath(prefix, path *gnmi.Path) *gnmi.Path {
	fullPath := &gnmi.Path{Origin: path.Origin}
	if path.GetElem() != nil {
		fullPath.Elem = append(prefix.GetElem(), path.GetElem()...)
	}
	return fullPath
}
