#!/bin/bash
set -eux

while IFS=' ' read -r team_id ip_address; do
    resolved_ip=$(dig +short pipe.u.isucon.dev "@$ip_address")
    curl -s --resolve "pipe.u.isucon.dev:443:${resolved_ip}" "https://pipe.u.isucon.dev/api/livestream/search?limit=50" > "${team_id}.json"
done
