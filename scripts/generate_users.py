import argparse
from faker import Faker
import sys
import subprocess

SQL_FORMAT="INSERT INTO users (name, display_name, description, password) VALUES ('{name}', '{display_name}', '{description}', '{password}');"
INSERT_THEME_FORMAT="INSERT INTO themes (user_id, dark_mode) VALUES ({user_id}, {dark_mode});"
# livecomment_SQL_FORMAT="INSERT INTO livecomments (user_id, livestream_id, comment, tip) VALUES (:user_id, :livestream_id, :comment, :tip)"

DESCRIPTION_FORMAT="普段{job}をしています。\\nよろしくおねがいします！\\n\\n連絡は以下からお願いします。\\n\\nウェブサイト: {website}\\nメールアドレス: {mail}\\n"

fake = Faker('ja-JP')


def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('-n', type=int, default=10, help='生成数')

    return parser.parse_args()

def gen_user_theme_sql(user_id: int) -> str:
    dark_mode = ['true', 'false'][user_id % 2]
    return INSERT_THEME_FORMAT.format(**locals())

def gen_user_sql(user_id: int) -> str:
    profile = fake.profile()
    display_name = profile['name'] # Fakerではja-JP設定だとname=氏名になるので、これをdisplay_nameとして扱う
    name = profile['username']
    description = DESCRIPTION_FORMAT.format(
        job=profile['job'],
        website=profile['website'][0],
        mail=profile['mail'],
    ) 
    non_hashed_password = fake.password()
    print(non_hashed_password, file=sys.stderr)
    result = subprocess.run(['bcrypt-tool', 'hash', non_hashed_password], encoding='utf-8', stdout=subprocess.PIPE)
    password = result.stdout.rstrip("\n")

    insert_user_sql = SQL_FORMAT.format(**locals())
    return insert_user_sql + "\n" + gen_user_theme_sql(user_id)

# def gen_livecomment_sql():
    # return fake.text()

def main():
    args = get_args()
    for i in range(args.n):
        sql = gen_user_sql(i + 1)
        print(sql)

    # for _ in range(args.n):
        # print(gen_livecomment_sql())

if __name__ == '__main__':
    main()
