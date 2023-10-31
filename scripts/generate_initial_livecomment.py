import random

N = 100000

SQL_FORMAT = "\t(1, 1, '{comment}'),\n"



with open('./initial-data/positive_sentence.txt', 'r') as f:
    livecomments = list(line.rstrip() for line in f.readlines())

with open('/tmp/livecomments.sql', 'w') as f:
    f.write("INSERT INTO livecomments (user_id, livestream_id, comment)\n")
    f.write("VALUES\n")
    for _ in range(N):
        comment = random.choice(livecomments)
        sql = SQL_FORMAT.format(comment=comment)
        f.write(sql)
    f.write("\t(1, 1, 'こんにちは');\n")

