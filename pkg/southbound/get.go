package southbound

import "github.com/Juniper/go-netconf/netconf"

// Requests the full configuration for the target "running"
func GetFullConfig() *netconf.RPCReply {
	reply := sendRPCRequest(netconf.MethodGetConfig("running"))

	return reply
}

// Requests partial configuration according to the xmlRequest for the target "running"
func GetConfig(xmlRequest string) *netconf.RPCReply {
	reply := sendRPCRequest(netconf.MethodGetConfig("running"))

	return reply
}
