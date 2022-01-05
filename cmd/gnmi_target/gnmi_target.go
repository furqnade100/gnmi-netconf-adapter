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

// Binary gnmi_target implements a gNMI Target with in-memory configuration and telemetry.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"time"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"golang.org/x/crypto/ssh"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/onosproject/gnxi-simulators/pkg/gnmi"
	"github.com/onosproject/gnxi-simulators/pkg/gnmi/modeldata"
	"github.com/onosproject/gnxi-simulators/pkg/gnmi/modeldata/gostruct"

	"github.com/google/gnxi/utils/credentials"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

var log = logging.GetLogger("main")

func main() {
	model := gnmi.NewModel(modeldata.ModelData,
		reflect.TypeOf((*gostruct.Device)(nil)),
		gostruct.SchemaTree["Device"],
		gostruct.Unmarshal,
		gostruct.ΛEnum)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Supported models:\n")
		for _, m := range model.SupportedModels() {
			fmt.Fprintf(os.Stderr, "  %s\n", m)
		}
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	opts := credentials.ServerCredentials()
	g := grpc.NewServer(opts...)

	var configData []byte
	if *configFile != "" {
		var err error
		configData, err = ioutil.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("Error in reading config file: %v", err)
		}
	}

	s, err := newServer(model, configData)

	if err != nil {
		log.Fatalf("Error in creating gnmi target: %v", err)
	}
	pb.RegisterGNMIServer(g, s)
	reflection.Register(g)

	log.Infof("Starting gNMI agent to listen on %s", *bindAddr)
	listen, err := net.Listen("tcp", *bindAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {

		for {
			s.SetDateTime()
			time.Sleep(time.Second * 1)
		}

	}()

	if flag.Lookup("insecure").Value.String() == "true" {
		log.Infof("Initiating connection now")

		var r = mm()
		log.Infof(r.Data)

		log.Infof("Establishing connection")

	}

	log.Infof("Starting gNMI agent to serve on %s", *bindAddr)
	if err := g.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func mm() *netconf.RPCReply {

	sshConfig := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password("")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	s, err := netconf.DialSSH("192.168.0.1", sshConfig)

	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	// Sends raw XML
	// reply, err := s.Exec(netconf.MethodGetConfig("running"))

	const changes = `<interfaces xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces">
	<interface>
	   <name>sw0p5</name>
	   <max-sdu-table xmlns="urn:ieee:std:802.1Q:yang:ieee802-dot1q-sched">
		  <traffic-class>0</traffic-class>
		  <queue-max-sdu>1504</queue-max-sdu>
	   </max-sdu-table>
	</interface>
 </interfaces>`

	reply, err := s.Exec(MethodEditConfig("running", changes))
	if err != nil {
		panic(err)
	}
	return reply
}

// MethodEditConfig files a NETCONF edit-config request with the remote host
func MethodEditConfig(database string, dataXml string) netconf.RawMethod {
	const editConfigXml = `<edit-config>
	<target><%s/></target>
	<default-operation>merge</default-operation>
	<error-option>rollback-on-error</error-option>
	<config>%s</config>
	</edit-config>`
	return netconf.RawMethod(fmt.Sprintf(editConfigXml, database, dataXml))
}
