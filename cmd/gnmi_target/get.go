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
	"time"

	"github.com/google/gnxi/utils/credentials"
	"github.com/openconfig/gnmi/proto/gnmi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Get overrides the Get func of gnmi.Target to provide user auth.
func (s *server) Get(ctx context.Context, req *gnmi.GetRequest) (*gnmi.GetResponse, error) {
	msg, ok := credentials.AuthorizeUser(ctx)
	if !ok {
		log.Infof("denied a Get request: %v", msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}

	log.Infof("allowed a Get request: %v", msg)

	/**********************************************************
	Implementation of data conversion should be initiated here.
	***********************************************************/
	// Example of data conversion initiation
	log.Infof("The incoming get request contains: %s", req.String())
	//dataConv.Convert(req, "Get")

	notifications := make([]*gnmi.Notification, 1)
	prefix := req.GetPrefix()
	ts := time.Now().UnixNano()

	notifications[0] = &gnmi.Notification{
		Timestamp: ts,
		Prefix:    prefix,
	}

	resp := &gnmi.GetResponse{Notification: notifications}

	return resp, nil
	// return s.Server.Get(ctx, req)
}
