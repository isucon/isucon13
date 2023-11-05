SQL_FORMAT="\t(1, 1, '{word}', UNIX_TIMESTAMP()),\n"

with open('./initial-data/initial_ngwords.txt', 'r') as f:
    ngwords = list(line.rstrip() for line in f.readlines())

with open('/tmp/ngwords.sql', 'w') as f:
    f.write('INSERT INTO ng_words (user_id, livestream_id, word, created_at)\n')
    f.write('VALUES\n')
    for ngword in ngwords:
        sql = SQL_FORMAT.format(word=ngword)
        f.write(sql)
    f.write("\t(1, 1, '椅子パイプ', UNIX_TIMESTAMP());\n")
