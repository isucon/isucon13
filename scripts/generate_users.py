import argparse
from collections import defaultdict
from enum import Enum
from faker import Faker
import sys
import subprocess
import pprint
import random

# NOTE: bcrypt-toolを使ったパスワードハッシュ生成が非常に遅いので、事前に不足しない程度に生成してある
passwords = [('jGa8b0W85B', '$2a$04$0BRufcJy.4FPaIDjWxsKz.5JBQQMoGZzVniBFU0s7jaC.8pHSiuaK'),
 ('VK0bYU5qrx', '$2a$04$GvL4Wts8RiOULOE9qAYZWeX6pqOaW9F3MoWrqiAtMt3K.tepN85Xq'),
 ('5RVJlhuS5i', '$2a$04$OHdW6thFVfbi6tPhy5lXGeiFAShOFDz0xFVG2iSjM0yZ.jnAa0Uqq'),
 ('31QFwk5psg', '$2a$04$0PulOoA5M9qCiCU4oL0O1OqnrGD./.dGLQPwuTf9SfioljG1ImG7.'),
 ('KwyQ3Wqs2F', '$2a$04$lBKwG4y37PBniflzaf/iY.9KYWo2ZBd6JSfvgT9UG0XA80uSShAJW'),
 ('7wI6Epahes', '$2a$04$pEqzGw.XCz0voYRH3QMJTO59V1WsTacBbGb/K2UETzzAEqaz6xU96'),
 ('WuMaea0Ef9', '$2a$04$j./lZgNj/fSYZqg7O7BDue3T7Ijeun6VVUizLRGFMheTUMIqcYK2y'),
 ('DrWWKwxx1P', '$2a$04$sLLGF.m9gbfbm5pTeUjpvuW89BW.yNEylciEbeGzsacMmbCN1DnSC'),
 ('NoRidWjt81', '$2a$04$uu1XhQxhkzIJsJ7Tzw8qfuH/wivo4hkEYkFKOP0xKulBk2nxZPEqS'),
 ('RBu6Il1u7O', '$2a$04$qUPj6ngqW37MzPM4LsJ7QOBsEW7qF0fUVyE4T3QJYH4AKVozbNC9q'),
 ('1JsCqfob2V', '$2a$04$ah/ZrA65mvjSG9M2zRSUcegW1d1d.lzyMkKl0leUmnkmTEBGogPp.'),
 ('9PBW7gAD2t', '$2a$04$sdAySrC/p9LWk5J.r5LM.udKeDM02c0eSXn3GFCrJ0uMUOmDlQDy2'),
 ('xVKjUfZqQ1', '$2a$04$uHf5EFSsbhKGqYandZv/Z.kwJ.UHjKBL5qD9hr8us.PEoIZTxNQrq'),
 ('14obMlDhCG', '$2a$04$2HuLJJa1611votr5CHKX/OS3xK1fAREGgbmLKSWo4Ewf5bzDGvCwi'),
 ('VOhU5DZvyh', '$2a$04$LTNRQBExFHY8lQCn44lo/ehB/NRpBIpddcfgIVLGGUYtvJmzJ3/9O'),
 ('JKBL7JWn7M', '$2a$04$vJVvUq12hZbehFmPzP.c8OnRsuSErg/5.1dLO7hq3pPFlvIQgaobu'),
 ('8rKHqG6Lf1', '$2a$04$zoGEpfPL0Ypk8OengyTkvuOdM0QVy30SWsY1C7y1aBvWBICJYu51y'),
 ('R6EFdVhNjE', '$2a$04$NFEv49zphsEdwCNLi6QYz.FbSSEyNqGuIsJ9hNj.nbXE5Z624plBm'),
 ('CqV7e0hZ8g', '$2a$04$mAbZO/nJE.f8CjUSmm/IE.hqbx6XvnWlJmbNNi0ikU4p4mRcSDfuy'),
 ('2WZkQzPzMT', '$2a$04$aOtXqzAbj22O492SA6pp4euu5zHmRfxxNoG.Yp4qQf9O2NiKpmVJu'),
 ('pAC2LJQlm5', '$2a$04$/0I5Z3FbgSF8C7k8qSfp3.ytDI6zVKZQ0u1qb4OeREpHxQzl6CTFW'),
 ('v3TUBsnte9', '$2a$04$b9YqfFmp0r1WI.OGkygre./9HgRDPIFojEZZerozLRkDmQI1Fgz9a'),
 ('nlKNfYol8W', '$2a$04$kdb7klvhAf6bAf.prygvoungTlvAxg6KZkBOYyo4W6Zi3sU3I495m'),
 ('nIiWV3bH0B', '$2a$04$AXKaWtL0qEK/8xW.80xPMey9ariOJJgDcAxIsh9/aZ0fp7NSIC8cq'),
 ('ppfEtph41U', '$2a$04$rJ7mfRo9Aw9VW5mcgkH/se52xwvhBDUVDZEKjsrpxkUsp5ITv2zOS'),
 ('wmQSs3vcp8', '$2a$04$xPHuVYa.1gExp.Xh3kUESOhgaBBHfbTwutEYI8rrOzwOIWM.xLHg.'),
 ('J2rASrg4bx', '$2a$04$se4XG.MHrkFFuSXB/aSX1uNxqY0LdYnetRSqOvTrrgFtuB9UP84TK'),
 ('6fvAjnILCJ', '$2a$04$As9ZsAB/qUV3RHXTPW0LDOecwR.GqPIhbBHyXmgTHxVOE3xE/8G8i'),
 ('VWg4BzvoHH', '$2a$04$a2FyUeWJPpgE//4QaPNUSePy8ie9/KM46MavQwzvTx32r05ZE2vPC'),
 ('yQL5lVVa1V', '$2a$04$1eF6i15wMQvGByE78X53Su7GiDOGKGBGtLuESNq.QS5YwXsICaIx6'),
 ('85OXZPxcl3', '$2a$04$dAHk5cZpQrHAyj1mw5.HW.W15pHZm5nwY0uPmN/4hknber.FNS.m.'),
 ('r4MszAnqDZ', '$2a$04$nvAPnhFN3DbDGykNeY9u6e7uZx/RH8ID7JhwXIjNImjsPzZeTJQXu'),
 ('6EAjGLecX3', '$2a$04$te4Lr4UAR1nJRlW.n17STuSR3HDjx1wCvnEKyQso4x/tlpWtmOtaa'),
 ('9ayb3VwvQI', '$2a$04$6CW3TBbvY3ja/vT.gfdDS.m0M1BmgxjZRE7Xyunk6BFDEkbqlaqn6'),
 ('Ga1xJJzckM', '$2a$04$dSbZPImycO340SYKNxZ/Q.Gwbz9g7hIJ7Cnddjqptv6Zmf6HVgyzC'),
 ('5Txefv5avX', '$2a$04$aGeuD2L8RycBjr3DqWbZUOs2w90DCowMfOFLpkW8sabWl1s3FmRoK'),
 ('KG4OXlfe5f', '$2a$04$jlMvCu6723IQhhKTEgwZWeRA0MoVuDp5Jm1TXkr4VYaunzWewwI0G'),
 ('Ok2F4fE1g2', '$2a$04$P3AbZWSr0kefGE86Jzy.rOK3eTvN1XZA4R5igmI.qgmuj8gvV29J.'),
 ('tiUPXJmf61', '$2a$04$di0xhcmFLQYhttsk59.KIec.lYk2Pg2IzidWkst1sdYRjFPfIZE26'),
 ('lIlSRkBtG1', '$2a$04$Kyjwam4jnpx3kgsHO2LI1uNeXehTyAOKx6bLS/P2vfUimriVhpxoi'),
 ('OIoWvWVg5a', '$2a$04$cGmoHuamN4w.LvrIqU1mwO8BKvYlZbmxZQFGRmSDCMRAmRCXt1B7.'),
 ('9reFZM9nCr', '$2a$04$YT65ydKjNDzNXvxK2C3NCeR2MNydvke77NqauyRGtknDidvbDng2W'),
 ('20duB9rCH2', '$2a$04$VcfyEdo7EZPDyAXcbt/f9.kJH2of1qAn1wbfMsQZ/WBfgN8JH4hLC'),
 ('j336MTNwH1', '$2a$04$ycsZAgjpnNgA671FYbukYeT6LCor5Db/92e3gqRdJM6cYpaHam4K6'),
 ('1SGjqzuQ3s', '$2a$04$DJwI/7petpqGJCRmDSG5nuQnYchTl20az2rsgxu8Uc7wAImKveQCu'),
 ('RAMLlgPjO1', '$2a$04$UwZb6IFm4tgS1pLvujdaV.sK3Gq9zN2IJyIuGjqtMyuSL/dZXfqjK'),
 ('4NW9VChCK9', '$2a$04$p.G7Rfg50hqn6oFyNHEFCesVkulf1gIMikz8y8uqA1MGDoswG8R7y'),
 ('LTCAgv9M6R', '$2a$04$PUrpyrjNpP70jWQyFIS90OhOzVk3PRTdtDSn97v5GlHG5//.Pc8ma'),
 ('aO0WnMsRQm', '$2a$04$/xwDxJw0l7il8nerRB7Ey.zwvtCMT3HP8ha6Sil9l6ZNL8eYuTSd6'),
 ('pTENv0rN8F', '$2a$04$Bfs9ObF/qjudHK.wofdamOQpZWHt3yPiU2qsUEDwprjtPcPNYqGqu'),
 ('Sj1HFGZyyW', '$2a$04$GWjJPRhvDfqpEvp8bp7dW.ug.BHMx9em4WL8BSRPuMXj/ylhr8TmW'),
 ('7I5g9FXs6J', '$2a$04$Wyg2OV0Eg7eSCnoFqtJLgujNnkduv2.4hrKlyiAANYCN0uIq/Df0O'),
 ('12uDTPVna9', '$2a$04$t4O0rPbs5QVsxI9VFAXoyeWsJ/rHAdpFs4ChmoeOn4WStCBmHc9dC'),
 ('ntd2Ews4rm', '$2a$04$p555Qy6kgTbVo67k29l9VuWeqo1vPANms6He5BGOgn2ypqolyR3Re'),
 ('6dYPghdRm3', '$2a$04$0puas75g4hIjKHehCQqrMe1GCX8szA3.I6oCpQ3.tK3kudFHCLcRO'),
 ('KD1unsVh2s', '$2a$04$Mfrba/tPNPbWeMHJ67dVG.mbMrbkI/J3QQrZpKr7H4.oYt.cWZ1Ui'),
 ('u2RXexyc6G', '$2a$04$7ZWkDLISNMr//oWec9KFceiQBYZ6czBLle3kvgjL9ek5B/FA82wD6'),
 ('2ibZjBYIOJ', '$2a$04$TMA.ea7HW0wmcLE0vd6hQ.PsJ1mZ4niDzhZN1gi7yiWPoPnawgsrm'),
 ('2JgOu8llWF', '$2a$04$807xAItPSriQOKMQ9iBW5OBfqwuWei.fJIhMJ3nlP9B9MX5yC/yo6'),
 ('dSrzfRfU4D', '$2a$04$0PccloP90oBYVZSg1GLWFO0FJ7hRVaK6/NL2dFfEtFjoe8gUYiuN6'),
 ('e7yRhQlegy', '$2a$04$q3ocvbPjvd5Sp1MYZwHciOn3yzDEQWVkpxkeoDzywyHlJMQX8CmXq'),
 ('RbCl8zlg3d', '$2a$04$C5kAIMdMDC9Fcap9LTzBK.4NEn8gIGZgPWzelYFLddPbIni7MZETG'),
 ('87T7OU0q87', '$2a$04$RbZ/5L2foJkNG2EgXMUblOgZ6xmEdxGTxQjZqyu4lCaqGR9yuCr3y'),
 ('snk1A3No8R', '$2a$04$DpI07.7ql.45WjIfl28LP.8nI005.krxyHLudkIdX0CKmcW4OCh6m'),
 ('SgN9ebRbzS', '$2a$04$.GkE/OUCwvZ2gheBHs7q4uOf/YMfAj9PNpYn6jCwrOEx16dspw2qa'),
 ('PRx7yzcs97', '$2a$04$3EUNF8B0yBBbGUWMH0jo1.2p.tlXVGi3Q3cdvU8rJuHgmJo4QwOQG'),
 ('9FEArrqlzw', '$2a$04$s.JWzrfBPfrBq5gyPL7sv.ldvADiNe2US9iKWtL.qY5qm1zZeAFge'),
 ('Rnnlm6Sy4d', '$2a$04$WE0FHGKNQ0HAEBZ8Xi9NDe32IIIyjTv9X3yuBQhvbS5U9BrylEMEW'),
 ('26Ec9UBup9', '$2a$04$NTzkO97hOWpI1U3mx.odt.LreHylHcR9RjZkBPhSqbnJ.U6jOhI2a'),
 ('E5CU7wRzrx', '$2a$04$Z27Z/RwqmDnOXnXFiqbNBOZN.D.GqX0VV/AZQUgUQ7kqfA5MMBn7m'),
 ('Fi1DkMf48w', '$2a$04$KMktrTFWS6MIiy9InssireRwhpPrhWbgjy0eQqyoNIKaboUvqJ6.K'),
 ('Khk8eJrk0C', '$2a$04$xpe.xozmfHv7Z/2jNRuaee1bD8IXeEi4WDXM4zF76kcJ0UCZK0aLW'),
 ('IPT2oeh313', '$2a$04$5CnWWFdt4IauwyC4BjqTtuDrLiQ7Jkvi1MkaOiojVvQ1vMQER.p.W'),
 ('ADEzNpgA4B', '$2a$04$GO61jM8wsg0iiiPC6KWmZe05mDmajliBl0GtxRLCnK2eq5Mo/.C/O'),
 ('4tbLp3EuIQ', '$2a$04$nBWPPUXpkUpHWWZ7kJYR2OLxhYsiLuAVeDsDCO1WcQZEQ3Q8O9jFC'),
 ('5f1bBrJ6vg', '$2a$04$T5oIWCDo9kr3TdZy8Ek1TegBsy3rzF0u/9HrJex1uRdED8xwE8Hhy'),
 ('8BwzS8xwpF', '$2a$04$uyoIa6HiiwtLEadlWJBcLeWrQ02LtC9hBtKy99UWDMr1fekNacxBS'),
 ('tIPoy1bb6H', '$2a$04$wNDY/crOYymoNs8G1gyRNO36/YSWS/aL8Mcv.h2nDm9..TFoJ7hCa'),
 ('dhzErI8w2Y', '$2a$04$kajn7WCr5lgnPy73ZudHxOy.XPkhEMTCPwIW.ANUhrqTViKKrWwbm'),
 ('l3RC8E3iot', '$2a$04$EhXqaOrq7tGcR61FKb1vxOVNSeLubLpFoRJDXxERMvUQZH5a.tT/q'),
 ('5Muu7xKeV8', '$2a$04$5Lh81d1JW1AeHNueEENUVumXWpa6qhU.QRn0YJ1XFWRWg69Tc4PkG'),
 ('LyiVSLc819', '$2a$04$uj2kC5wJNTrXKzFGTBfdhOKC.DYwJ6tejgVWdkZbzVXp/wY7AHW6m'),
 ('3awXYfvTIe', '$2a$04$mOWpE7fPGGnKrMDLz0o71u6FPNfIu/60SvKe9CNTKAAQ4cGd48a1m'),
 ('222PoKxk2V', '$2a$04$evS0vE3fQ4gtple5.lY6BOatdNX6jKG6eQlpbRtYQDNBQxbVRoiFW'),
 ('0EEnjTtY0S', '$2a$04$4g7GzmasnxaXMkyaniG3AuuyP/TS0sq1x5oTi9KphY.59juebN51W'),
 ('T7Rgd7kxO7', '$2a$04$q20IW8IrwmTn9BWDq349ROYis2Tr3iQ3dvagCKd21VSKSlB1bLNla'),
 ('H6FesFRrn6', '$2a$04$W1M2RwbskgN4Xcj3SfoNreWw4p6TbM7cbhQHE5vxbDxQDEefwUxny'),
 ('5a3Pbz8r8Z', '$2a$04$15IR25zhI4AmeMXNpzRC9eXFostWbJzQ0AkT7qD4.E5VvgwVIeKJq'),
 ('0sYz8ToWRZ', '$2a$04$34kAlIoUPVRvakcqe/xy8.YXw0R3Xq6c91qsch0rfR/MfiBTNkaTi'),
 ('JeyH5yHXz5', '$2a$04$lAp9YLJjs/1DljHgTubJBuL5r/.2NZDMnTylFpYBsdYeaznUcDkJ.'),
 ('AMyEyaNz7L', '$2a$04$ywZmBv6LQYbWGpA3yxaMcOIEQn1G.czkAd5uEK7fTeq3lfoaMoVLW'),
 ('UCFpheZn5N', '$2a$04$w5.MtTa6RDKEbD9j7XnUnOAYBcpug2XAVrjLgFg3wrlRHppbO52Gy'),
 ('4M4D5JngOt', '$2a$04$0z1ni/zkE6a9Sfsaq3vb6O.Xma1va8H3h5/IfFVBH0MxrUI.OZRNm'),
 ('6p4iVPAopm', '$2a$04$f1BJVM5MCc6MpOOEbCtJGOI8xY3Alas0Ov8fVNuYej8b1idv59cCW'),
 ('Ov4EugVZ0G', '$2a$04$JNkWqunO87FabYimlZkUqOyC02iDIbfQB35m.PKW7N/x2U8uqCEEW'),
 ('4hZ2ziNnjs', '$2a$04$cyjbSBfNcFM.VG2ovtnqg.xlmhsOaEy.yygmJ7hlNIM2STJ4Wgh5.'),
 ('4H8stzAJog', '$2a$04$qziCjx2Kh1JpW/bYzjKFc.v7jrvki3LJjpAxZqnlRJN.qn.NBT492'),
 ('8UWpBpmGjm', '$2a$04$6/rZQ./QnEYIa/XCChrIlOtQ.62Z0ZSRmWR27AmuY43XKPXcn8UmW'),
 ('54mxZwBnA8', '$2a$04$4R3oaiwgL3jFDx4XkSPgWeK.uk35cdQm.wKhYVFSFMREEtEUtuPZu'),
 ('9Ywtlayp0Q', '$2a$04$/v16fIbxYBiHvtEmtjgydeJ/fUI2H0OhCgNdTReh5WZUtHYvubDDi')]

