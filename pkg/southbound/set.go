package southbound

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

const test = "<interfaces xmlns=\"urn:ietf:params:xml:ns:yang:ietf-interfaces\"><interface><name>sw0p5</name><enabled>false</enabled></interface></interfaces>"

// Updates the configuration accoring to the input xml for the target "running"
func UpdateConfig(xmlChanges string) {

	//reply := sendRPCRequest(methodEditConfig("running", xmlChanges))
	//reply := sendRPCRequest()
	log.Infof(string(methodEditConfig("running", test)))
}
