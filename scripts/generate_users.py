import argparse
from faker import Faker
import sys
import subprocess

SQL_FORMAT="INSERT INTO users (name, display_name, description, password) VALUES ('{name}', '{display_name}', '{description}', '{password}');"
# SUPERCHAT_SQL_FORMAT="INSERT INTO superchats (user_id, livestream_id, comment, tip) VALUES (:user_id, :livestream_id, :comment, :tip)"

DESCRIPTION_FORMAT="普段{job}をしています。\\nよろしくおねがいします！\\n\\n連絡は以下からお願いします。\\n\\nウェブサイト: {website}\\nメールアドレス: {mail}\\n"

fake = Faker('ja-JP')


def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('-n', type=int, default=10, help='生成数')

    return parser.parse_args()

def gen_user_sql():
    profile = fake.profile()
    name = profile['name']
    display_name = profile['username']
    description = DESCRIPTION_FORMAT.format(
        job=profile['job'],
        website=profile['website'][0],
        mail=profile['mail'],
    ) 
    non_hashed_password = fake.password()
    print(non_hashed_password, file=sys.stderr)
    result = subprocess.run(['bcrypt-tool', 'hash', non_hashed_password], encoding='utf-8', stdout=subprocess.PIPE)
    password = result.stdout

    return SQL_FORMAT.format(**locals())

# def gen_superchat_sql():
    # return fake.text()

def main():
    args = get_args()
    for _ in range(args.n):
        sql = gen_user_sql()
        print(sql)

    # for _ in range(args.n):
        # print(gen_superchat_sql())

if __name__ == '__main__':
    main()