# SQL
SQL_FORMAT="INSERT INTO users (id, name, display_name, description, password) VALUES ({user_id}, '{name}', '{display_name}', '{description}', '{password}');"
INSERT_THEME_FORMAT="INSERT INTO themes (user_id, dark_mode) VALUES ({user_id}, {dark_mode});"

# Go
GO_CONSTRUCTOR_FORMAT="&User{{ Name: \"{name}\", DisplayName: \"{display_name}\", Description: \"{description}\", RawPassword: \"{raw_password}\", HashedPassword: \"{hashed_password}\", DarkMode: {dark_mode} }},"

#
DESCRIPTION_FORMAT="普段{job}をしています。\\nよろしくおねがいします！\\n\\n連絡は以下からお願いします。\\n\\nウェブサイト: {website}\\nメールアドレス: {mail}\\n"

fake = Faker('ja-JP')


with open("./initial-data/display_names.txt", "r") as f:
    display_names = f.readlines()
    display_names = list(display_name.rstrip() for display_name in display_names)


def get_display_name():
    return random.choice(display_names)


# NOTE: bcryptの最小コストは4。ログイン負荷は趣旨ではないので最も低いコストを採用
def hash_password(raw_password):
    result = subprocess.run([
        'bcrypt-tool', 'hash', raw_password, '4'
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

        n = random.randint(1, 10)
        dark_mode = ['true', 'false'][n % 2]
        theme_sql = INSERT_THEME_FORMAT.format(user_id=user['user_id'], dark_mode=dark_mode)
        output += f'{theme_sql}\n'

    return output


def format_go(users, initial_users):
    mid = len(users)//2

    output = "package scheduler\n"

    # 初期ユーザ
    output += "var initialUserPool = []*User{\n"
    for user in initial_users:
        ctor = GO_CONSTRUCTOR_FORMAT.format(
            name = user['name'],
            display_name = user['display_name'],
            description = user['description'],
            raw_password = user['raw_password'],
            hashed_password = user['hashed_password'],
            dark_mode = user['dark_mode'],
        )
        output += f'\t{ctor}\n'
    output += "}\n\n"

    # 配信者
    output += "var streamerPool = []*User{\n"
    for user in users[:mid]:
        ctor = GO_CONSTRUCTOR_FORMAT.format(
            name = user['name'],
            display_name = user['display_name'],
            description = user['description'],
            raw_password = user['raw_password'],
            hashed_password = user['hashed_password'],
            dark_mode = user['dark_mode'],
        )
        output += f'\t{ctor}\n'
    output += "}\n\n"

    # 視聴者
    output += "var viewerPool = []*User{\n"
    for user in users[mid:]:
        ctor = GO_CONSTRUCTOR_FORMAT.format(
            name = user['name'],
            display_name = user['display_name'],
            description = user['description'],
            raw_password = user['raw_password'],
            hashed_password = user['hashed_password'],
            dark_mode = user['dark_mode'],
        )
        output += f'\t{ctor}\n'
    output += "}\n"

    return output


def gen_users(n: int):
    # NOTE: test001は、フロントエンド検証用ユーザ。マニュアルに記載
    users = [
        dict(
            user_id=1,
            name = 'test001',
            display_name = '検証用ユーザ',
            description='社内検証用',
            raw_password = 'test',
            hashed_password = '$2a$04$LBt4Dc0Uu3HE0c.8KVMtbOnXwd4PHCboGxa2I57RmJFQVba/B0U8a',
            dark_mode="true"
        )
    ]
    username_indexes = defaultdict(int)
    for user_id in range(2, n+1):
        profile = fake.profile()
        name = profile['username']
        name_idx = username_indexes[name]
        username = f"{name}{name_idx}"

        website = f'http://{name}.example.com/'
        mail = f'{name}@example.com'
        raw_password, hashed_password = passwords[(user_id-1)%len(passwords)]
        dark_mode = str(fake.boolean()).lower()

        users.append(dict(
            user_id = user_id,
            name = username,
            display_name = get_display_name(),
            description = DESCRIPTION_FORMAT.format(
                job=profile['job'],
                website=website,
                mail=mail,
            ),
            raw_password = raw_password,
            hashed_password = hashed_password,
            dark_mode = dark_mode,
        ))

        username_indexes[name] += 1
    return users


def generate(users) -> str:
    initial_users = users[:1000]
    workload_users = users[1000:]

    # 負荷データ
    with open('/tmp/user.go', 'w') as f:
        f.write(format_go(workload_users, initial_users) + '\n')

    # 初期データ
    with open('/tmp/user.sql', 'w') as f:
        f.write('-- NOTE: パスワードは `test`\n')
        f.write(format_sql(initial_users) + '\n')


def dump_passwords():
    passwords = []
    for _ in range(100):
        raw_password = fake.password(special_chars=False)
        hashed_password = hash_password(raw_password)
        passwords.append((raw_password, hashed_password, ))
    pprint.pprint(passwords)


def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('-n', type=int, default=5000, help='生成数')
    return parser.parse_args()


def main():
    args = get_args()
    users = gen_users(args.n)
    generate(users)


if __name__ == '__main__':
    main()
    # dump_passwords()


