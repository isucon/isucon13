import argparse
from collections import defaultdict
from datetime import datetime, timedelta, timezone
import json
import random
import sys
sys.setrecursionlimit(3000)

from faker import Faker
import requests


fake = Faker('ja-JP')

# FIXME: livestream_texts、もうちょっと増やしたい

# FIXME: id, usernameを外す
GO_FORMAT="\tmustNewReservation({id}, \"{title}\", \"{description}\", \"{start_at}\", \"{end_at}\", \"{playlist_url}\", \"{thumbnail_url}\"),\n"

lives = []

def fetch_live(stream_id: int):
    num_samples = 10

    global lives
    if len(lives) > num_samples:
        return random.choice(lives)

    url = "https://media.xiii.isucon.dev/api/{stream_id}/live/".format(stream_id=stream_id)
    resp = requests.get(url)
    resp.raise_for_status()
    j = json.loads(resp.text)

    live = (j['playlist_url'], j['thumbnail_url'], )
    lives.append(live)
    return live

livestream_texts = [
("夜のゲーム実況！新作RPG攻略中","新しいRPGゲームを実況しながら進めていきます！みんなでワイワイ攻略しよう！",),
("早起き朝活配信！今日も元気に","朝の時間を有効に使いたい皆さん、一緒に活動を始めましょう！",),
("歌ってみたライブ！リクエスト募集中","皆さんからのリクエスト曲を歌っていきます！楽しんでくださいね♪",),
("料理ライブ！今夜の晩御飯は…","手軽にできる料理を実際に作りながら紹介していきます！",),
("映画レビュー！最新映画の感想を語る","最新の映画について、感想や考察を深く掘り下げて話していきます。",),
("旅行記！先日の旅行の思い出話","先日の旅行での出来事やおすすめスポットを紹介します！",),
("新曲初披露！歌配信","オリジナルの新曲を初めて披露します！感想をお聞かせください！",),
("アートライブ！絵を描きながらのんびりトーク","絵を描きながら、リラックスしたトークを楽しんでください。",),
("最新ガジェット紹介！これは買い？","最新のテクノロジーガジェットを紹介していきます。",),
("プログラミングライブ！コードを書きながら学ぼう","初心者向けに、実際にコードを書きながらプログラミングを学んでいきます。",),
("健康トーク！毎日のワークアウト方法","健康的な生活のためのワークアウト方法や食事について話していきます。",),
("手相占いライブ！あなたの運命は？","手相をもとに、あなたの未来や運命を占います。",),
("アニメトーク！最新アニメの感想交換","最新のアニメについて、感想や考察をみんなで交換しましょう！",),
("コメディライブ！笑いでリフレッシュ","日常の出来事やニュースをユーモアたっぷりに解説します！笑顔で過ごしましょう！",),
("ビジネスライブ！起業の秘訣","実際に起業を経験したゲストと、成功の秘訣や失敗談について語り合います。",),
("ファッションライブ！今季のトレンドは？","最新のファッショントレンドやおすすめアイテムを紹介します！",),
("ペットと一緒の夜ふかし配信","愛犬と一緒にのんびりとトークを楽しむ時間。ペットの話題もお待ちしています！",),
("心理学トーク！あなたの心を読み解く","日常生活での心理学の活用方法や面白い心理テストを紹介します。",),
("宇宙トーク！宇宙の不思議を探る","宇宙や天体に関する最新の情報や知識を深く掘り下げて話していきます。",),
("歴史トーク！過去の出来事から学ぶ","世界の歴史や有名な出来事について、詳しく解説します。",),
("手芸ライブ！今日は何を作ろう","手芸好きの皆さんと一緒に、新しい作品を作っていきます！",),
("音楽制作ライブ！新曲の制作過程を公開","新しい楽曲の制作過程をリアルタイムで公開します。",),
("Q&Aセッション！視聴者の質問に答えます","視聴者の皆さんからの質問に、直接答えていきます！",),
("ガーデニングライブ！今日のお花はこれ","庭やベランダでのガーデニングの様子を公開します。",),
("子育てトーク！育児の悩み解決方法","子育て中のママやパパのための、育児の悩みや解決方法について話し合います。",),
("料理ライブ！手軽に作れるデザート特集","家にある材料で簡単に作れるデザートレシピを紹介します！",),
("スポーツ実況！今夜の試合はこれ！","今夜のスポーツ試合の実況と解説を行います！",),
("読書の時間！おすすめの一冊を紹介","今月読んだ本の中から特に感動した一冊を紹介します。",),
("日本語レッスン！基本の文法から","日本語を学びたい方のための基本的な文法レッスンを行います。",),
("釣りライブ！海辺でのんびり","今日は海辺での釣りの様子を配信します。どんな魚が釣れるかな？",),
("お料理対決！視聴者投票で勝者を決めよう","2人の料理人が対決！視聴者の皆さんの投票で、どちらの料理が上手いかを決めます。",),
("ホラートーク！怖い話を共有しよう","夜の時間、一緒に怖い話を共有する時間です。",),
("実況！今夜のサッカー試合","今夜のサッカー試合の実況と解説を行います！",),
("アイドルとのコラボ配信！特別な時間","人気アイドルとの特別なコラボ配信！裏話やQ&Aも！",),
("映画の裏話！監督とのスペシャルトーク","大ヒット映画の裏話を、監督自らが語ります。",),
("ヨガライブ！朝のリフレッシュタイム","朝の時間をヨガでリフレッシュ！初心者も大歓迎！",),
("子猫の生活！成長記録を公開","我が家の子猫の成長記録を公開します。",),
("英語トーク！日常英会話を学ぼう","日常で使える基本的な英会話フレーズを紹介します。",),
("夜の音楽ライブ！リラックスタイム","夜の時間を、心地良い音楽とともに過ごしましょう。",),
("天体観測！今夜の星空を一緒に","都市の喧騒を離れ、美しい星空を観測します。",),
("ダンスライブ！今夜の特訓の様子を公開","ダンスの特訓の様子をリアルタイムで公開します。",),
("ゲーム実況！新作アクションゲーム","新作のアクションゲームの実況プレイを行います！",),
("資格取得の勉強方法！効率的な学習法を伝授","様々な資格試験に合格するための勉強方法を伝授します。",),
("キャンプライブ！自然の中での生活を公開","キャンプ場からのライブ！自然の中での生活や料理を公開します。",),
("コスメ紹介！この春のおすすめアイテム","この春の新作コスメやおすすめのアイテムを紹介します。",),
("絵画の制作過程！アートライブ","実際の絵画の制作過程をリアルタイムで公開します。",),
("日常の出来事トーク！今日のハイライト","今日あった日常の出来事やハイライトを話していきます。",),
("健康ライブ！美容と健康の秘訣","美容と健康のための食事やエクササイズの秘訣を伝授します。",),
("旅行先からのライブ！今日の景色を共有","今日訪れた旅行先の美しい景色や文化を共有します。",),
("歴史の謎を解明！歴史探訪ライブ","世界各地の歴史的な場所を訪れ、その謎や背景を解明します。",),
("実況！今夜のバスケットボール試合","今夜のバスケットボール試合の実況と解説を行います！",),
("恋愛トーク！恋愛相談に答えます","恋愛の悩みや相談に、経験豊富なゲストと一緒に答えていきます。",),
("家の中の大掃除！一緒に片付けよう","家の中を一緒に大掃除！片付けのコツやおすすめのアイテムを紹介します。",),
("ビューティーライブ！スキンケアの秘訣","肌に優しいスキンケア方法やおすすめのアイテムを紹介します。",),
("実況！今夜のボクシング試合","今夜のボクシング試合の実況と解説を行います！",),
("山登りライブ！絶景の山頂から","実際に山を登りながら、その絶景をリアルタイムで公開します。",),
("日常のDIY！家具の組み立て","新しい家具の組み立ての様子やDIYのコツを紹介します。",),
("実況！今夜のテニス試合","今夜のテニス試合の実況と解説を行います！",),
("ファッションライブ！夏のおすすめコーディネート","この夏のおすすめのファッションコーディネートを紹介します。",),
("料理ライブ！子供も喜ぶレシピ","子供が喜ぶ簡単で美味しい料理レシピを紹介します。",),
("実況！今夜の格闘技大会","今夜の格闘技大会の実況と解説を行います！",),
("ガーデニングライブ！春の花を植えよう","春の花の植え付けやガーデニングのコツを紹介します。",),
("読書会ライブ！今月のおすすめ本を読み進める","一緒に今月のおすすめの本を読み進め、その感想を共有します。",),
("音楽制作ライブ！楽曲のアレンジ方法","実際の音楽制作の過程を公開し、楽曲のアレンジ方法などを紹介します。",),
("実況！今夜のバレーボール試合","今夜のバレーボール試合の実況と解説を行います！",),
("ペットとの日常！うちの犬の生活を公開","我が家の愛犬の日常や生活の様子を公開します。",),
("DIYライブ！手作りのアクセサリー作り","実際に手作りのアクセサリーを作りながら、その方法やコツを紹介します。",),
("美容ライブ！最新の美容機器を試してみる","最新の美容機器を実際に試し、その効果や感想を公開します。",),
("子供向けライブ！絵本の読み聞かせ","子供向けの絵本を読み聞かせる時間。就寝前のリラックスタイムに。",),
("実況！今夜のプロレス試合","今夜のプロレス試合の実況と解説を行います！",),
("トレーニングライブ！筋トレの方法を紹介","筋トレの基本的な方法やコツを紹介しながら、実際にトレーニングを行います。",),
("写真術ライブ！基本の撮影技術を学ぶ","プロのカメラマンが、写真の基本的な撮影技術やコツを伝授します。",),
("野菜の収穫ライブ！畑での作業を公開","畑での野菜の収穫や作業の様子をリアルタイムで公開します。",),
("お菓子作りライブ！手作りクッキーのレシピ","一緒に美味しい手作りクッキーを作る時間。簡単なレシピを紹介します。",),
("マジックライブ！手品の秘密を少しだけ教えます","マジシャンが実際の手品を披露し、その秘密を少しだけ公開します。",),
("資産運用の基本！投資初心者のためのライブ","資産運用や投資の基本的な知識を、専門家がわかりやすく説明します。",),
("アウトドアライブ！BBQの楽しみ方を紹介","アウトドア好きが集まるBBQの様子を公開し、楽しみ方やおすすめの食材を紹介します。",),
("映画クリティック！新作映画のレビュータイム","映画評論家が新作映画の内容や見どころを詳しく説明します。",),
("手芸ライブ！かわいい刺繍の作り方","手芸好きのためのライブ。今回はかわいい刺繍の作り方を紹介します。",),
("車のメンテナンスライブ！基本の作業を紹介","車の基本的なメンテナンス方法や作業のポイントを専門家が解説します。",),
("美容室ライブ！春の新作ヘアスタイルを公開","美容師が春の新作ヘアスタイルやカラーを実際に公開します。",),
("実況！今夜のゴルフトーナメント","ゴルフのトーナメントの進行や選手の動向をリアルタイムで実況します。",),
("絵描きライブ！風景画の描き方を学ぶ","プロの画家が、風景画の基本的な描き方や技術を伝授します。",),
("子育てライブ！子供の健康や教育について","子育ての専門家が、子供の健康や教育に関する悩みや質問に答えます。",),
("ライブコンサート！アコースティックセッション","アーティストがアコースティックセッションを公開。心温まる音楽の時間です。",),
("歴史講座ライブ！戦国時代の英雄たち","歴史学者が、戦国時代の英雄たちやその背景を詳しく解説します",),
]

