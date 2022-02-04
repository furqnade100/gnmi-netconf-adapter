package dataConversion

import (
	//sb "github.com/onosproject/gnmi-netconf-adapter/pkg/southbound"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	//"github.com/openconfig/gnmi/proto/gnmi"
	"context"

	"github.com/damianoneill/net/v2/netconf/ops"
	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var log = logging.GetLogger("main")
var ncs ops.OpSession

// // Takes in a gnmi get request and returns a gnmi get response.
// func Convert(req *gnmi.SetRequest) { //*gnmi.GetRequest, typeOfRequest string) {

// 	/************************************************************
// 	Implementation of data conversion should be implemented here.
// 	*************************************************************/
// 	// Example of data conversion initiations furqan updates
// 	log.Infof(req.String())
// 	/*xmlRequest := json2Xml(req.String())

// 	var reply = ""

// 	// Initiate southbound NETCONF client, sending the xml
// 	reply = sb.UpdateConfig(xmlRequest).Data

// 	// Logs the reply, before sending back the response a conversion from xml to json should be implemented.
// 	log.Infof(reply)
// 	*/
// 	// Simulated response.
// 	//notifications := make([]*gnmi.Notification, 1)
// 	//prefix := req.GetPrefix()
// 	//ts := time.Now().UnixNano()

// 	//notifications[0] = &gnmi.Notification{
// 	//	Timestamp: ts,
// 	//	Prefix:    prefix,
// 	//}

// 	//resp := &gnmi.GetResponse{Notification: notifications}
// 	//return resp
// 	//return reply
// }

func Set(ctx context.Context, req *gnmi.SetRequest) (*gnmi.SetResponse, error) {

	prefix := req.GetPrefix()
	var results []*gnmi.UpdateResult

	// Execute operations in order.
	// Reference: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#34-modifying-state

	// Execute Deletes
	for _, path := range req.GetDelete() {
		res, grpcStatusError := executeOperation(gnmi.UpdateResult_DELETE, prefix, path, nil)
		if grpcStatusError != nil {
			return nil, grpcStatusError
		}
		results = append(results, res)
	}

	// Execute Replaces
	for _, upd := range req.GetReplace() {
		res, grpcStatusError := executeOperation(gnmi.UpdateResult_REPLACE, prefix, upd.GetPath(), upd.GetVal())
		if grpcStatusError != nil {
			return nil, grpcStatusError
		}
		results = append(results, res)
	}

	// Execute Updates
	for _, upd := range req.GetUpdate() {
		res, grpcStatusError := executeOperation(gnmi.UpdateResult_UPDATE, prefix, upd.GetPath(), upd.GetVal())
		if grpcStatusError != nil {
			return nil, grpcStatusError
		}
		results = append(results, res)
	}

	return &gnmi.SetResponse{
		Prefix:   prefix,
		Response: results,
	}, nil
}

func executeOperation(op gnmi.UpdateResult_Operation, prefix, path *gnmi.Path, val *gnmi.TypedValue) (*gnmi.UpdateResult, error) {

	request, err := gnmiToNetconfOperation(op, prefix, path, val)
	if err != nil {
		return nil, err
	}

	err = ncs.EditConfigCfg(ops.CandidateCfg, request)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "edit failed %s", err)
	}

	return &gnmi.UpdateResult{
		Path: path,
		Op:   op,
	}, nil
}

func gnmiToNetconfOperation(op gnmi.UpdateResult_Operation, prefix, path *gnmi.Path, inval *gnmi.TypedValue) (interface{}, error) {

	// fullPath := path
	// if prefix != nil {
	// 	fullPath = gnmiFullPath(prefix, path)
	// }

	// entry := getSchemaEntryForPath(fullPath)
	// if entry == nil {
	// 	return nil, status.Errorf(codes.NotFound, "path %v not found (Test)", fullPath)
	// }

	// var buf bytes.Buffer
	// enc := xml.NewEncoder(&buf)

	// mapPathToNetconf(fullPath, op, enc)

	// if op != gnmi.UpdateResult_DELETE {
	// 	err := a.mapSetValueToNetconf(enc, entry, inval)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	// // Close off the XML elements defined by the path.
	// for i := len(fullPath.Elem) - 1; i >= 0; i-- {
	// 	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: fullPath.Elem[i].Name}})
	// }

	// // Return the XML document.
	// _ = enc.Flush()
	// filter := buf.String()
	// if len(filter) == 0 {
	// 	return nil, nil
	// }
	// return filter, nil
	return nil, nil
}
