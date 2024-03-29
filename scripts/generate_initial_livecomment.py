import random
import json

N = 1000

SQL_FORMAT = "\t({user_id}, {livestream_id}, '{comment}', UNIX_TIMESTAMP()),\n"
GO_FORMAT="\t&InitialLivecomment{{ UserID: {user_id}, LivestreamID: {livestream_id}, Comment: \"{comment}\" }},\n"

with open('./initial-data/autogenerated_user_livestream_foreignkey_pairs.json', 'r') as f:
    fk_pairs = json.loads(f.read().rstrip())

with open('./initial-data/positive_sentence.txt', 'r') as f:
    livecomments = list(line.rstrip() for line in f.readlines())

sql_values = []
go_ctors = []
for _ in range(N):
    user_id, livestream_id = random.choice(fk_pairs)
    comment = random.choice(livecomments)
    sql = SQL_FORMAT.format(comment=comment, user_id=user_id, livestream_id=livestream_id)
    sql_values.append(sql)
    go_ctor = GO_FORMAT.format(comment=comment, user_id=user_id, livestream_id=livestream_id)
    go_ctors.append(go_ctor)
user_id, livestream_id = random.choice(fk_pairs)
sql_values.append(f"\t({user_id}, {livestream_id}, 'こんにちは', UNIX_TIMESTAMP());\n")
go_ctors.append(f"\t&InitialLivecomment{{ UserID: {user_id}, LivestreamID: {livestream_id}, Comment: \"こんにちは\" }},\n")

with open('/tmp/livecomments.sql', 'w') as f:
    f.write("INSERT INTO livecomments (user_id, livestream_id, comment, created_at)\n")
    f.write("VALUES\n")
    for sql_value in sql_values:
        f.write(sql_value)

with open('/tmp/livecomments.go', 'w') as f:
    f.write('package scheduler\n\n')
    f.write('var initialLivecommentPool = []*InitialLivecomment{\n')
    for go_ctor in go_ctors:
        f.write(go_ctor)
    f.write('}\n')
