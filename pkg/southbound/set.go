package southbound

import "github.com/Juniper/go-netconf/netconf"

/***********************************************
Example of xml that updates a queue-max-sdu tag.
************************************************/

/*
*	<interfaces xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces">
*		<interface>
*			<name>sw0p5</name>
*			<max-sdu-table xmlns="urn:ieee:std:802.1Q:yang:ieee802-dot1q-sched">
*				<traffic-class>0</traffic-class>
*				<queue-max-sdu>1504</queue-max-sdu>
*			</max-sdu-table>
*		</interface>
*	</interfaces>
 */

// Updates the configuration accoring to the input xml for the target "running"
func UpdateConfig(xmlChanges string) *netconf.RPCReply {

	//reply := sendRPCRequest(methodEditConfig("running", xmlChanges))
	//reply := sendRPCRequest()

	//const changes = `<interfaces xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces">
	//<interface>
	//   <name>sw0p5</name>
	//   <max-sdu-table xmlns="urn:ieee:std:802.1Q:yang:ieee802-dot1q-sched">
	//	  <traffic-class>0</traffic-class>
	//	  <queue-max-sdu>500</queue-max-sdu>
	//   </max-sdu-table>
	//</interface>
	//</interfaces>`

	reply := sendRPCRequest(methodEditConfig("running", xmlChanges))

	return reply
}
