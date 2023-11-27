# ISUCON13 問題

## 当日に公開したマニュアルおよびアプリケーションについての説明

- [ISUCON13 当日マニュアル](https://gist.github.com/kazeburo/bccc2d2b2b9dc307b5640ae855f3e0bf)
- [ISUCON13 アプリケーションマニュアル](https://gist.github.com/kazeburo/70b352e6d51969b214f919bcf0794ba6)


## ディレクトリ構成

```
.
+- bech           # ベンチマーカー
+- development    # 開発環境用 docker compose
+- docs           # ドキュメント類
+- envcheck       # EC2サーバー 環境確認用プログラム
+- frontend       # フロントエンド
+- provisioning   # Ansible および Packer
+- scripts        # 初期、ベンチマーカー用データ生成用スクリプト
+- validated      # 競技後、最終チェックに用いたデータ
+- webapp         # 参考実装
```

## ISUCON13 予選当日との変更点

### Node.JSへのパッチ

当日、アプリケーションマニュアルにて公開した Node.JSへのパッチは適用済みです。[#408](https://github.com/isucon/isucon13/pull/408)

## TLS証明書について

ISUCON13で使用したTLS証明書は `provisioning/ansible/roles/nginx/files/etc/nginx/tls` 以下にあります。

本証明書は有効期限が切れている可能性があります。定期的な更新については予定しておりません。

## ISUCON13のインスタンスタイプ

- 競技者 VM 3台
  - InstanceType: c5.large (2vCPU, 4GiB Mem)
  - VolumeType: gp3 40GB
- ベンチマーカー VM 1台
  - ECS Fargate (8vCPU, 8GB Mem)

## AWS上での過去問環境の構築方法

### 用意されたAMIを利用する場合

リージョン ap-northeast-1 AMI-ID ami-06c947ddf8c38c43c で起動してください。 

このAMIは予告なく利用できなくなる可能性があります。

このAMIで Node.JS の参考実装を利用する際は [#408](https://github.com/isucon/isucon13/pull/408) の変更を適用してください。

### 自分でAMIをビルドする場合

上記AMIが利用できなくなった場合は、 `provisioning/packer` 以下でmake buildを実行するとAMIをビルドできます。packer が必要です。(運営時に検証したバージョンはv1.9.4)

ベースとなるAMIが利用できない場合は以下のパッチを適用する必要があります。

```
diff --git a/provisioning/packer/isucon13.pkr.hcl b/provisioning/packer/isucon13.pkr.hcl
index 4123481..0b7a676 100644
--- a/provisioning/packer/isucon13.pkr.hcl
+++ b/provisioning/packer/isucon13.pkr.hcl
@@ -52,7 +52,7 @@ source "amazon-ebs" "isucon13" {
   tags          = local.ami_tags
   snapshot_tags = local.ami_tags
 
-  source_ami    = "ami-03bd3273f34a1f122"
+  source_ami    = "${data.amazon-ami.ubuntu-jammy.id}"
   region        = "ap-northeast-1"
   instance_type = "c5.4xlarge"
```

### AMIからEC2を起動する場合の注意事項

- 起動に必要なEBSのサイズは最低8GBですが、ベンチマーク中にデータが増えると溢れる可能性があるため、大きめに設定することをお勧めします(競技環境では40GiB)
- セキュリティグループは `TCP/443` 、 `TCP/22` に加え、 `UDP/53` を必要に応じて開放してください
- 適切なインスタンスプロファイルを設定することで、セッションマネージャーによる接続が可能です
- 起動時に指定したキーペアで `ubuntu` ユーザーでSSH可能です
  - その後 `sudo su - isucon` で `isucon` ユーザーに切り替えてください

## docker compose での構築方法

開発に利用した docker composeで環境を構築することもできます。ただし、スペックやTLS証明書の有無など競技環境とは異なります。

```
$ cd development
$ make go/down
$ make go/up
```

go以外の環境の起動は `{言語実装名}/down`  および `{言語実装名}/up` で行えます。


## ベンチマーカーの実行

### ベンチマーカーのビルド

ベンチマーカーは参考実装のAMIには含まれません。本Repositoryを git clone しビルドを行なってください。

```
$ cd bench
$ make
```

macOSとLinux用のバイナリが作成されます。

### ベンチマーカーの実行

docker compose 環境の場合、次のようにベンチマークを実行します

```
$ ./bench_darwin_arm64 run --dns-port=1053 # M1系macOSの場合
```

競技環境に向けては次のように実行します

```
$ ./bench_linux_amd64 run --target https://pipe.u.isucon.dev --nameserver 163.43.129.52 --enable-ssl --webapp {他のサーバ1} --webapp {他のサーバ2}
```
- `--nameserver`　は、ベンチマーカーが名前解決を行うサーバーのIPアドレスを指定して下さい
- `--webapp` は、名前解決を行うDNSサーバーが名前解決の結果返却する可能性があるIPアドレスを指定して下さい
  - 1台のサーバーで競技を行う場合は指定不要です
  - 複数台で競技を行う場合は、`--nameserver` に指定したアドレスを除いた、競技に使用するサーバーのIPアドレスを指定してください
またベンチマークコマンドに、 `--pretest-only` を付加することで、初期化処理と整合性チェックのみを行うことができます。アプリケーションの動作確認に利用してください。

## フロントエンドおよび動画配信について

フロントエンドの実装はリポジトリに存在していますが、競技の際に利用した動画とサムネイルについては配信サーバを廃止しており、表示できません。


## Links

- [ISUCON13 まとめ](https://isucon.net/archives/57801192.html)

