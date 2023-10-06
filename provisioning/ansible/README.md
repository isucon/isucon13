# Ansible

## 利用方法

```
$ cd bench
$ make build
$ mv bin/bench_linux_amd64 ../provisioning/ansible/roles/bench/files
$ cd ..
$ tar -zcvf webapp.tar.gz webapp
$ mv webapp.tar.gz provisioning/ansible/roles/webapp/files
$ mv webapp.tar.gz provisioning/ansible/roles/webapp/files
$ ansible-playbook -i inventory/hosts.yaml application.yaml
$ ansible-playbook -i inventory/hosts.yaml benchmarker.yaml
```
