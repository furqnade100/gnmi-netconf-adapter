package admin

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/onosproject/onos-api/go/onos/config/admin"
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func ListPlugins() error {
	// opts := []grpc.DialOption{
	// 	grpc.WithInsecure(),
	// }
	var opts []grpc.DialOption
	//opts := credentials.ClientCredentials("onos-config")
	cert, err := tls.X509KeyPair([]byte(certs.DefaultClientCrt), []byte(certs.DefaultClientKey))
	if err != nil {
		fmt.Println("custom: error getting crts")
		return err
	}
	opts = []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		})),
	}

	conn, err := grpc.Dial("onos-config:5150", opts...)
	if err != nil {
		return err
	} else {
		fmt.Println("Connection succesfull")
	}
	client := admin.CreateConfigAdminServiceClient(conn)

	req := admin.ListModelsRequest{
		Verbose: true,
	}

	stream, err := client.ListRegisteredModels(context.Background(), &req)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println("List plugins succesfull")
	}
	fmt.Println(stream)

	return nil
}
