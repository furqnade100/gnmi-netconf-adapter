package admin

import (
	"context"
	"fmt"

	"github.com/google/gnxi/utils/credentials"
	"github.com/onosproject/onos-api/go/onos/config/admin"
	"google.golang.org/grpc"
)

func ListPlugins() error {
	// opts := []grpc.DialOption{
	// 	grpc.WithInsecure(),
	// }

	opts := credentials.ClientCredentials("onos-config")
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
