package southbound

import (
	"fmt"

	"github.com/Juniper/go-netconf/netconf"
)

func UpdateConfig() *netconf.RPCReply {

	// sshConfig := &ssh.ClientConfig{
	// 	User:            "root",
	// 	Auth:            []ssh.AuthMethod{ssh.Password("")},
	// 	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	// }

	// s, err := netconf.DialSSH("192.168.0.1", sshConfig)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer s.Close()

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

	reply := sendRPCRequest(MethodEditConfig("running", changes))
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