## 同一枠長の採用数制限
SAME_PATTERN_LIMIT = 40
# 時間枠のパターン(hourの粒度)
patterns = list(n for n in range(1, 25))

def solve(total, counters, assigned, schedules):
    """スケジュール候補を生成します"""
    if total < 0: # overflow
        return schedules
    if total == 0:
        return schedules + [assigned]

    for pattern in patterns:
        if counters[pattern] >= SAME_PATTERN_LIMIT:
            continue

        counters[pattern] += 1
        new_schedules = solve(total-pattern, counters, assigned+[pattern], schedules)
        if len(new_schedules) > len(schedules):
            schedules = new_schedules
        else:
            counters[pattern] -= 1

    return schedules


def dump(base_time, schedules, gopath):
    """生成されたスケジュールをSQLファイル、Goファイルに出力します"""
    def create_timeslice(base_time, hours):
        """渡されたbase_timeを基点として、指定時間の枠を表現するタイムスライスを作成します"""
        delta_hours = timedelta(hours=hours)
        return (base_time, base_time+delta_hours,)

    go_schedules = []
    schedule_idx = 1
    for schedule in schedules:
        cursor_time = base_time
        for hours in random.sample(schedule, len(schedule)): # スケジュール要素をシャッフル
            title, description = livestream_texts[schedule_idx%len(livestream_texts)]
            start_at, end_at = create_timeslice(cursor_time, hours)
            playlist_url, thumbnail_url = fetch_live(schedule_idx)

            go_schedules.append(GO_FORMAT.format(
                id=schedule_idx,
                title=title,
                description=description,
                start_at=start_at.strftime("%Y-%m-%d %H:%M:%S"),
                end_at=end_at.strftime("%Y-%m-%d %H:%M:%S"),
                playlist_url=playlist_url,
                thumbnail_url=thumbnail_url,
            ))
            print(end_at - start_at)

            cursor_time = end_at
            schedule_idx += 1

    with open(gopath, 'w') as gofile:
        gofile.write('package scheduler\n\n')
        gofile.write('var reservationPool = []*Reservation{\n')
        for go_schedule in go_schedules:
            gofile.write(go_schedule)
        gofile.write('}')
        gofile.write('\n')


