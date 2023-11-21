from datetime import datetime, timedelta, timezone

# 予約枠生成

NUM_SLOTS = 5
SQL_FORMAT="\t({slot}, {start_at}, {end_at})"

# delta = timedelta(hours=1)
base_time = datetime(2023, 11, 25, 1, tzinfo=timezone.utc)
total_hours = (24*365)-1

with open('/tmp/reservation_slot.sql', 'w') as f:
    f.write('INSERT INTO reservation_slots (slot, start_at, end_at)\n')
    f.write('VALUES\n')
    for idx, delta_hour in enumerate(range(total_hours)):
        start_delta = timedelta(hours=delta_hour)
        start_at = base_time + start_delta
        end_delta = timedelta(hours=delta_hour+1)
        end_at = base_time + end_delta

        sql = SQL_FORMAT.format(slot=NUM_SLOTS, start_at=int(start_at.timestamp()), end_at=int(end_at.timestamp()))
        print(f'start_at={start_at.isoformat()}', f'end_at={end_at.isoformat()}')
        if idx == total_hours-1:
            f.write(sql + ';\n')
        else:
            f.write(sql + ',\n')



