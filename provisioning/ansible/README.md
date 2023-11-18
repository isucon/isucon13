# Ansible

## 利用方法

```
# ベンチマーカ、webappのアーカイブをアップデート　
$ ./make_latest_files.sh

# ローカルの場合
$ ansible-playbook -i inventory/localhost application.yml
$ ansible-playbook -i inventory/localhost benchmark.yml

# sacloud試し環境へのリモート実行
$ ansible-playbook -i inventory/sacloud application.yml
$ ansible-playbook -i inventory/sacloud benchmark.yml

```

すでに対象サーバに /home/isucon/webapp/sql がある場合、tarをアップロードするだけで展開はしません

## 証明書について

*.t.isucon.dev の証明書はISUCON12予選のレポジトリにあるものを使っています
https://github.com/isucon/isucon12-qualify


## make_lastest_filesの中身

```
$ cd bench
$ make linux_amd64
$ mkdir -p ../provisioning/ansible/roles/bench/files
$ mv bin/bench_linux_amd64 ../provisioning/ansible/roles/bench/files
$ cd ..
$ tar -zcvf webapp.tar.gz webapp
$ mv webapp.tar.gz provisioning/ansible/roles/webapp/files
$ cd provisioning/ansible
```