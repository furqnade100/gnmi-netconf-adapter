# Demo

The GNMI CLI can be used to demo the `gnmi-netconf-adapter` proof-of-concept in a µONOS kind cluster.

Some example proto files for gNMI requests are supplied in this folder.  

Run adapter e.g. `onit  add gnmi-netconf-adapter   --name nc1428`

Note: As described in [../README.md](../README.md), the device target in the gNMI request from `onos-config` is ignored and a single fixed device used by the adapter; you need access to that device from your host for this demo.
 
We will describe options to interact with the `gnmi-netconf-adapter` for :
1. GNMI CLI to onos-config from the µONOS CLI pod 
2. GNMI CLI  (bypassing onos-config) direct to an exposed port on the adapter


## GNMI CLI to onos-config from the µONOS CLI pod 
Refer to  onos-config's [gnmi.md](https://github.com/onosproject/onos-config/blob/master/docs/gnmi.md) for gnmi details; the examples below are specific to demoing the adapter. 

Logon to CLI pod:
````
 ONOS_CLI_POD=`kubectl -n onos get pods -l type=cli --template  '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'`
 kubectl -n onos exec -it pod/${ONOS_CLI_POD} -- /bin/sh

````

`gnmi-netconf-adapter` appears as a device, e.g. 

````
~ $ onos topo get devices
ID       ADDRESS        VERSION    STATE
nc1428   nc1428:11161   19.3.1.8   GNMI: {Connectivity: REACHABLE, Channel: CONNECTED, Service: AVAILABLE}
````

Via onos-config retrieve a selected part of its model from the  `gnmi-netconf-adapter`. 
Notice no value is returned first time, as only read-only items are stored by onos-conig after the initial device discovery, and this demo is for config items only.

````
 gnmi_cli -address onos-config:5150 -get \
-timeout 5s -en PROTO -alsologtostderr    \
-client_crt /etc/ssl/certs/client1.crt -client_key /etc/ssl/certs/client1.key -ca_crt /etc/ssl/certs/onfca.crt \
-proto "path: < target: 'nc1633',   elem: <name: 'configuration'> elem: <name: 'version'>  >" 


notification: <
  timestamp: 1583864707
  update: <
    path: <
      elem: <
        name: "configuration"
      >
      elem: <
        name: "version"
      >
      target: "nc1633"
    >
  >
>
````

Write a value, e.g. 
````
gnmi_cli -address onos-config:5150 -set \
-proto "update: <path: <target: 'nc1633', elem: <name: 'configuration'> elem: <name: 'version' >> val: <string_val: 'µONOS DEMO'>>" \
-timeout 5s -en PROTO -alsologtostderr \
-client_crt /etc/ssl/certs/client1.crt -client_key /etc/ssl/certs/client1.key -ca_crt /etc/ssl/certs/onfca.crt


response: <
  path: <
    elem: <
      name: "configuration"
    >
    elem: <
      name: "version"
    >
    target: "nc1633"
  >
  op: UPDATE
>
timestamp: 1583865072
extension: <
  registered_ext: <
    id: 100
    msg: "eloquent_wright"
  >
>
````
 
Repeat with different values, and `onos-config` keeps track of changes ...
 
````
onos config get network-changes -v
CHANGE                          INDEX  REVISION  PHASE    STATE     REASON   MESSAGE
amazing_pike                    1       26302    CHANGE   COMPLETE  NONE     
	Device: nc1633 (19.3.1.8)
	|/configuration/version                            |(STRING) FOO                            |false  |


eloquent_wright                 2       26327    CHANGE   COMPLETE  NONE     
	Device: nc1633 (19.3.1.8)
	|/configuration/version                            |(STRING) FOO2                           |false  |


happy_bohr                      3       26700    CHANGE   COMPLETE  NONE     
	Device: nc1633 (19.3.1.8)
	|/configuration/version                            |(STRING) µONOS DEMO                     |false  |

````




## GNMI CLI  (bypassing onos-config) direct to an exposed port on adapter
Set up port-forwarding e.g `kubectl port-forward  -n onos nc1428  11161:11161`  

Run `gnmi_cli` on laptop, e.g.
````
gnmi_cli  \
-address  localhost:11161   \
-client_key  ../pkg/certs/client1.key  \
-client_crt  ../pkg/certs/client1.crt    \
-ca_crt ../pkg/certs/onfca.crt  \
-get -proto "$(cat get.version.gnmi)"


notification: <
  timestamp: 1583768826252168100
  update: <
    path: <
      elem: <
        name: "configuration"
      >
      elem: <
        name: "version"
      >
      target: "nc1428"
    >
    val: <
      string_val: "µONOS demo"
    >
  >
>
````

 Note: 
Insecure mode is also supported e.g.  
 ```` 
 gnmi_cli  \
 -address  localhost:11161  \
 -insecure  \
 -get -proto "$(cat get.version.gnmi)"
````

Retrieve capabilities, e.g. 
````
gnmi_cli  -address  localhost:11161  -insecure  -capabilities 
supported_models: <
  name: "junos-conf-interfaces"
  organization: "Juniper"
  version: "2019-01-01"
>
supported_models: <
  name: "junos-conf-system"
  organization: "Juniper"
  version: "2019-01-01"
>
supported_encodings: JSON
gNMI_version: "0.7.0"

````
