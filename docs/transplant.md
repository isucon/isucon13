# 移植マニュアル


## Webappについて

本問題のWebappは、以下のミドルウェアと連携して動作します

* Nginx, MySQL8 (ここは例年同様なので細かい話は割愛)
* PowerDNS (新顔)
    * `pipe.u.isucon.dev` 及び `<username>.u.isucon.dev` のドメインを管理しています
    * これらドメインの名前解決により、Nginxにアクセスする想定です
    * SSL証明書の準備を進めており、HTTPS化を予定しております
    * webappの環境変数である `ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS` は、これらサブドメインのAレコードを解決した際のIPアドレスが指定されます
        * 検証環境では同一ホスト上で動作させているため、一貫して `127.0.0.1` を指定しています
        * 競技環境では、競技サーバのIPアドレスが指定されることになります
    * webappは、ユーザ登録時に `<username>.u.isucon.dev` のドメイン登録処理を行います
        * これには `pdnsutil` というPowerDNSのCLIを用います
        * docker-composeでは、webappコンテナにpowerdnsパッケージをインストールしています
            * debianの場合、 `pdns-server` および `pdns-backend-mysql` パッケージが必要です
            * pdnutilに必要な設定は `docker-compose-go.yml` を確認してください

これらミドルウェア群は、CI上以下のdocker-composeによって起動されるようになっています

https://github.com/isucon/isucon13/blob/main/development/docker-compose-common.yml

また、ユーザのプロフィール管理上、画像を扱います

* webappは、ユーザのアイコン取得、アイコンアップロード機能を提供します
    * アイコン取得はBlobを返します
    * アップロードでは、JSONの中に `image` という画像をbase64した値を入れたkvを含めてリクエストします
    * 一度もアップロードされてないユーザの画像取得を行った場合、フォールバックに用いられる画像 `NoImage.jpg` が返されます
        * この画像はwebappのワーキングディレクトリに配置される必要があります


まとめると、CIでのWebapp動作には以下のファイルが必要です

* init.sh
    * SQLファイルを実行し、DBを初期化します
    * webappの `POST /api/initialize` 処理中に実行されます
    * レコード登録数が多く、体感でわかるほど時間がかかります
        * 支障が出ないよう、過度にならないように今後調整が入る可能性があります
* SQLファイル
    * 初期データのINSERT文が入っている
    * webapp/init.sql、webapp/initial_xxx.sqlがこれに対応
* ダミーのpdnsutil
* フォールバック画像 (NoImage.jpg)

## self-hosted runnerについて

2023/11/4 時点で、5台のself-hosted runnerがさくらのクラウド上で動作しています。

https://scrapbox.io/isucon13operation/isucon13_CI%E7%92%B0%E5%A2%83

なるべく衝突が起きないように台数動かしていますが、あまり多すぎても管理時にこまるので程々の台数に抑えています

ただ、不手際で２台ほど不足することに気づきましたので、週明け用意予定です

`.github/workflows/xxx.yml` を定義することで、指定したイベントがトリガーされるとActionが実行されるようになります

基本的な書き方は以下を踏襲しつつ、Goなど言語指定部分を各言語に置き換えていただくこととなります

https://github.com/isucon/isucon13/blob/main/.github/workflows/go.yml

各言語ごとのサンプル実装として以下も参考になりますので、ご参照ください

https://github.com/isucon/isucon12-qualify/tree/main/.github_/workflows

## Makefileについて

以下３つのタスクが定義されています

1. make `up/<言語>`
    * docker compose upを行います
    * `docker-compose-common.yml` の考慮もされています
2. make `down/<言語>`
    * docker compose downを行います
3. make bench
    * ベンチマークを実行します

## 進め方

### CI実行環境用意

まずは、割り当てられたself-hosted runnerにSSHログインできることを確認してください。事前にGitHub.comのsshkeysを登録していますが、これ以外の公開鍵登録が必要な場合は Slackで作問メンバーにメンションいただけると助かります。

webappのDockerイメージをビルドします。`webapp/<各言語>` 配下にDockerfileを作成してビルド、動作確認をしてください。

