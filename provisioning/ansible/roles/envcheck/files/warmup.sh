#!/bin/sh

find /var/lib/mysql -type f | xargs cat > /dev/null
find /home/isucon/webapp -type f | xargs cat > /dev/null