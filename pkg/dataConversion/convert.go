package dataConversion

import (
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
	// Example of data conversion initiations furqan updates
	log.Infof(gnmi.SetRequest.Data)
	xmlRequest := json2Xml(req.String())

	var reply = ""

	// Initiate southbound NETCONF client, sending the xml
	reply = sb.UpdateConfig(xmlRequest).Data

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
