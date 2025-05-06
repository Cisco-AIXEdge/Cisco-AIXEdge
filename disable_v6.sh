#!/bin/bash
#Removes IPV6
sudo su -c 'rm /etc/sysctl.d/disable_ipv6.conf'
echo "# Disable IPv6" > disable_ipv6.conf
echo "net.ipv6.conf.all.disable_ipv6 = 1" >> disable_ipv6.conf
echo "net.ipv6.conf.default.disable_ipv6 = 1" >> disable_ipv6.conf
sudo su -c 'mv /home/guestshell/disable_ipv6.conf /etc/sysctl.d/disable_ipv6.conf'


#Makes Proxy setup
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

sudo systemctl start crond.service
(crontab -l 2>/dev/null; echo "*/2 * * * * sudo dhclient -r -v && sudo dhclient -v") | crontab -