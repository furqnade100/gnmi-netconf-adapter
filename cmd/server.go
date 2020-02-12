// Copyright 2020-present Open Networking Foundation.
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

package cmd

import (
	"crypto/tls"
	"crypto/x509"
	log "github.com/golang/glog"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	grpccredentials "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"net"
	"strconv"
)

var (
	myca         *string
	mykey        *string
	mycert       *string
	myIsInsecure *bool
	myport       *int
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts up a gNMI listener",
	Run:   RunGnmiServer,
}

func init() {

	myca = serverCmd.Flags().String("ca", "", "path to CA certificate")
	mykey = serverCmd.Flags().String("key", "", "path to client private key")
	mycert = serverCmd.Flags().String("cert", "", "path to client certificate")
	myport = serverCmd.Flags().Int("port", 10999, "port to listen")
	myIsInsecure = serverCmd.Flags().Bool("insecure", false, "whether to enable skip verification")

	rootCmd.AddCommand(serverCmd)

}

// RunGnmiServer provides an indirection so that the logic can be tested independently of the cobra infrastructure
func RunGnmiServer(command *cobra.Command, args []string) {
	log.Info("Run GNMI Server... ")
	err := Serve(func(started string) {
		log.Info("Started listener on ", started)
	})

	log.Error("Error running Serve", err)

}

// Serve starts the NB gNMI server.
func Serve(started func(string)) error {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(*myport))
	if err != nil {
		return err
	}

	tlsCfg := &tls.Config{}
	clientCerts, err := tls.LoadX509KeyPair(*mycert, *mykey)
	if err != nil {
		log.Info("Error loading certs", clientCerts, err)
	}
	tlsCfg.Certificates = []tls.Certificate{clientCerts}
	tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
	tlsCfg.ClientCAs = getCertPool(*myca)
	if *myIsInsecure {
		tlsCfg.ClientAuth = tls.RequestClientCert
	} else {
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
	}

	opts := []grpc.ServerOption{grpc.Creds(grpccredentials.NewTLS(tlsCfg))}
	grpcServer := grpc.NewServer(opts...)

	s, err := newGnmiServer(model)
	if err != nil {
		log.Fatal("newGnmiServer error ", err)
	}

	pb.RegisterGNMIServer(grpcServer, s)
	reflection.Register(grpcServer)

	started(lis.Addr().String())
	return grpcServer.Serve(lis)
}

func getCertPool(CaPath string) *x509.CertPool {
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(CaPath)
	if err != nil {
		log.Warning("could not read ", CaPath, err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Warning("failed to append CA certificates")
	}
	return certPool
}
