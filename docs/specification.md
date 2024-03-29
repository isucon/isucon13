# ISUPipe仕様


ISUPipeは、2024/04/01 ~ 2025/03/31の1年間のロードマップを発表する

## できること


* ユーザ操作
    * 登録
    * ログイン
    * ユーザプロフィール取得

* 配信視聴
    * 視聴開始、視聴終了 (フロントエンドのmountedなどで自動処理)
    * 配信の視聴
        * 配信環境は運営側で用意 (フロントエンドの見た目だけ)
        * ライブコメント投稿
            * Tipが設定されると投げ銭として扱われる
        * ライブコメントのスパム報告
        * 



## 広告費用



## ベンチマークからの負荷



## ボトルネック要素

* 配信予約
* ライブコメント投稿 (投げ銭も兼ねる)

## 配信予約

* 配信を開始するためには、予約をしなければならない。
* 予約は最小単位で1hであり、ベンチマーカーはこれに従う
* 予約の枠は限られており、同一時間帯に予約可能である配信枠は限定される
    * おおむね、人気VTuberの予約を優先した方が投げ銭を稼げる
    * ただ、固定的に人気VTuberの予約を優先するだけでは対応できない
    * 時期によっては登録したばかりのVTuberが人気を獲得する場合があり、ベンチ走行中刻一刻と状況が変化する
    * 時期によって、新人VTuberを応援しようと視聴者が殺到する時期がある
        * 従って、全体の時期を一貫して人気VTuberだけ優先するわけにはいかない

## 投げ銭

投げ銭はライブコメントの構造と全く同じとなる。違いは

ライブコメントの場合、Tip=0である
投げ銭の場合、Tip>0である

という一点に尽きる

## スパム報告

ISUPipeでは、治安維持のため、配信者が個人個人でスパム対応できる仕組みを導入している

* 視聴者は、特定の配信の特定のライブコメントについてスパム報告をすることが可能である
* 配信者は、視聴者からのスパム報告を閲覧することが可能である。この投稿を参考にして対処が可能である
* また、配信者はもでレーションを行うことができ、スパム対策のNGワード登録が可能である
    * ISUPipeは、NGワードに基づき、新規ライブコメント投稿を自動フィルタリングする
    * NGワードは、技術的な用語がベースとなる。プログラミング言語の予約語、技術用語などを用いる
    * ベースとなる文章はChatGPTである程度生成し、そこにNGワードを埋め込むことでベンチマーカーがライブコメント投稿を行う
    * 特定の配信にスパムが投稿できた数が多いほど、投げ銭の数が少なくなる、金額が減るようにベンチマーカーを調整する
    * 


# 配信予約

## ベンチマーカーが予約する時間の粒度について

最小単位は1h。

のちほどこれでは不足するようならより細かい粒度も検討する

# チップの概念について

isupipeでは、行われている配信に対してライブコメントを投稿することができ、
ライブコメントにはチップ(tips)が含まれます。
これは、一般的に投げ銭システムと呼ばれるものです。

isupipeでは、このtipsについて、以下のようなレベル分けを行っております。

|金額帯|色|
|:--:|:--:|
|1 ~ 499| 青 |
|500 ~ 999 | 緑 |
|1000 ~ 4999| 黄色 |
|5000 ~ 9999 | オレンジ |
|10000 ~ 20000 | 赤 |

