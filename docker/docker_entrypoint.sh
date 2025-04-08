#!/bin/bash
var1=$1

#set -x

case $var1 in
"bash" | "sh")
  echo $var1
  exec "$@"
  ;;

"tool")
  /app/anylink "$@"
  ;;

*)
  #sysctl -w net.ipv4.ip_forward=1
  #iptables -t nat -A POSTROUTING -s "${IPV4_CIDR}" -o eth0+ -j MASQUERADE
  #iptables -nL -t nat

  # Start the service and first determine whether the configuration file exists
   if [[ ! -f /app/conf/profile.xml ]]; then
    /bin/cp -r /home/conf-bak/* /app/conf/
    echo "After the configuration file is initialized, the container will be forcibly exited. Restart the container."
    exit 1
  fi

  # Compatible with old versions of iptables
  if [[ $IPTABLES_LEGACY == "on" ]]; then
    rm /sbin/iptables
    ln -s /sbin/iptables-legacy /sbin/iptables
  fi

  exec /app/anylink "$@"
  ;;
esac