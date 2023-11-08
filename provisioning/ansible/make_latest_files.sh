#!/usr/bin/env bash

set -eux
cd $(dirname $0)

cd ../../bench
make linux_amd64
mkdir -p ../provisioning/ansible/roles/bench/files
mv bin/bench_linux_amd64 ../provisioning/ansible/roles/bench/files
cd ..
tar -zcvf webapp.tar.gz webapp
mv webapp.tar.gz provisioning/ansible/roles/webapp/files

cd $(dirname $0)



