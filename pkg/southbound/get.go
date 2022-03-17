package southbound

import (
	"fmt"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"golang.org/x/crypto/ssh"
)

// Requests the full configuration for the target "running"
func GetFullConfig() *netconf.RPCReply {
	reply := sendRPCRequest(netconf.MethodGetConfig("running"))

	return reply
}

// // Requests partial configuration according to the xmlRequest for the target "running"
// func GetConfig(xmlRequest string) *netconf.RPCReply {
// 	reply := sendRPCRequest(netconf.Meth("<running /></source><filter><interfaces xmlns=\"urn:ietf:params:xml:ns:yang:ietf-interfaces\"><interface/></interfaces></filter>"))

// 	return reply
// }

// GetConfig returns the full configuration, or configuration starting at <section>.
// <format> can be one of "text" or "xml." You can do sub-sections by separating the
// <section> path with a ">" symbol, i.e. "system>login" or "protocols>ospf>area".
func GetConfig(section, format string) (string, error) {
	secs := strings.Split(section, ">")
	nSecs := len(secs) - 1
	command := fmt.Sprintf("<get-config><source><%s/>", format)
	// if section == "full" {
	// 	command += "</source></get-config>"
	// }
	// if section == "interfaces" {
	command += "</source><filter><interfaces xmlns=\"urn:ietf:params:xml:ns:yang:ietf-interfaces\"><interface/></interfaces></filter></get-config>"
	//}
	if nSecs > 0 {
		command += "<filter>"
		for i := 0; i < nSecs; i++ {
			command += fmt.Sprintf("<%s>", secs[i])
		}
		command += fmt.Sprintf("<%s/>", secs[nSecs])

		for j := nSecs - 1; j >= 0; j-- {
			command += fmt.Sprintf("</%s>", secs[j])
		}
		command += fmt.Sprint("</configuration></get-configuration>")
	}

	sshConfig := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password("")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	//  Start connection to network device
	s, err := netconf.DialSSH(switchAddr, sshConfig)

	if err != nil {
		log.Fatal(err)
	}

	// Close connetion to network device when this function is done executing
	defer s.Close()

	r := netconf.RawMethod(command)
	fmt.Println(r)
	reply, err := s.Exec(r)
	if err != nil {
		return "", err
	}

	return reply.Data, nil
}
