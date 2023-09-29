import argparse
from datetime import datetime, timedelta
import random
import time

from faker import Faker


fake = Faker('ja-JP')


# 投入するSQL文のフォーマット
SQL_FORMAT="INSERT INTO livestreams (user_id, title, description, start_at, end_at) VALUES ({user_id}, '{title}', '{description}', '{start_at}', '{end_at}');"

# FIXME: Goコードのフォーマット
GO_SCHEDULER_PATTERN_FORMAT="""
package scheduler

var {livestream_schedule_pattern_name} = []LivestreamPattern{{
{schedules}
}}
"""

GO_SCHEDULER_PATTERN_UNIT_FORMAT="    newLivestreamPattern({user_id}, \"{title}\", \"{description}\", \"{start_at}\", \"{end_at}\"),\n"

# 時間枠のパターン(hourの粒度)
patterns = [1,3,5,10,20,24]

def solve(total, limits, assigned, schedules, max_schedules=10):
    """スケジュール候補を生成します"""
    def dec_limit(limits, idx):
        """ストック制限数から、指定idxのパターンを消費します"""
        new_limits = limits[:]
        new_limits[idx] -= 1
        return new_limits

    if total < 0:
        return schedules
    if sum(limits) == 0:
        return schedules
    if total == 0:
        return schedules + [assigned]

    for idx, pattern in enumerate(patterns):
        if limits[idx] <= 0:
            continue
        if len(schedules) >= max_schedules:
            return schedules

        new_schedules = solve(total-pattern, dec_limit(limits, idx), assigned[:]+[pattern], schedules)
        if len(new_schedules) > len(schedules):
            schedules = new_schedules

    return schedules


def generate_season1(max_schedules, max_users):
    # NOTE: season1は、初心者救済期間
    #       可能な限り多くの予約を取れるようにし、たくさんライブコメント、投げ銭、リアクションを稼げるようにしてあげる
    #       あまりにもスコアが上がりすぎる場合は大型枠を増やして調整

    # NOTE:１日引いている(exclusiveな期間)
    available_hours = (24*(30+31+30)) - (24*1)
    # 各時間枠のストック (雑設定。生成できるだけあればいい)
    limits = [400,40,40,40,40,40]

    schedules = solve(available_hours, limits, [], [], max_schedules)
    dump_sql(available_hours, schedules, max_schedules, max_users)


def generate_season2(max_schedules, max_users):
    # NOTE: season2は、新人VTuberからの予約が殺到する
    # 新人VTuber宛のライブコメント、投げ銭、リアクションはそこまで多くないが、予約が多い期間
    # 様々な予約パターンを試せると良さそうなので、一旦均等に分配する

    available_hours = (24*(31+31+30)) - (24*1)
    limits = [100,100,100,100,100,100]

    schedules = solve(available_hours, limits, [], [], max_schedules)
    base_time = datetime(2024, 7, 1, 0)
    dump_go_scheduler(base_time, available_hours, schedules, max_schedules, max_users)


def generate_season3(max_schedules, max_users):
    # NOTE: season3は、新人VTuberが成長する期間。season2ではリクエストが少なかったVTuberにもリクエストが飛ぶようになる
    # 投げ銭やライブコメント、リアクションがたくさんくるので、人気VTuberだけ優遇するアプローチ一本ではうまくいかない
    # このとき、予約枠の大きさは大きな配信枠が多めになり、衝突が増えてくる

    available_hours = (24*(31+30+31)) - (24*1)
    limits = [40,40,40,140,140,200]

    schedules = solve(available_hours, limits, [], [], max_schedules)
    dump_sql(available_hours, schedules, max_schedules, max_users)


def generate_season4(max_schedules, max_users):
    # NOTE: season4は、まんべんなくリクエストがくる。広告費用係数に応じて負荷レベルがここから先際限なく上昇していく
    # 予約は大きめな枠がおおくなる
    available_hours = (24*(31+28+31)) - (24*1)
    limits = [40,40,40,140,140,200]

    schedules = solve(available_hours, limits, [], [], max_schedules)
    dump_sql(available_hours, schedules, max_schedules, max_users)

def create_timeslice(base_time, hours):
    """渡されたbase_timeを基点として、指定時間の枠を表現するタイムスライスを作成します"""
    delta_hours = timedelta(hours=hours)
    return (base_time, base_time+delta_hours,)

def dump_go_scheduler(base_time, available_hours, schedules, max_schedules, max_users):
    for schedule in schedules[:max_schedules]:
        assert(sum(schedule) == available_hours)

        go_schedule_units = []

        # Go生成
        for hours in random.sample(schedule, len(schedule)):
            start_at, end_at = create_timeslice(base_time, hours)
            go_schedule_unit = GO_SCHEDULER_PATTERN_UNIT_FORMAT.format(
                user_id=random.randint(1, max_users),
                title=''.join(fake.random_letters()),
                description=fake.text().replace('\n', ''),
                start_at=start_at.strftime("%Y-%m-%d %H:%M:%S"),
                end_at=end_at.strftime("%Y-%m-%d %H:%M:%S"),
            )
            go_schedule_units.append(go_schedule_unit)
            base_time = end_at

        print(GO_SCHEDULER_PATTERN_FORMAT.format(
            livestream_schedule_pattern_name="abc",
            schedules=''.join(go_schedule_units),
        ))




def dump_sql(available_hours, schedules, max_schedules, max_users):
    for schedule in schedules[:max_schedules]:
        assert(sum(schedule) == available_hours)

        # SQL生成
        # 2024/4/1から開始
        base_time = datetime(2024, 4, 1, 0)
        for hours in random.sample(schedule, len(schedule)):
            start_at, end_at = create_timeslice(base_time, hours)
            sql = SQL_FORMAT.format(
                user_id=random.randint(1, max_users),
                title=''.join(fake.random_letters()),
                description=fake.text().replace('\n', ''),
                start_at=start_at.strftime("%Y-%m-%d %H:%M:%S"),
                end_at=end_at.strftime("%Y-%m-%d %H:%M:%S"),
            )
            print(sql)
            base_time = end_at


def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('--users', type=int, default=10, help='対象ユーザ数')
    # max_schedulesは生成するスケジュールの最大数。枠数と一致する
    # NOTE: 枠数を設定する場合、枠数だけスケジュールを重ねがけする
    #       すべてきれいに敷き詰めるように書いているので、単純に複数のスケジュール分SQL生成すればよい
    parser.add_argument('-n', type=int, default=1, help='生成スケジュール数(=同一時間帯の枠数)')
    # Goコードを生成する
    parser.add_argument('--go', action='store_true', help='season2,3,4のスケジュールパターンのGoコードを生成')

    return parser.parse_args()


def main():
    args = get_args()
    assert(args.users > 0)
    # NOTE: 枠数は10000ぐらいが生成限度.
    assert(args.n < 10000)

    # スケジュール生成
    if not args.go:
        print("# Season1")
        generate_season1(args.n, args.users)
    else:
        # FIXME: goコードを生成する実装
        print("// Code generated by scripts/generate_livestream_reservation.py; DO NOT EDIT.")
        generate_season2(args.n, args.users)
        # print("// Code generated by scripts/generate_livestream_reservation.py; DO NOT EDIT.")
        # generate_season3(args.n, args.users)
        # print("// Code generated by scripts/generate_livestream_reservation.py; DO NOT EDIT.")
        # generate_season4(args.n, args.users)


if __name__ == '__main__':
    main()
