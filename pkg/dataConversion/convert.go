package dataConversion

import (
	"fmt"
	"strconv"

	sb "github.com/onosproject/gnmi-netconf-adapter/pkg/southbound"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/openconfig/gnmi/proto/gnmi"
)

var log = logging.GetLogger("main")

// Takes in a gnmi get request and returns a gnmi get response.
func Convert(req *gnmi.SetRequest) { //*gnmi.GetRequest, typeOfRequest string) {

	/************************************************************
	Implementation of data conversion should be implemented here.
	*************************************************************/
	// Example of data conversion initiation
	// log.Infof(req.Prefix.Origin)
	// log.Infof(req.Prefix.Target)

	// xmlRequest := json2Xml(req.String())

	// var reply = ""

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

	// Initiate southbound NETCONF client, sending the xml
	reply := sb.UpdateConfig(xmlPath).Data

	// Logs the reply, before sending back the response a conversion from xml to json should be implemented.
	log.Infof(reply)

	// Simulated response.
	//notifications := make([]*gnmi.Notification, 1)
	//prefix := req.GetPrefix()
	//ts := time.Now().UnixNano()

	//notifications[0] = &gnmi.Notification{
	//	Timestamp: ts,
	//	Prefix:    prefix,
	//}

	//resp := &gnmi.GetResponse{Notification: notifications}
	//return resp
	//return reply
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

// func gnmiFullPath(prefix, path *gnmi.Path) *gnmi.Path {
// 	fullPath := &gnmi.Path{Origin: path.Origin}
// 	if path.GetElem() != nil {
// 		fullPath.Elem = append(prefix.GetElem(), path.GetElem()...)
// 	}
// 	return fullPath
// }
