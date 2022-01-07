package southbound

import (
	"github.com/Juniper/go-netconf/netconf"
	"github.com/onosproject/onos-lib-go/pkg/logging"
)

var log = logging.GetLogger("main")

func GetConfig(xmlRequest string) *netconf.RPCReply {

	reply := sendRPCRequest(netconf.MethodGetConfig("running"))

	return reply
}
