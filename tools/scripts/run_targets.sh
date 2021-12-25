#!/bin/bash

hostname=${HOSTNAME:-localhost}

# $HOME/certs/generate_certs.sh $hostname

# echo $hostname

# if [ "$hostname" != "localhost" ]; then \
#     IPADDR=`ip route get 1.2.3.4 | grep dev | awk '{print $7}'`
#     $HOME/certs/generate_certs.sh $hostname > /dev/null 2>&1;
#     echo "Please add '"$IPADDR" "$hostname"' to /etc/hosts and access with gNMI client at "$hostname":"$GNMI_PORT; \
# else \
#     echo "gNMI target in secure mode is on $hostname:"${GNMI_PORT};
#     echo "gNMI target insecure mode is on $hostname:"${GNMI_INSECURE_PORT};
# fi
sed -i -e "s/replace-device-name/"$hostname"/g" $HOME/target_configs/typical_ofsw_config.json && \
sed -i -e "s/replace-motd-banner/Welcome to gNMI service on "$hostname":"$GNMI_PORT"/g" $HOME/target_configs/typical_ofsw_config.json

gnmi_target \
    -bind_address :$GNMI_INSECURE_PORT \
    -alsologtostderr \
    -notls \
    -insecure \
    -config $HOME/target_configs/typical_ofsw_config.json &

gnmi_target \
    -bind_address :$GNMI_PORT \
    -key $HOME/certs/localhost.key \
    -cert $HOME/certs/localhost.crt \
    -ca $HOME/certs/onfca.crt \
    -alsologtostderr \
    -config $HOME/target_configs/typical_ofsw_config.json
    # -key $HOME/certs/$hostname.key \
    # -cert $HOME/certs/$hostname.crt \

# ls certs/