# 初期予約３ヶ月分生成用
def generate_initial(gopath, num_workload_schedules, num_slots=5, num_users=1000):
    base_time = datetime(2023, 8, 1, 1, tzinfo=timezone.utc)
    total_hours = (31+30+31)*24

    livestream_id = 1
    values = []
    schedule_idx = num_workload_schedules + 1
    go_schedules = []
    foreign_keys = []
    while total_hours > 0:
        hours=random.randint(1, 2)
        delta = timedelta(hours=hours)
        start_at = base_time
        start_unix = int(start_at.timestamp()) 
        end_at = start_at + delta
        end_unix = int(end_at.timestamp())

        for slot_idx in range(num_slots):
            user_id = 1 + (livestream_id % (num_users-1))
            assert user_id != 0, f"livestream_id = {livestream_id}, num_users={num_users}, value = {livestream_id % (num_users-1)}"
            title, description = livestream_texts[(user_id+slot_idx)%len(livestream_texts)]
            playlist_url, thumbnail_url = fetch_live((user_id+slot_idx))
            values.append(f'\t({user_id}, \"{title}\", \"{description}\", \"{playlist_url}\", \"{thumbnail_url}\", {start_unix}, {end_unix})')
            go_schedules.append(GO_FORMAT.format(
                id=schedule_idx,
                title=title,
                description=description,
                start_at=start_at.strftime("%Y-%m-%d %H:%M:%S"),
                end_at=end_at.strftime("%Y-%m-%d %H:%M:%S"),
                playlist_url=playlist_url,
                thumbnail_url=thumbnail_url,
            ))
            foreign_keys.append((user_id, livestream_id, ))

            livestream_id += 1
            schedule_idx += 1

        total_hours -= hours
        base_time = end_at

    with open('/tmp/initial_livestream.sql', 'w') as f:
        f.write('INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at)\n')
        f.write('VALUES\n')
        for idx, value in enumerate(values):
            f.write(value)
            if idx == len(values)-1:
                f.write(';\n')
            else:
                f.write(',\n')

    with open(gopath, 'a') as gofile:
        gofile.write('var initialReservationPool = []*Reservation{\n')
        for go_schedule in go_schedules:
            gofile.write(go_schedule)
        gofile.write('}')
        gofile.write('\n')

    with open('./initial-data/autogenerated_user_livestream_foreignkey_pairs.json', 'w') as f:
        f.write(json.dumps(foreign_keys))

