import argparse
from enum import Enum
from faker import Faker
import sys
import subprocess
import pprint

# NOTE: bcrypt-toolを使ったパスワードハッシュ生成が非常に遅いので、事前に不足しない程度に生成してある
passwords = [('wd44Ivk041', '$2a$10$gjaotHZoumKefYz8oUlqCefwbx4IrsUzZ10RDyIWFGtbGkWAHV0xi'),
 ('8rJDRjD1R8', '$2a$10$lqLqAQHiIHYMkIREWFGJUeTDQHPsvyOYV.O/bltC137DJ.HjQ25we'),
 ('IHU8Kytw9v', '$2a$10$TyLaDfioqHiEvmIfnK9LnOuIxyiSG4zE4q5cwQW3arjpUrKnuRBvm'),
 ('l971DAoDB1', '$2a$10$zwGxkLE5WYeQrikiYxgk5etg8CU0.BKBmput4XuG6.r7Q7nkBFzFy'),
 ('8OZkvnvJDx', '$2a$10$2XhabwzviUjs0HjvnsVsDOp/d9ajhxeXChk/BqXhTTY.43EI15zJy'),
 ('xlbvLqjoL5', '$2a$10$ljdf6sPQB1WLi00etu9nau7/ePqJPVOGKIBn/xHEEuvCgjaKR0XFW'),
 ('v67NdpYSr7', '$2a$10$ihinBGEg8LUfFs2GXfJT8Op2k2j4fN8MrThVeTsa5KXRiAapVIPSG'),
 ('idYNLPnoo4', '$2a$10$Tn7UUWbJazRTVBffn.xrxuViXcNgWgzc.G62sWafw8FFf4KtcsGsG'),
 ('T8siQjvjXE', '$2a$10$PDoDxFbal/THC.JLuJDAXebpwdY1laJ5FDNX7OisYei3XEvEGIEWO'),
 ('s5EPffuPk8', '$2a$10$xmpR3FmhuzT6gLGnLvY25ey7fGeFbi0royNW6Bq6OLn3MLy.84RAq'),
 ('3cQeJqT4SN', '$2a$10$PVGej/JNk6T5CCKIkR1QkOJgZ/ysOem218fE1zcvAdRgXkmX1EJ8m'),
 ('2vXlfoR5XW', '$2a$10$fOmB7HAll0DVwNHdcHRDYOIF796ItZaPoarpLQZ7wuCEJOUrKHYGu'),
 ('1iH2sTDw7j', '$2a$10$X7M2fYcEmAoEtwNgvWcL1uGusQ18b1T4Y3yjOkqSAGVAFfZCU1.Ky'),
 ('N8LQtc1tUs', '$2a$10$I6ihAqeGHrYh1tlizNcLq.SyUZJ.2NB7XpRjVzcjJo1BzGzE2oalS'),
 ('7lHt28hoGx', '$2a$10$UG09WLt0k80UKUDtYbkbo.VhYgLBtA4jGzpGG3RPB4SIXBtWvfNdq'),
 ('uMixfaLE8F', '$2a$10$pVnVhN91b6Xurg06C3f7zORYs16TNKMijB8lDeHsp1HgQoMssTh4m'),
 ('P11I0HKuiK', '$2a$10$i4rGLRqBs4J.ZO5EKnEDJOBCTOmeiHyZz8RcX9KWpbG06OqJBaHwq'),
 ('3FfAQdsq19', '$2a$10$n4UzUYaXnnTOQsMDMaBfke5kJCFKpufvgZB2TgSdo07HSQJFlCKm6'),
 ('9UAqle5SUf', '$2a$10$U0CQpFEeJQ4GbUURM.LQo.SJzWx3ykVgVma4p9AAVN2xQO4STBDLa'),
 ('ZxAQpcc056', '$2a$10$.9o9APBSmIgPBdCGsyWFUOtei8u8KipTiOf8D4SoiL0vvwDyTUnGm'),
 ('LJWtpDsSL5', '$2a$10$FCqdrBfQZuVe.HKN6q448eJF8uPlfzSE83LHC8Rz9bAc388mPJZcu'),
 ('E9xpRrpPDU', '$2a$10$f8G8NBZWyoSmTzXBZylbvOgY1LMKtMGuL6jhO4zoy81QV9FB5tgd6'),
 ('0QK0Ysd47h', '$2a$10$o25anRQnmnMbgi9z3Fs3ruCwLiSL9P4ytksMiU9KCPGcZVZ/AkGjm'),
 ('0Vn6KsvK78', '$2a$10$f91JS4d9zkFs58Em7wQFg.IfJRRhzfn/g.fX8AQLEmP2dpurDDP6.'),
 ('184WHHKtLY', '$2a$10$pT41458JnuCwWNrCRtUO2.UJ2wtdzXRsBarquucxEZRbdO.4eN3vS'),
 ('3Ov6NwlnxY', '$2a$10$yhDh4MvuqJHFLTY0f1jXVOUCEBQUIMC9FlcMFcqWW4gn0NIFF7Rui'),
 ('H3HaXAgUtn', '$2a$10$gdJlG2.i5MbNxkYLGGxn3e7mm5vhBHTVZdqMKIhNbAHdoKJ8Ub5IW'),
 ('gSBKA2Fu1b', '$2a$10$Y68ncfdPmZ6lLvw4HoolXeNbdM2071SZuNYXuvTgRqehvJBLcn3Mq'),
 ('9c8S8vU9Vy', '$2a$10$RG/V1jBRKaOHDJwW.s6cSujNEeq2/58vWSzt5bcBIZ3oK9gBGnDwS'),
 ('85LnFPKhxJ', '$2a$10$iBU4vxtQTg4d.NAStK1MH.Oymg6pBIfXg0F8wcqZL/pXDXrydydkG'),
 ('pQJvfo1B48', '$2a$10$2lp.9qdFiefFLBlbHKkrqefbftVgW2R5bSt0yi6Jqm6BTg8BlwvDG'),
 ('7VU4pZgEzO', '$2a$10$SNu1lK2q1ddYmE9piGAAD.r1sntx1nA1N8Gi9EO7nYixFoB53idDW'),
 ('Q8FORKhBW6', '$2a$10$wBkAIoiijPyRU/VYf5CGw.TedsG0ZDqBp6eWto6kp0CDcJzt4xBA2'),
 ('7x6UHJXxSw', '$2a$10$dJQX6Y9TFYr/clYlGc7KL.HWmaf/25Gv90FXwvOf3C/2mBu6f.2CW'),
 ('FRgj4Azrbd', '$2a$10$i/9mdUfLmSpZEF1Qr3HsyeZkFzj2w8LR.F2CvogztSk1DGOecI8Oq'),
 ('1IacWKlmcA', '$2a$10$8kBSlcJvfxAdkp15FJPPuOeKR4KIVUsQS60sbd.mNcmN89thNABVm'),
 ('4bmxpN9xrP', '$2a$10$sAsWQ0Uc/Yz8PxA2QGif/uQiS3r8TDArtuwW0tz.KmPGIxfZLRLv6'),
 ('0R5KdTU6yZ', '$2a$10$hDkN1JXb3JONDJ/xRvmVzuY0pv6YsYBPJdnuGS8xSpzCqUeUfO1vC'),
 ('x6BAll0qP3', '$2a$10$fER61B1tmQ8nCOF/5g8z9eP4HJBu.pjdgZhqj11tmdJVy4lkmQzyi'),
 ('XjbBhzgh1d', '$2a$10$JK43I4xnygLgyrtuotZ2u.XZ3sFDEmJoxY.eTqDDXj9NYRQur/DNm'),
 ('FKAZBubn6V', '$2a$10$xIsZWWGEEZUpy7zW0LvEJegtdWs3f1UukEDviB1UerIF3ffDczZIa'),
 ('OiSUwfEE24', '$2a$10$IwsKq1Z3jBi7/m739ReoWe77UR8TvEG/VCoVk7.ZJE5HS/SGZ1GIe'),
 ('9j7WzTbRoV', '$2a$10$we52LKLUxtwof.bdGVbnyeM92zaxkOYNqxn2gIFD85hoadgMrZnbO'),
 ('pR5HyGfr7T', '$2a$10$FkAcqDUHwIoDIRsuEXVON.Az6IwN/uBbl8a/CBSyF8uY1.ENk3I5K'),
 ('qZT6L6GdLN', '$2a$10$i0M1gA6hSsD1wc8CGrae3uWmmwb9wAV3qjMRJadw9gqyeEjUxopEq'),
 ('sPeVNS7D1Z', '$2a$10$8rwGwLa6l4V.VXYVsM9FDuJqPPGTqTTgbKZIS4ltH4dGKd7a7qIru'),
 ('sbAi9iyBU4', '$2a$10$mjutEV91hAlmEB0tqi03V.xBHFNWKU.eCAeUADrTdV6muq/Oq6uXy'),
 ('VMDAFmZd1v', '$2a$10$s5zN4IkB1WrI28Gjf1Q7Vuex3gPpUu6CvC5kgNJWMPBVYZJtZbKLy'),
 ('n9HYRaoQo5', '$2a$10$t5K310LasESOU1N4b1BpqOpRJW8oX8Pi4mwgWkPZnylz5zCpqQq9G'),
 ('QY37W4XlBn', '$2a$10$qXDd80VLLuoRt4MufCGzGOV3Six.jNAvSrlura0GJVRdyEv5EeV/e'),
 ('2AQvSPElnb', '$2a$10$kIUTF2NKYDELpqItNyUQAOQgiZS4ILpUSP7QcFFkKfS8y.YlEW4V6'),
 ('iiryAfDw9Q', '$2a$10$FiyB09ggT/MctT7rVFNzfeWz54stAU.AB53PU186UbvjhhsFoudBW'),
 ('P18BQm0dIo', '$2a$10$xSMauwAGkC368qeSc.UmfOL1m6ye4KPi7xcW0uDekhIylKW9GTQP6'),
 ('uewD93MZ2M', '$2a$10$2PRH3pYMlozvYXP5774f8OmlIEqtNBTGbRE12tFeTfvL5pdH7i2zS'),
 ('8RcfvVBpyt', '$2a$10$aMY6pSdUc7UYvn/nSW0PHefr2O/hbFROliFZI2Iqgw5DfOdekf15y'),
 ('W7TF8c3Qz8', '$2a$10$Dqk5aiJkf3K/AdSbsgKfU.kA2ICwxzyk88FG1Ki77nKl51HBRRt8a'),
 ('oTZOlIJf8T', '$2a$10$K0psmr6Wh7Pn2p5OOLWEG.oX61.2QMhNu0iiSRY.BZ.Nhtd2apivO'),
 ('fy6TWJklF3', '$2a$10$49jVEPoyhyRyYuiy2LkjNOnspWuznUd/somH5InlbPay/Cx8n9eWm'),
 ('ZLqeGvSb6B', '$2a$10$oCoDWivX7DQ5bFbz/rB8UeVtl4LLcoDjU08Cr29HDGH/ROx2JOpFe'),
 ('nuhQrGOm6Z', '$2a$10$wVlv7OSHqJjK1V8G.Q8wm.Khcsn710IJgIOQq4XxF1P5XfeoYm0im'),
 ('ZT3mG7KnU5', '$2a$10$4KzLpRoSbgjnDJajU.royeV9FIX6Th07L8jKY.UDkvpJPmbSektri'),
 ('QVi47gEhBZ', '$2a$10$OOu5hMNSUompXtXa34eqKeZnFeIZcXyByXF8.6QdQ5xbUX1wQHIrG'),
 ('9WKKiFvuxg', '$2a$10$l6LIPpYsOH3MEeef62Mcd.eSkKQGyk7ha3NXmEjwybHjGA21uWLYW'),
 ('4Ol1N34wz0', '$2a$10$xf6/qgIj4YkNHJiOwe6FtemxIh/Xy/4HnhrKzsXgNOG7VQoqwls/6'),
 ('5vnw0T9kqY', '$2a$10$OwjPOnokDZoxZWoMJm83Su2JeCcOzc/SYN7uUXSmIoPcgjYWB9aEm'),
 ('ej70aGcbMU', '$2a$10$cyhZdUt.ks8JX6f1IKCf8uXsSXSliSA6zSVX91F638hZjR3xwz0YW'),
 ('Dhy9YqIp12', '$2a$10$dwfFDQOt0ZAyprhbXu/CCOAJPOdM09VY/hIgS29zEM2pBCO6ZYpae'),
 ('70RfwGmNQD', '$2a$10$B5AbwLSGVi8ACCYePhtvwOiHWSmbpeUcmnblykU.iGfHraYPBPbvW'),
 ('3KKozcpcqK', '$2a$10$kBDsTgFXJOWh6yoNAlZt6OsFSOqgwxO/kwqLUWbut2hVHgiEcUyIe'),
 ('e7lB4Se0gS', '$2a$10$JgmW2VpWmYh3Y1MSsoNYK.o6lBG4uo9OTmI76HE2qhqMloCyzJyEW'),
 ('7JZpGWUasj', '$2a$10$hzM/pHNCLlEB5cmBY/rk4u/IuA3ud3714zp3vAiWcmjovBX1PsWFy'),
 ('WfExQFmL5A', '$2a$10$8db79EdFD/7Fsj0Rpc2T6e68h9tjQHN.v12Vy4BVvt/MU1EyRZROq'),
 ('8j5sqOKfEv', '$2a$10$xdO/hxMrnNslDi85zTvMo.7hkHZt9Cbx2qSouz4y7um/LUhxNgf2e'),
 ('enkH5dOmGY', '$2a$10$FldtJJOY8y9GenY2AjUizuLqxBPlKviVjSiAf0KWG9zzNChuD8xxm'),
 ('3varSzdL1I', '$2a$10$DJGtDHUyV8qyzKkyOVB9bOgN6guSy2JNHCTk956oL.XUqgmFiHUz2'),
 ('51Z9wEBEks', '$2a$10$3vaD2w3NbE3G1lmPhQ/V5e3mbflj7WVlfhMBCElc5d6RIFH6UbEbO'),
 ('tRJXVXczU2', '$2a$10$5c.DSlirIzOqFcyYjmE42.Rg6BVIh/qv3Kb0UevAsFGtXr986oI9e'),
 ('OXiq5aGcSL', '$2a$10$j8OELyLq3bM7i.k0JNQ3AO8ga6sXZn08TwPM1.N7sbsdxvdSZ6YlG'),
 ('lNy6jVpff8', '$2a$10$PU9j0uj.YF3r3/O77VBWfu3klTfLM29uazXigX4L4N8uxvAIsTTsG'),
 ('pb46sSvkUh', '$2a$10$fZaIwhe6/i/nAVfjMRkGju7cBLBGkhQuMuuNvYKV2G4wCejSBgDte'),
 ('23LXGvgSdd', '$2a$10$46MQmNWQqbjrhBY0cojiH.rE6VzrSject0T/wi93gBPRcHE3Rc7hq'),
 ('BDLWzg6e0o', '$2a$10$5/sVSqHLq0n/uVrXHzveSelj6Ibkr4KfJT9VhqK1ekW6P/t48D0aW'),
 ('2kvHYfOv2I', '$2a$10$2PME/EzNNg46CBuAKULeLuvllmX9.k490bfuJ.kwS.4DLXSm.jFKC'),
 ('W3UD9VhhJv', '$2a$10$To0yPabYydufDgkrhOPfX.jOUJR7zHhN8ZM8DG9xZnaDDwsPi1.Mu'),
 ('WpzBSeyRI7', '$2a$10$cF52a819m.UyVzhFJtgqgOtaZ0kMxgqRFKYC9qX6WPnNSM9mwtFva'),
 ('1OXMz8Ni9y', '$2a$10$vw9DluQOYtVR/1qGWcNEBu7xsEm8TWjSstLRDUc3wOt6H3QdY9l.q'),
 ('G33JMhAhWP', '$2a$10$GvQtDC0gBNjxUoe/vrPRbOWjJZLuWjXR9J1P/zZd5GNfbnF0uS2cS'),
 ('IT40qwBydR', '$2a$10$iq7x2S7XilCjlsj2slDmieTSmmhNGyieEQfVrNr3x1H2Dx2im9vHK'),
 ('50lVdXTReX', '$2a$10$rvIZZoPpL42P.PikqA98Y.RUznbvIGC5gVvUDRQfIiF/b7IxIHyTW'),
 ('E0oBVvgzIU', '$2a$10$6kkfQuc4ulsQqbRuW4bDzupCDYLenlS09GeU/s53WpHqy73v3k1b.'),
 ('8XcRtInnCV', '$2a$10$r62r1OnYRYXT1SeygeyGPOWSlj9KWgpM3LmdetZTZH7K8P3TNLSFi'),
 ('563AqyfjFk', '$2a$10$b7z10rWKPkg/ga.b652fxuj1cSKxQ0p6VeTVcv4NnZ1dPO11yvLNy'),
 ('X9MIXZVnUb', '$2a$10$SbS/6V7T2ch6BuWxZF29Juzpf72.vPNEOny2ni5rfI.7x1g1YqaUy'),
 ('xLlzYIiT52', '$2a$10$X6HMzuRe/QURXjluyd41Tup4yHFcbUf9FK3AnsAWF57iks3z4xDeC'),
 ('7dw8vmaijR', '$2a$10$nuOx3nFQQtG5nxAo9kJIWeORojL7gcHDS6KWqdN6w34ivzVw1cKua'),
 ('uUmpUh5b06', '$2a$10$T294VCeUIyXrBDNGqeDZceAhZ6LLbo7nHCJBbueahwESfo7FWN9aC'),
 ('6rx3MsWpjA', '$2a$10$Rzsp8nDezrRADQIkMBFF5.dpeoG5kZwjHRFWjZGxW5ovuDEE6k.Zm'),
 ('xdD3MjwsEU', '$2a$10$gAI6Wcq7mKSXy3sFzk5afelzZCV9ezuDrJIfstOZp.P7oNBJGU8JO'),
 ('za8DDED1n4', '$2a$10$w6DufZf17XMnjSkpzOjsXug0HzgUSWsqfLdzyCDoWOAmIpjRhpKh6'),
 ('6LrTeZLbu6', '$2a$10$tiHELF/Ycc98w8DfXGrQneik73LuLQg6FvYCRu3hsVPHMxpS5m/6u')]


