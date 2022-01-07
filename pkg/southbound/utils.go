package southbound

import (
	"github.com/Juniper/go-netconf/netconf"
	"golang.org/x/crypto/ssh"
)

const switchAddr = "192.168.0.1"

// Takes in an RPCMethod function and executes it, then returns the reply from the network device
func sendRPCRequest(fn netconf.RPCMethod) *netconf.RPCReply {
	//  Define config for connection to network device
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

	// Executes the function passed as fn
	reply, err := s.Exec(fn)

	if err != nil {
		panic(err)
	}

	return reply
}
