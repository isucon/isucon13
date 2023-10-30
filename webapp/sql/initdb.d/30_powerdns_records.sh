#!/usr/bin/env bash

# NOTE: 競技環境ではAnsibleにより適用する。テスト環境は当該スクリプトでAレコードを登録。
# * dockertest(CNAME): webapp (DBはwebappより前に起動しなければならないので、webappのアドレスが事前にわからない. 余計な名前解決が生じるが、clientの動作テストだけで性能が影響しないので許容する)
# * ローカル: 127.0.0.1
# * CI: 準備中
webapp_addr="${WEBAPP_A_RECORD_CONTENT:-127.0.0.1}"
echo "[*] webapp = $webapp_addr"

if [[ $webapp_addr = "webapp" ]]; then
    echo "loading cname records..."
    mysql -uisudns -pisudns isudns -e "INSERT INTO records (domain_id, name, type, content, ttl, prio) VALUES (1, 'u.isucon.dev', 'CNAME', '${webapp_addr}', 3600, NULL);"
    mysql -uisudns -pisudns isudns -e "INSERT INTO records (domain_id, name, type, content, ttl, prio) VALUES (1, 'pipe.u.isucon.dev', 'CNAME', '${webapp_addr}', 3600, NULL);"
    mysql -uisudns -pisudns isudns -e "INSERT INTO records (domain_id, name, type, content, ttl, prio) VALUES (1, 'test001.u.isucon.dev', 'CNAME', '${webapp_addr}', 3600, NULL);"
else
    echo "loading a records..."
    mysql -uisudns -pisudns isudns -e "INSERT INTO records (domain_id, name, type, content, ttl, prio) VALUES (1, 'u.isucon.dev', 'A', '${webapp_addr}', 3600, NULL);"
    mysql -uisudns -pisudns isudns -e "INSERT INTO records (domain_id, name, type, content, ttl, prio) VALUES (1, 'pipe.u.isucon.dev', 'A', '${webapp_addr}', 3600, NULL);"
    mysql -uisudns -pisudns isudns -e "INSERT INTO records (domain_id, name, type, content, ttl, prio) VALUES (1, 'test001.u.isucon.dev', 'A', '${webapp_addr}', 3600, NULL);"
fi