# SQL
SQL_FORMAT="INSERT INTO users (id, name, display_name, description, password) VALUES ({user_id}, '{name}', '{display_name}', '{description}', '{password}');"
INSERT_THEME_FORMAT="INSERT INTO themes (user_id, dark_mode) VALUES ({user_id}, {dark_mode});"

# Go
GO_CONSTRUCTOR_FORMAT="&User{{ UserId: {user_id}, Name: \"{name}\", DisplayName: \"{display_name}\", Description: \"{description}\", RawPassword: \"{raw_password}\", HashedPassword: \"{hashed_password}\"}},"

#
DESCRIPTION_FORMAT="普段{job}をしています。\\nよろしくおねがいします！\\n\\n連絡は以下からお願いします。\\n\\nウェブサイト: {website}\\nメールアドレス: {mail}\\n"

fake = Faker('ja-JP')


def hash_password(raw_password):
    result = subprocess.run([
        'bcrypt-tool', 'hash', raw_password
    ], encoding='utf-8', stdout=subprocess.PIPE)
    hashed_password = result.stdout.rstrip('\n')
    return hashed_password


def format_sql(users):
    output = ""
    for user in users:
        hashed_password = user['hashed_password']
        sql = SQL_FORMAT.format(
            user_id=user['user_id'],
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
    mid = len(users)//2

    output = "package scheduler\n"

    # 配信者
    output += "var vtuberPool = []*User{\n"
    for user in users[:mid]:
        ctor = GO_CONSTRUCTOR_FORMAT.format(
            user_id = user['user_id'],
            name = user['name'],
            display_name = user['display_name'],
            description = user['description'],
            raw_password = user['raw_password'],
            hashed_password = user['hashed_password'],
        )
        output += f'\t{ctor}\n'
    output += "}\n"

    # 視聴者
    output += "var viewerPool = []*User{\n"
    for user in users[mid:]:
        ctor = GO_CONSTRUCTOR_FORMAT.format(
            user_id = user['user_id'],
            name = user['name'],
            display_name = user['display_name'],
            description = user['description'],
            raw_password = user['raw_password'],
            hashed_password = user['hashed_password'],
        )
        output += f'\t{ctor}\n'
    output += "}\n"

    return output

def gen_user_theme_sql(user_id: int) -> str:
    dark_mode = ['true', 'false'][user_id % 2]
    return INSERT_THEME_FORMAT.format(**locals())


def gen_user(n: int) -> str:
    users = []
    password_cache = dict()
    for user_id in range(1, n+1):
        profile = fake.profile()
        name = profile['username']
        if name in password_cache:
            # 名前が再利用される場合、そのパスワードも再利用
            raw_password, hashed_password = password_cache[name]
        else:
            # 名前が新規に振られる場合、パスワードを作ってキャッシュに入れておく
            raw_password, hashed_password = passwords[(user_id-1)%100]
            password_cache[name] = (raw_password, hashed_password,)

        users.append(dict(
            user_id = user_id,
            name = profile['username'],
            display_name = profile['name'],
            description = DESCRIPTION_FORMAT.format(
                job=profile['job'],
                website=profile['website'][0],
                mail=profile['mail'],
            ),
            raw_password = raw_password,
            hashed_password = hashed_password,
        ))

    return format_sql(users) + '\n\n\n' + format_go(users)


def dump_passwords():
    passwords = []
    for _ in range(100):
        raw_password = fake.password(special_chars=False)
        hashed_password = hash_password(raw_password)
        passwords.append((raw_password, hashed_password, ))
    pprint.pprint(passwords)


def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('-n', type=int, default=10, help='生成数')
    return parser.parse_args()


def main():
    args = get_args()
    output = gen_user(args.n)
    print(output)


if __name__ == '__main__':
    main()
    # dump_passwords()


