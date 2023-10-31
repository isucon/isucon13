# Ansible

## 利用方法

```
$ cd bench
$ make linux_amd64
$ mkdir -p ../provisioning/ansible/roles/bench/files
$ mv bin/bench_linux_amd64 ../provisioning/ansible/roles/bench/files
$ cd ..
$ tar -zcvf webapp.tar.gz webapp
$ mv webapp.tar.gz provisioning/ansible/roles/webapp/files
$ cd provisioning/ansible

# ローカルの場合
$ ansible-playbook -i inventory/localhost application.yml
$ ansible-playbook -i inventory/localhost benchmark.yml

# リモートの場合
$ ansible-playbook -i inventory/hosts.yaml application.yml
$ ansible-playbook -i inventory/hosts.yaml benchmark.yml
```

## 証明書について

*.t.isucon.dev の証明書はISUCON12予選のレポジトリにあるものを使っています
https://github.com/isucon/isucon12-qualify