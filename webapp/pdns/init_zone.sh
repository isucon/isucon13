#!/usr/bin/env bash

set -eux
cd $(dirname $0)

if test -f /home/isucon/env.sh; then
	. /home/isucon/env.sh
fi
if test -f /home/isucon/env-isucon-subdomain-address.sh; then
	. /home/isucon/env-isucon-subdomain-address.sh
fi

temp_dir=$(mktemp -d)
trap 'rm -rf $temp_dir' EXIT
sed 's/<ISUCON_SUBDOMAIN_ADDRESS>/'$ISUCON_SUBDOMAIN_ADDRESS'/g' u.isucon.dev.zone > ${temp_dir}/u.isucon.dev.zone
pdnsutil load-zone u.isucon.dev ${temp_dir}/u.isucon.dev.zone

