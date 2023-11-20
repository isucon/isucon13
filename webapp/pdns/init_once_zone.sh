#!/usr/bin/env bash

set -eux
cd $(dirname $0)
# PowerDNS の起動後に呼ばれ、ゾーンがない場合に初期化を行います

if test -f /home/isucon/env.sh; then
	. /home/isucon/env.sh
fi

ISUCON_SUBDOMAIN_ADDRESS=${ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS:-127.0.0.1}

if !(pdnsutil list-all-zones | grep  u.isucon.dev); then
    pdnsutil create-zone u.isucon.dev
    pdnsutil add-record u.isucon.dev "." A 30 $ISUCON_SUBDOMAIN_ADDRESS
    pdnsutil add-record u.isucon.dev "pipe" A 30 $ISUCON_SUBDOMAIN_ADDRESS
    pdnsutil add-record u.isucon.dev "test001" A 30 $ISUCON_SUBDOMAIN_ADDRESS
fi

exit
