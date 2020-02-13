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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/openconfig/gnmi/value"

	"github.com/openconfig/goyang/pkg/yang"

	"github.com/damianoneill/net/v2/netconf/ops"

	pb "github.com/openconfig/gnmi/proto/gnmi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog"
)

// Get implements the Get RPC in gNMI spec.
func (a *Adapter) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {

	if err := a.checkEncodingAndModel(req.GetEncoding(), req.GetUseModels()); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	prefix := req.GetPrefix()
	paths := req.GetPath()
	notifications := make([]*pb.Notification, len(paths))

	for i, path := range paths {
		// Get schema node for path from config struct.
		fullPath := path
		if prefix != nil {
			fullPath = gnmiFullPath(prefix, path)
		}

		if fullPath.GetElem() == nil && fullPath.GetElement() != nil {
			return nil, status.Error(codes.Unimplemented, "deprecated path element type is unsupported")
		}

		entry := a.getSchemaEntryForPath(fullPath)
		if entry == nil {
			return nil, status.Errorf(codes.NotFound, "path %v not found (Test)", fullPath)
		}

		filter := getSubtreeFilterForPath(fullPath)

		result := ""
		err := a.ncs.GetConfigSubtree(filter, ops.CandidateCfg, &result)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "Failed to get config for %v %v", fullPath, err)
		}
		ts := time.Now().UnixNano()

		netconfMap := a.netconfToJSON(result)

		target, err := getTarget(netconfMap, fullPath)
		if err != nil {
			return nil, err
		}

		if entry.IsLeaf() {
			var err error
			var val *pb.TypedValue
			val, err = value.FromScalar(reflect.ValueOf(&target).Elem().Interface())
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
		if entry.IsDir() {
			if err != nil {
				msg := fmt.Sprintf("error in constructing %s JSON tree from requested node: %v", "Internal", err)
				log.Error(msg)
				return nil, status.Error(codes.Internal, msg)
			}

			jsonDump, err := json.Marshal(target)
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
		var nextmap map[string]interface{}
		switch v := value.(type) {
		case map[string]interface{}:
			nextmap = v
		case []interface{}:
			nextmap = v[0].(map[string]interface{})
		}
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

func (a *Adapter) getSchemaEntryForPath(path *pb.Path) *yang.Entry {
	rootEntry := a.model.schemaTreeRoot
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

func (a *Adapter) netconfToJSON(result string) map[string]interface{} {
	dec := xml.NewDecoder(strings.NewReader(result))

	type eldesc struct {
		schema   *yang.Entry
		tag      string
		value    interface{}
		children map[string]interface{}
	}
	top := make(map[string]interface{})
	stack := []*eldesc{}
	var cureld *eldesc

	schema := a.model.schemaTreeRoot

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
					if cureld.schema != nil {
						if cureld.schema.IsLeaf() || cureld.schema.IsLeafList() {
							// TODO List!
							cureld.value = getLeafValue(v, cureld.schema)
						}
					}
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

func getLeafValue(v xml.CharData, schema *yang.Entry) interface{} {
	// TODO SJ Handle data according to leaf type.

	switch schema.Type.Kind {
	case yang.Ystring:
		return strings.TrimSpace(string(v))
	case yang.Yunion:
		// TODO Iterate over types
		val, _ := getUnionValue(strings.TrimSpace(string(v)), schema.Type.Type)
		//val, _ := strconv.ParseUint(strings.TrimSpace(string(v)), 10, 64)
		return val
	case yang.Yenum:
		// TODO Check what else needs done
		return strings.TrimSpace(string(v))
	}
	// TODO Handle other kinds
	fmt.Printf("Leaf kind %s still to be supported\n", schema.Type.Kind)
	return strings.TrimSpace(string(v))
}

func getChildSchema(name string, parent *yang.Entry) *yang.Entry {
	// Ignore any elements that are not in the schema.
	return parent.Dir[name]
}

func getUnionValue(v string, types []*yang.YangType) (interface{}, error) {
	for _, t := range types {
		switch t.Kind {
		case yang.Ystring:
			if isValidString(v, t) {
				return v, nil
			}
		case yang.Yint32:
			val := isValidInt(v, t)
			if val != nil {
				return val, nil
			}
			// TODO Add other kinds.
		}
	}
	return nil, status.Errorf(codes.NotFound, "failed to set union value: %s", v)
}

func isValidString(v string, t *yang.YangType) bool {
	if !anyPatternMatches(v, t.Pattern) {
		return false
	}
	// TODO Range checks?
	return true
}

func isValidInt(v string, t *yang.YangType) interface{} {
	val, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return nil
	}

	for _, r := range t.Range {
		if val >= int64(r.Min.Value) && val <= int64(r.Max.Value) {
			return val
		}
	}

	return nil
}

func anyPatternMatches(v string, patterns []string) bool {
	for _, p := range patterns {
		if !patternMatches(v, p) {
			return false
		}
	}
	return true
}

func patternMatches(v string, p string) bool {

	r, err := regexp.Compile(fixYangRegexp(p))
	if err != nil {
		return false
	}
	// fixYangRegexp adds ^(...)$ around the pattern - the result is
	// equivalent to a full match of whole string.
	return r.MatchString(v)
}

// Following function is lifted unchanged from https://github.com/openconfig/ygot/blob/master/ytypes/string_type.go

// fixYangRegexp takes a pattern regular expression from a YANG module and
// returns it into a format which can be used by the Go regular expression
// library. YANG uses a W3C standard that is defined to be implicitly anchored
// at the head or tail of the expression. See
// https://www.w3.org/TR/2004/REC-xmlschema-2-20041028/#regexs for details.
func fixYangRegexp(pattern string) string {
	var buf bytes.Buffer
	var inEscape bool
	var prevChar rune
	addParens := false

	for i, ch := range pattern {
		if i == 0 && ch != '^' {
			buf.WriteRune('^')
			// Add parens around entire expression to prevent logical
			// subexpressions associating with leading/trailing ^ / $.
			buf.WriteRune('(')
			addParens = true
		}

		switch ch {
		case '$':
			// Dollar signs need to be escaped unless they are at
			// the end of the pattern, or are already escaped.
			if !inEscape && i != len(pattern)-1 {
				buf.WriteRune('\\')
			}
		case '^':
			// Carets need to be escaped unless they are already
			// escaped, indicating set negation ([^.*]) or at the
			// start of the string.
			if !inEscape && prevChar != '[' && i != 0 {
				buf.WriteRune('\\')
			}
		}

		// If the previous character was an escape character, then we
		// leave the escape, otherwise check whether this is an escape
		// char and if so, then enter escape.
		inEscape = !inEscape && ch == '\\'

		buf.WriteRune(ch)

		if i == len(pattern)-1 {
			if addParens {
				buf.WriteRune(')')
			}
			if ch != '$' {
				buf.WriteRune('$')
			}
		}

		prevChar = ch
	}

	return buf.String()
}
