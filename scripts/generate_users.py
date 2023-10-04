import argparse
from enum import Enum
from faker import Faker
import sys
import subprocess

# SQL
SQL_FORMAT="INSERT INTO users (name, display_name, description, password) VALUES ('{name}', '{display_name}', '{description}', '{password}');"
INSERT_THEME_FORMAT="INSERT INTO themes (user_id, dark_mode) VALUES ({user_id}, {dark_mode});"

# Go
GO_CONSTRUCTOR_FORMAT="newUser({user_id}, '{name}', '{display_name}', '{description}', '{raw_password}', '{hashed_password}')"

#
DESCRIPTION_FORMAT="普段{job}をしています。\\nよろしくおねがいします！\\n\\n連絡は以下からお願いします。\\n\\nウェブサイト: {website}\\nメールアドレス: {mail}\\n"

fake = Faker('ja-JP')


class OutputFormat(Enum):
    SQL = "sql"
    Go = "go"

    def __str__(self):
        return self.value

def hash_password(raw_password):
    result = subprocess.run([
        'bcrypt-tool', 'hash', raw_password
    ], encoding='utf-8', stdout=subprocess.PIPE)
    hashed_password = result.stdout.rstrip('\n')
    return hashed_password


def format_sql(users):
    output = ""
    for user in users:
        hashed_password = hash_password(user['raw_password'])
        sql = SQL_FORMAT.format(
            name=user['name'],
            display_name=user['display_name'],
            description=user['description'],
            password=hashed_password,
        )
        output += f'{sql}\n'

        theme_sql = gen_user_theme_sql(user['user_id'])
        output += f'{theme_sql}\n'

    return output


def format_go(users):
    output = "package scheduler\n"
    output += "var phase2UserPool = []*User{\n"
    for user in users:
        hashed_password = hash_password(user['raw_password'])
        ctor = GO_CONSTRUCTOR_FORMAT.format(
            user_id = user['user_id'],
            name = user['name'],
            display_name = user['display_name'],
            description = user['description'],
            raw_password = user['raw_password'],
            hashed_password = hashed_password,
        )
        output += f'\t{ctor}\n'
    output += "}\n"

    return output

def gen_user_theme_sql(user_id: int) -> str:
    dark_mode = ['true', 'false'][user_id % 2]
    return INSERT_THEME_FORMAT.format(**locals())


def gen_user(n: int, format: OutputFormat) -> str:
    users = []
    for user_id in range(1, n+1):
        profile = fake.profile()
        users.append(dict(
            user_id = user_id,
            name = profile['username'],
            display_name = profile['name'],
            description = DESCRIPTION_FORMAT.format(
                job=profile['job'],
                website=profile['website'][0],
                mail=profile['mail'],
            ),
            raw_password = fake.password(),
        ))

    if format == OutputFormat.SQL:
        return format_sql(users)
    elif format == OutputFormat.Go:
        return format_go(users)
    else:
        raise ValueError(f'unsupported output format {format}')


def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('-n', type=int, default=10, help='生成数')
    parser.add_argument('--format', type=str, choices=list(OutputFormat), default=OutputFormat.SQL, help='出力フォーマット(sql, go)')
    return parser.parse_args()


def main():
    args = get_args()
    output = gen_user(args.n, args.format)
    print(output)


if __name__ == '__main__':
    main()
