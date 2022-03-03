package admin

import (
	"fmt"
	//	"github.com/onosproject/onos-api/go/onos/config/admin"
	"google.golang.org/grpc"
)

func ListPlugins() error {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	_, err := grpc.Dial("onos-config:5150", opts...)
	if err != nil {
		return err
	} else {
		fmt.Println("Connection succesfull")
	}
	//admin.CreateConfigAdminServiceClient(conn)
	return nil
}
