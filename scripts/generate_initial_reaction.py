import random

N = 100000


SQL_FORMAT = "\t('{emoji_name}', 1, 1),\n"

with open('./initial-data/emoji.txt', 'r') as f:
    emoji_list = list(line.rstrip() for line in f.readlines())


with open('/tmp/reactions.sql', 'w') as f:
    f.write('INSERT INTO reactions (emoji_name, user_id, livestream_id)\n')
    f.write('VALUES\n')
    for _ in range(N):
        emoji_name = random.choice(emoji_list)
        sql = SQL_FORMAT.format(emoji_name=emoji_name)
        f.write(sql)
    f.write("\t('+1', 1, 1);\n")