Dockerイメージをビルドできるようになったら、`docker-compose-<言語>.yml` の準備をします。
`development` ディレクトリ配下に `docker-compose-<言語>.yml` を作成してください。
ミドルウェアの起動は `docker-compose-common.yml` で定義済みですので、言語ごとの `docker-compose-<言語>.yml` で指定しないでください。
また、注意点として、self-hosted runnerはインターネット上から到達性のあるサーバであり、ポートを開放してしまうと意図せぬアクセスを受ける可能性があります。そのため、ポートフォワーディング指定では必ず127.0.0.1を指定してリッスンするようにしてください。

ここまでできれば、webappやミドルウェアのコンテナを立ち上げ、ベンチマーカーから接続できるようになるので、 `.github/workflows/<言語>.yml` を作成してください。
基本的には、checkout後にコンテナ群を立ち上げ、ベンチマーカーを起動する流れになります。お好みで、以下のようにAction実行後にお掃除をすることもできます

https://github.com/isucon/isucon12-qualify/blob/main/.github_/workflows/go.yml#L53,L57

ただ、問題を調査したい場合はコンテナを立ち上げっぱなしにした上で、self-hosted runner上でデバッグができますので、適宜判断していただくようお願いいたします。


## 移植進行サイクル

### ベンチマーカーについて

ベンチ実行すると、以下のようなログがActionに記録されます

```
...
2023-11-03T13:47:13.295Z	info	isupipe-benchmarker	webapp: http://pipe.u.isucon.dev:8080
2023-11-03T13:47:13.296Z	info	isupipe-benchmarker	===== Prepare benchmarker =====
2023-11-03T13:47:13.296Z	info	isupipe-benchmarker	webappの初期化を行います
2023-11-03T13:47:48.123Z	info	isupipe-benchmarker	ベンチマーク走行前のデータ整合性チェックを行います
2023-11-03T13:47:48.539Z	info	isupipe-benchmarker	ベンチマーク走行を開始します
2023-11-03T13:47:48.539Z	info	isupipe-benchmarker	負荷レベル: 10
2023-11-03T13:48:48.539Z	info	isupipe-benchmarker	===== ベンチ走行中エラー =====
2023-11-03T13:48:48.540Z	warn	isupipe-benchmarker	[リクエストタイムアウト] POST /api/register: Post "http://pipe.u.isucon.dev:8080/api/register": context deadline exceeded (Client.Timeout exceeded while awaiting headers): タイムアウトによりリクエスト失敗
2023-11-03T13:48:48.540Z	info	isupipe-benchmarker	===== ベンチ走行結果 =====
2023-11-03T13:48:48.540Z	info	isupipe-benchmarker	売上: 27052
```

ここでは、以下の確認が必要となります

* 整合性チェックがエラーになっていないこと
    * 現状実装不足で、今後ここの改善にとりかかっていきます
* 走行中エラーで不審なものがないこと
    * context deadlineが発生しているのは、APIリクエストに時間がかかりすぎる場合です
    * 初期実装の時点で負荷をかけると登録処理でタイムアウトするので、このような出力になっています
    * これ以外のエラーが出ていたり、あまりにも多くのエンドポイントでcontext canceledが発生する場合は移植できていない可能性があります
* 売上が0以上になっており、25000 ~ 35000になっていること
    * 今後スコア周りの修正を加えるのでスコア変動しますが、2023/11/4時点ではこの範囲のスコアに収まります
    * 特に0になっている場合はなにかしら異常が起きている可能性が高いです

### サイクルを回す

基本的なサイクルは以下のようになります。

1. webappに変更を加える
2. pushし、Actionの結果を待つ (initializeが遅く、3minくらいかかります. うち、ベンチ走行は1minです)
3. 問題が起きた場合、self-hosted runnerにSSHログインし、以下のようなデバッグ
    * docker psでコンテナID確認
    * docker logsでログ確認
    * docker execでコンテナ内に入り、init.shを実行して確認
    * ホスト上で `mysql -uisucon -pisucon isupipe` を実行してMySQLシェルを開き、登録データの調査
    * ホスト上で `mysql -uisucon -pisucon isudns` を実行してMySQLシェルを開き、サブドメインの調査
4. 原因が分かり次第1に戻る


## 参考

2023/11/4時点でまだ整備中です :bow:

移植の際に参考にしていただければと考えておりますので、随時共有させていただいたタイミングで確認いただけると助かります。

* [アプリケーション仕様](https://github.com/isucon/isucon13/blob/main/docs/specification.md)
* [用語集](https://github.com/isucon/isucon13/blob/main/docs/terminology.md)