#!/bin/sh

find /var/lib/mysql -type f | xargs cat > /dev/null
find /home/isucon/webapp/sql -type f | xargs cat > /dev/null