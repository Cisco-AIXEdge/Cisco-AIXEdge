#!/bin/bash

output=$(dohost "show runnin | sec proxy-port")
ip=$(echo "$output" | grep -oP 'proxy-server \K[^ ]+')
port=$(echo "$output" | grep -oP 'proxy-port \K[^ ]+')
sudo su -c 'rm /etc/profile.d/proxy.sh'
if [ -n "$ip" ] && [ -n "$port" ]; then
    echo "export http_proxy=http://$ip:$port/" > "/home/guestshell/proxy.sh"
    echo "export HTTP_PROXY=http://$ip:$port/" >> "/home/guestshell/proxy.sh"
    echo "export https_proxy=http://$ip:$port/" >> "/home/guestshell/proxy.sh"
    echo "export HTTPS_PROXY=http://$ip:$port/" >> "/home/guestshell/proxy.sh"
    sudo su -c 'mv /home/guestshell/proxy.sh /etc/profile.d/proxy.sh'
    source /etc/profile.d/proxy.sh
fi


