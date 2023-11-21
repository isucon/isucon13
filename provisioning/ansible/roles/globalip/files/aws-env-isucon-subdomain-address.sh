#!/usr/bin/env bash
set -eux

if (test -f /opt/aws-env-isucon-subdomain-address.sh.lock); then
  exit
fi

touch /opt/aws-env-isucon-subdomain-address.sh.lock

if !(grep Amazon /sys/devices/virtual/dmi/id/bios_vendor); then
  exit
fi

instance_ip=$(curl -s https://checkip.amazonaws.com)
if [ -n $instance_ip ]; then
  sed -Ei 's/^ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS=.+$/ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS="'${instance_ip}'"/' /home/isucon/env.sh
fi
chmod 0755 /home/isucon/env.sh
chown isucon:isucon /home/isucon/env.sh