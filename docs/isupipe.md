
# ISUPipe アプリケーションマニュアル

[ISUCON13 当日マニュアル](https://gist.github.com/kazeburo/bccc2d2b2b9dc307b5640ae855f3e0bf) も合わせて確認してください。

## ISUPipeとは

こんいす〜。

オンラインでのやりとりが普及した現代においては、時代に即した交流が求められています。

ライブ配信はリアルタイムに体験を共有し、個人のキャラクターを多くの人々に届ける強力な手段です。

ISUPipeでは配信者ごとのドメインを活用した柔軟なテーマ設定が可能です。また、ライブ配信や配信者を特徴づける統計情報、ライブ配信に対する「チップ」投稿などリアルタイムな体験を支援する機能が盛りだくさんです。

最初は視聴者から始めて、慣れてきたら配信に挑戦することも可能です。あなたもこの新しい体験をしてみませんか？

## 用語

* ユーザー
  * 配信者(streamer): ライブ配信を行う配信者です。
  * 視聴者(viewer): ライブ配信を視聴する視聴者です。

* ライブストリーム(Livestream): ライブ配信です。また、予約期間を保持するため、予約としての機能も兼ねています。
  * 同一時間帯の予約には枠数制限があります。枠数制限を超えて予約することはできません。
  * 予約可能な期間には限りがあります。現在2023/11/25 10:00:00 JST ~ 2024/11/25 10:00:00 JSTの期間中の予約を受け付けています。

* ライブコメント(Livecomment): ライブ配信に対するリアルタイムなコメントです。
* 投げ銭(Tip): ライブ配信、ライブ配信者に対する送金・寄付です。

* リアクション(Reaction): ライブ配信に対するリアクションです。 `:innocent:` や `:+1:` などフロントエンドで絵文字として解釈されて表示されます。

* 配信者ごとのドメイン: ライブ配信者が個々に所有するサブドメインです。このドメインに合わせてフロントエンドの見た目が変わります。
  * ISUPipeでは、この機能を実現するためDNSサーバーを運用しています
  * ライブ配信関連機能へのアクセス時、このドメインが用いられます。
    * それ以外の場合、 `pipe.u.isucon.dev` というISUPipeが所有する特別なドメインにアクセスします

* 長時間配信者: 10時間以上のライブ配信を行う配信者です。
* 通常配信者: 10時間未満のライブ配信を行う配信者です。

* ユーザ配信統計情報: ユーザが投稿したライブコメント、投げ銭、リアクションに関する統計情報です。
* ライブ配信統計情報: ライブ配信に投稿されたライブコメント、投げ銭、リアクションに関する統計情報です。

* スパム
  * 配信に対してネガティブなコメントのことを指します。配信者にとっては迷惑行為であり、対処していくことで配信品質を向上させることが視聴者の良い評価につながります。

* スパム報告
  * 視聴者はスパムに対してスパム報告を行うことができます。
  * 即座にスパムに対処される保証はありませんが、配信者にこの報告が届くので、配信の改善に貢献できます。
  * 虚偽の報告を行ったユーザがいると判明した場合、ISUPipeからは然るべき対応をさせていただく場合がございます。
    * アカウントのBANなど、重いペナルティが課せられます

* モデレーション
  * 配信者は、スパムに対してモデレーションを行うことができます。
  * モデレーションを行うと、過去の対象となるスパムと、新たに投稿されたスパムを排除することができます。
  * 視聴者からのスパム報告も参考になるでしょう。

* 課金(Payment): ISUPipeで稼いだ総チップ金額です。

* ISUCOIN: ISUPipeにおいては日本円とはことなる独自の通貨を利用します。ISUCOIN(ISUと表記されることもあります)と呼ばれるもので、ISUPipeではISUCOINの売上が評価対象となります。

## 運営から提供される環境

* HLS配信サーバー (media.xiii.isucon.dev)
  * フロントエンドでの表示上必要なHLS配信機能が提供されます。
  * また、配信のサムネイルを取得することが可能です。
  * 本サーバーについてはチューニング対象ではありません。

## アイコン画像配信についての特記事項

ISUPipeのクライアント(ベンチマーカーを含む)は、`/api/user/:username/icon`というURLに対してGETリクエストを行い、ユーザーのアイコン画像を取得します。

このリクエストでは、他のAPIで返されるユーザー情報に含まれる `icon_hash` という値を利用し、HTTPリクエストヘッダを次のように付与して条件付きGETリクエストを行うことがあります。

```
GET /api/user/:username/icon HTTP/1.1
Host: pipe.u.isucon.dev
If-None-Match: "{icon_hashの値}"
(他のヘッダは省略)
```

ISUPipeのサーバーはこのリクエストに対し、ユーザーのアイコン画像のSHA256値と送信された `icon_hash`　の値を比較して、一致する場合にはステータスコード 304 でレスポンスすることができます(MAY)。

条件付きGETリクエストでない場合、ステータスコード 304 をレスポンスしてはいけません(MUST NOT)。

アイコン画像が更新され、送信された `icon_hash` の値と一致しなくなった場合には、ステータスコード200でレスポンスし画像データを送信しなければなりません(MUST)。

APIで返却される `icon_hash` の値と `GET /api/user/:username/icon` で配信されるアイコン画像は、ユーザーのアイコン画像が更新された後、2秒以内に変更が反映されている必要があります(MUST)。

## ユーザーのパスワードについての特記事項

ISUPipeのサーバーでは、ユーザーのパスワードをbcryptでハッシュ化して保存しています。

パスワードのハッシュアルゴリズムを変更したり、bcryptのコストを変更してはいけません(MUST NOT)。

## 一部ログについてについての確認事項

この「一般エラー」はベンチに問題となり、減点対象ではありません

```
viewer_spam: failed to post livecomment (moderated spam): benchmark-application: [一般エラー] POST /api/livestream/7531/livecomment へのリクエストに対して、期待されたHTTPステータスコードが確認できませんでした (expected:400, actual:201)
```

## Node.js初期実装へのパッチ

Node.js の初期実装には負荷走行が不正終了することがあり、以下のパッチを適用してください

```
diff --git a/webapp/node/src/handlers/stats-handler.ts b/webapp/node/src/handlers/stats-handler.ts
index 0428e9db..10b14b90 100644
--- a/webapp/node/src/handlers/stats-handler.ts
+++ b/webapp/node/src/handlers/stats-handler.ts
@@ -55,7 +55,9 @@ export const getUserStatisticsHandler = [
           .catch(throwErrorWith('failed to count reactions'))
 
         const [[{ 'IFNULL(SUM(l2.tip), 0)': tips }]] = await conn
-          .query<({ 'IFNULL(SUM(l2.tip), 0)': number } & RowDataPacket)[]>(
+          .query<
+            ({ 'IFNULL(SUM(l2.tip), 0)': string | number } & RowDataPacket)[]
+          >(
             `
               SELECT IFNULL(SUM(l2.tip), 0) FROM users u
               INNER JOIN livestreams l ON l.user_id = u.id	
@@ -68,7 +70,7 @@ export const getUserStatisticsHandler = [
 
         ranking.push({
           username: user.name,
-          score: reaction + tips,
+          score: reaction + Number(tips),
         })
       }
 
@@ -219,7 +221,9 @@ export const getLivestreamStatisticsHandler = [
           .catch(throwErrorWith('failed to count reactions'))
 
         const [[{ 'IFNULL(SUM(l2.tip), 0)': totalTip }]] = await conn
-          .query<({ 'IFNULL(SUM(l2.tip), 0)': number } & RowDataPacket)[]>(
+          .query<
+            ({ 'IFNULL(SUM(l2.tip), 0)': number | string } & RowDataPacket)[]
+          >(
             'SELECT IFNULL(SUM(l2.tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l.id = l2.livestream_id WHERE l.id = ?',
             [livestream.id],
           )
@@ -228,7 +232,7 @@ export const getLivestreamStatisticsHandler = [
         ranking.push({
           livestreamId: livestream.id,
           title: livestream.title,
-          score: reactionCount + totalTip,
+          score: reactionCount + Number(totalTip),
         })
       }
       ranking.sort((a, b) => {
```