def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('--users', type=int, default=1000, help='対象ユーザ数')
    # max_schedulesは生成するスケジュールの最大数。枠数と一致する
    # NOTE: 枠数を設定する場合、枠数だけスケジュールを重ねがけする
    #       すべてきれいに敷き詰めるように書いているので、単純に複数のスケジュール分SQL生成すればよい
    parser.add_argument('-n', type=int, default=5, help='生成スケジュール数(=同一時間帯の枠数)')
    parser.add_argument('--gopath', type=str, default='/tmp/livestream.go')

    return parser.parse_args()


# NOTE: 実行時間は2, 3秒程度
def main():
    args = get_args()
    assert(args.users > 0)
    # NOTE: 枠数は4000ぐらいが生成限度.
    assert(args.n <= 4000)

    available_days = 365
    available_hours = (24*available_days)-1

    # スケジュール生成
    schedules = solve(available_hours, defaultdict(int), [], [])

    # スケジュール選択
    # 生成されたスケジュールを一定条件でソートし、好ましいものの上位n件を結果として出力
    schedules = schedules[:args.n]
    assert(len(schedules) == args.n)
    assert(all(sum(schedule) == available_hours for schedule in schedules))

    base_time = datetime(2023, 11, 25, 1, tzinfo=timezone.utc)
    dump(base_time, schedules, args.gopath)

    # NOTE: generate_initialはgoファイルに追記するので、先にdumpを実行しておく必要がある
    generate_initial(args.gopath, sum(len(schedule) for schedule in schedules))

if __name__ == '__main__':
    main()
