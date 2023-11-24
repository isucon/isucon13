#!/usr/bin/env python

import hashlib
import io
import json
import os
import subprocess
import sys
import uuid
from base64 import b64decode
from dataclasses import asdict
from datetime import datetime, timedelta, timezone
from http.client import (
    BAD_REQUEST,
    CREATED,
    FORBIDDEN,
    INTERNAL_SERVER_ERROR,
    NOT_FOUND,
    OK,
    UNAUTHORIZED,
)
from typing import Any

import bcrypt
import models
import mysql.connector
from flask import Flask, Response, request, send_file, session
from mysql.connector.errors import DatabaseError
from sqlalchemy import create_engine


class Settings(object):
    LISTEN_PORT = 8080

    DB_HOST = os.getenv("ISUCON13_MYSQL_DIALCONFIG_ADDRESS", "127.0.0.1")
    DB_PORT = int(os.getenv("ISUCON13_MYSQL_DIALCONFIG_PORT", 3306))
    DB_USER = os.getenv("ISUCON13_MYSQL_DIALCONFIG_USER", "isucon")
    DB_PASSWORD = os.getenv("ISUCON13_MYSQL_DIALCONFIG_PASSWORD", "isucon")
    DB_NAME = os.getenv("ISUCON13_MYSQL_DIALCONFIG_DATABASE", "isupipe")

    POWERDNS_ENABLED = os.getenv("ISUCON13_POWERDNS_DISABLED") != "true"
    POWERDNS_SUBDOMAIN_ADDRESS = os.getenv("ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS")

    SESSION_COOKIE_DOMAIN = "u.isucon.dev"
    SESSION_COOKIE_PATH = "/"
    SESSION_SECRET_KEY = os.getenv(
        "ISUCON13_SESSION_SECRETKEY", "isucon13_session_cookiestore_defaultsecret"
    )
    PERMANENT_SESSION_LIFETIME = 600  # 10min

    DEFAULT_SESSION_ID_KEY = "SESSIONID"
    DEFAULT_SESSION_EXPIRES_KEY = "EXPIRES"
    DEFAULT_USER_ID_KEY = "USERID"
    DEFAULT_USER_NAME_KEY = "USERNAME"

    FALLBACK_IMAGE = "../img/NoImage.jpg"
    BCRYPT_DEFAULT_COST = 4


app = Flask(__name__)


# 初期化
@app.route("/api/initialize", methods=["POST"])
def initialize_handler() -> tuple[dict[str, Any], int]:
    # app.logger.info("start initialize")
    result = subprocess.run(["../sql/init.sh"], capture_output=True, text=True)
    if result.returncode != 0:
        app.logger.error(
            'init.sh failed with status=%d err="%s"', (result.returncode, result.stdout)
        )
        raise HttpException(result.stderr, INTERNAL_SERVER_ERROR)

    return {"language": "python"}, OK


# top
@app.route("/api/tag", methods=["GET"])
def get_tag_handler() -> tuple[dict[str, Any], int]:
    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM tags"
        c.execute(sql)
        rows = c.fetchall()
        if rows is None:
            raise HttpException("failed to get tags", INTERNAL_SERVER_ERROR)

        tags = models.Tags(tags=[models.Tag(**row) for row in rows])

        conn.commit()
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.close()
    return asdict(tags), OK


# 配信者のテーマ取得API
@app.route("/api/user/<string:username>/theme", methods=["GET"])
def get_streamer_theme_handler(username: str) -> tuple[dict[str, Any], int]:
    verify_user_session()

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT id FROM users WHERE name = %s"
        c.execute(sql, [username])
        row = c.fetchone()
        if row is None:
            raise HttpException("not found", NOT_FOUND)

        sql = "SELECT * FROM themes WHERE user_id = %s"
        c.execute(sql, [row["id"]])
        row = c.fetchone()
        if row is None:
            raise HttpException("not found", NOT_FOUND)
        theme_model = models.ThemeModel(**row)
        theme = models.Theme(id=theme_model.id, dark_mode=theme_model.dark_mode)
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()
    return asdict(theme), OK


# livestream
# reserve livestream
@app.route("/api/livestream/reservation", methods=["POST"])
def reserve_livestream_handler() -> tuple[dict[str, Any], int]:
    verify_user_session()

    user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
    if not user_id:
        raise HttpException("unauthorized", UNAUTHORIZED)

    req = get_request_json()

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        # 2023/11/25 10:00からの１年間の期間内であるかチェック
        term_start_at = datetime(2023, 11, 25, 1, 0, 0, tzinfo=timezone.utc)
        term_end_at = datetime(2024, 11, 25, 1, 0, 0, tzinfo=timezone.utc)
        reserve_start_at = datetime.fromtimestamp(
            float(req["start_at"]), tz=timezone.utc
        )
        reserve_end_at = datetime.fromtimestamp(float(req["end_at"]), tz=timezone.utc)

        if reserve_start_at >= term_end_at or reserve_end_at <= term_start_at:
            raise HttpException("bad reservation time range", BAD_REQUEST)

        # 予約枠をみて、予約が可能か調べる
        # NOTE: 並列な予約のoverbooking防止にFOR UPDATEが必要
        sql = "SELECT * FROM reservation_slots WHERE start_at >= %s AND end_at <= %s FOR UPDATE"
        c.execute(sql, [int(req["start_at"]), int(req["end_at"])])
        rows = c.fetchall()
        if rows is None:
            app.logger.error("予約枠一覧取得でエラー発生")
            raise HttpException(
                "failed to get reservation_slots",
                INTERNAL_SERVER_ERROR,
            )
        slots = [models.ReservationSlotModel(**row) for row in rows]

        for slot in slots:
            sql = (
                "SELECT slot FROM reservation_slots WHERE start_at = %s AND end_at = %s"
            )
            c.execute(sql, [slot.start_at, slot.end_at])
            row = c.fetchone()
            if row is None:
                raise HttpException(
                    "failed to get reservation_slots",
                    INTERNAL_SERVER_ERROR,
                )
            count = int(row["slot"])

            # app.logger.info(f"{slot.start_at}~{slot.end_at}予約枠の残数 = {count}")
            if count < 1:
                raise HttpException(
                    f"予約期間 {term_start_at.timestamp()} ~ {term_end_at.timestamp()}に対して、予約区間 {req['start_at']} ~ {req['end_at']}が予約できません",
                    BAD_REQUEST,
                )

        livestream_model = models.LiveStreamModel(
            id=0,  # 未設定
            user_id=user_id,
            title=req["title"],
            description=req["description"],
            playlist_url=req["playlist_url"],
            thumbnail_url=req["thumbnail_url"],
            start_at=int(req["start_at"]),
            end_at=int(req["end_at"]),
        )

        sql = "UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= %s AND end_at <= %s"
        c.execute(sql, [int(req["start_at"]), int(req["end_at"])])

        sql = "INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(%s, %s, %s, %s, %s, %s, %s)"
        c.execute(
            sql,
            [
                livestream_model.user_id,
                livestream_model.title,
                livestream_model.description,
                livestream_model.playlist_url,
                livestream_model.thumbnail_url,
                livestream_model.start_at,
                livestream_model.end_at,
            ],
        )

        livestream_model.id = c.lastrowid

        # タグ追加
        if "tags" in req and req["tags"]:
            for tag_id in req["tags"]:
                sql = "INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (%s, %s)"
                c.execute(sql, [livestream_model.id, tag_id])

        livestream = fill_livestream_response(c, livestream_model)
        if not livestream:
            raise HttpException("failed to fill livestream", INTERNAL_SERVER_ERROR)

        return asdict(livestream), CREATED
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# list livestream
@app.route("/api/livestream/search", methods=["GET"])
def search_livestreams_handler() -> tuple[list[dict[str, Any]], int]:
    key_tag_name = request.args.get("tag")
    limit_str = request.args.get("limit")

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        livestream_models = []
        if key_tag_name:
            # タグによる取得
            sql = "SELECT id FROM tags WHERE name = %s"
            c.execute(sql, [key_tag_name])
            rows = c.fetchall()
            if not rows:
                raise HttpException("failed to get tags", INTERNAL_SERVER_ERROR)
            tag_ids = [row["id"] for row in rows]

            sql = "SELECT * FROM livestream_tags WHERE tag_id IN (%s) ORDER BY livestream_id DESC"  # idかtag_idか要確認
            in_formats = ",".join(["%s"] * len(tag_ids))
            sql = sql % in_formats
            c.execute(sql, tag_ids)
            rows = c.fetchall()
            if rows is None:
                raise HttpException(
                    "failed to get keyTaggedLivestream", INTERNAL_SERVER_ERROR
                )
            key_tagged_livestreams = [models.LiveStreamTagModel(**row) for row in rows]

            for key_tagged_livestream in key_tagged_livestreams:
                sql = "SELECT * FROM livestreams WHERE id = %s"
                c.execute(sql, [key_tagged_livestream.livestream_id])
                row = c.fetchone()
                if row is None:
                    raise HttpException(
                        "failed to get livestream", INTERNAL_SERVER_ERROR
                    )

                livestream_model = models.LiveStreamModel(**row)

                livestream_models.append(livestream_model)

        else:
            # 検索条件なし
            sql = "SELECT * FROM livestreams ORDER BY id DESC"
            args = []
            if limit_str:
                sql += " LIMIT %s"
                args.append(int(limit_str))

            c.execute(sql, args)
            rows = c.fetchall()
            livestream_models = [models.LiveStreamModel(**row) for row in rows]

        livestreams = []
        for livestream_model in livestream_models:
            livestream = fill_livestream_response(c, livestream_model)
            if not livestream:
                raise HttpException("error", INTERNAL_SERVER_ERROR)

            # HTTPレスポンスに使うのでasdictしてからリストに突っ込む
            livestreams.append(asdict(livestream))

        return livestreams, OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


@app.route("/api/livestream", methods=["GET"])
def get_my_livestreams_handler() -> tuple[list[dict[str, Any]], int]:
    verify_user_session()

    user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
    if not user_id:
        raise HttpException("unauthorized", UNAUTHORIZED)

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM livestreams WHERE user_id = %s"
        c.execute(sql, [user_id])
        rows = c.fetchall()
        if rows is None:
            raise HttpException(
                "failed to get livestreams",
                INTERNAL_SERVER_ERROR,
            )
        if len(rows) == 0:
            rows = []
        livestream_models = [models.LiveStreamModel(**row) for row in rows]

        livestreams = []
        for livestream_model in livestream_models:
            livestream = fill_livestream_response(c, livestream_model)
            if not livestream:
                raise HttpException(
                    "failed to fill livestream",
                    INTERNAL_SERVER_ERROR,
                )
            livestreams.append(asdict(livestream))
        return livestreams, OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


@app.route("/api/user/<string:username>/livestream", methods=["GET"])
def get_user_livestreams_handler(username: str) -> tuple[list[dict[str, Any]], int]:
    verify_user_session()

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM users WHERE name = %s"
        c.execute(sql, [username])
        row = c.fetchone()
        if row is None:
            raise HttpException("user not found", NOT_FOUND)
        user = models.UserModel(**row)

        sql = "SELECT * FROM livestreams WHERE user_id = %s"
        c.execute(sql, [user.id])
        rows = c.fetchall()
        if rows is None:
            raise HttpException(
                "failed to get livestreams",
                INTERNAL_SERVER_ERROR,
            )

        livestream_models = [models.LiveStreamModel(**row) for row in rows]

        livestreams = []
        for livestream_model in livestream_models:
            livestream = fill_livestream_response(c, livestream_model)
            if not livestream:
                raise HttpException(
                    "failed to fill livestream",
                    INTERNAL_SERVER_ERROR,
                )
            livestreams.append(asdict(livestream))
        return livestreams, OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# get livestream
@app.route("/api/livestream/<int:livestream_id>", methods=["GET"])
def get_livestream_handler(livestream_id: int) -> tuple[dict[str, Any], int]:
    verify_user_session()

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM livestreams WHERE id = %s"
        c.execute(sql, [livestream_id])
        row = c.fetchone()
        if row is None:
            raise HttpException("not found", NOT_FOUND)
        livestream_model = models.LiveStreamModel(**row)

        livestream = fill_livestream_response(c, livestream_model)
        return asdict(livestream), OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# get polling livecomment timeline
@app.route("/api/livestream/<int:livestream_id>/livecomment", methods=["GET"])
def get_livecomments_handler(livestream_id: int) -> tuple[list[dict[str, Any]], int]:
    verify_user_session()

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM livecomments WHERE livestream_id = %s ORDER BY created_at DESC"
        args = [livestream_id]
        limit_str = request.args.get("limit")
        if limit_str:
            sql += " LIMIT %s"
            args.append(int(limit_str))
        c.execute(sql, args)

        rows = c.fetchall()
        if rows is None:
            raise HttpException("not found", NOT_FOUND)
        if len(rows) == 0:
            rows = []

        livecomment_models = [models.LiveCommentModel(**row) for row in rows]

        livecomments: list[dict[str, Any]] = []
        for livecomment_model in livecomment_models:
            livecomment = fill_livecomment_response(c, livecomment_model)
            livecomments.append(asdict(livecomment))

        return livecomments, OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# ライブコメント投稿
@app.route("/api/livestream/<int:livestream_id>/livecomment", methods=["POST"])
def post_livecomment_handler(livestream_id: int) -> tuple[dict[str, Any], int]:
    verify_user_session()

    user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
    if not user_id:
        raise HttpException("unauthorized", UNAUTHORIZED)

    req = get_request_json()

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM livestreams WHERE id = %s"
        c.execute(sql, [livestream_id])
        row = c.fetchone()
        if row is None:
            raise HttpException("livestream not found", NOT_FOUND)
        livestream_model = models.LiveStreamModel(**row)

        # スパム判定
        sql = "SELECT id, user_id, livestream_id, word FROM ng_words WHERE user_id = %s AND livestream_id = %s"
        c.execute(sql, [livestream_model.user_id, livestream_model.id])
        ng_words = c.fetchall()
        if ng_words is None:
            raise HttpException("failed to get NG words", INTERNAL_SERVER_ERROR)

        for ng_word in ng_words:
            sql = """
                    SELECT COUNT(*)
                    FROM
                    (SELECT %s AS text) AS texts
                    INNER JOIN
                    (SELECT CONCAT('%%', %s, '%%')	AS pattern) AS patterns
                    ON texts.text LIKE patterns.pattern;
                """
            c.execute(sql, [req["comment"], ng_word["word"]])
            hit_spam = c.fetchone()
            if not hit_spam:
                raise HttpException("failed to get hitspam", INTERNAL_SERVER_ERROR)
            hit_spam = hit_spam["COUNT(*)"]
            # app.logger.info(f"[hitSpam={hit_spam}] comment = {req['comment']}")
            if hit_spam >= 1:
                raise HttpException("このコメントがスパム判定されました", BAD_REQUEST)

        now = int(datetime.now().timestamp())

        sql = "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (%s, %s, %s, %s, %s)"
        c.execute(sql, [user_id, livestream_id, req["comment"], req["tip"], now])

        livecomment_id = c.lastrowid
        livecomment_model = models.LiveCommentModel(
            id=livecomment_id,
            user_id=user_id,
            livestream_id=livestream_id,
            comment=req["comment"],
            tip=req["tip"],
            created_at=now,
        )
        app.logger.info(livecomment_model)
        livecomment = fill_livecomment_response(c, livecomment_model)
        return asdict(livecomment), CREATED
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


@app.route("/api/livestream/<int:livestream_id>/reaction", methods=["POST"])
def post_reaction_handler(livestream_id: int) -> tuple[dict[str, Any], int]:
    verify_user_session()

    user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
    if not user_id:
        raise HttpException("unauthorized", UNAUTHORIZED)

    req = get_request_json()

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        now = int(datetime.now().timestamp())
        sql = "INSERT INTO reactions (user_id, livestream_id, emoji_name, created_at) VALUES (%s, %s, %s, %s)"
        c.execute(sql, [user_id, livestream_id, req["emoji_name"], now])

        reaction_id = c.lastrowid
        reaction_model = models.ReactionModel(
            id=reaction_id,
            user_id=user_id,
            livestream_id=livestream_id,
            emoji_name=req["emoji_name"],
            created_at=now,
        )

        reaction = fill_reaction_response(c, reaction_model)
        return asdict(reaction), CREATED
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


@app.route("/api/livestream/<int:livestream_id>/reaction", methods=["GET"])
def get_reactions_handler(
    livestream_id: int,
) -> tuple[list[dict[str, Any]] | dict[str, Any], int]:
    verify_user_session()

    limit_str = request.args.get("limit")

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = (
            "SELECT * FROM reactions WHERE livestream_id = %s ORDER BY created_at DESC"
        )
        args = [livestream_id]
        if limit_str:
            sql += " LIMIT %s"
            args.append(int(limit_str))

        c.execute(sql, args)
        rows = c.fetchall()
        if rows is None:
            # app.logger.info("reaction_models")
            raise HttpException("failed to get reactions", INTERNAL_SERVER_ERROR)
        reaction_models = [models.ReactionModel(**row) for row in rows]

        reactions = []
        for reaction_model in reaction_models:
            reaction = fill_reaction_response(c, reaction_model)
            reactions.append(asdict(reaction))

        return reactions, OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# (配信者向け)ライブコメントの報告一覧取得API
@app.route("/api/livestream/<int:livestream_id>/report", methods=["GET"])
def get_livecomment_reports_handler(
    livestream_id: int,
) -> tuple[list[dict[str, Any]], int]:
    verify_user_session()

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM livestreams WHERE id = %s"
        c.execute(sql, [livestream_id])
        row = c.fetchone()
        if not row:
            raise HttpException("failed to get livestream", INTERNAL_SERVER_ERROR)
        livestream_model = models.LiveStreamModel(**row)

        user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
        if not user_id:
            raise HttpException("unauthorized", UNAUTHORIZED)

        if livestream_model.user_id != user_id:
            raise HttpException(
                "can't get other streamer's livecomment reports",
                FORBIDDEN,
            )

        sql = "SELECT * FROM livecomment_reports WHERE livestream_id = %s"
        c.execute(sql, [livestream_id])
        rows = c.fetchall()
        if rows is None:
            # app.logger.info("report_model")
            raise HttpException(
                "failed to get livecomment reports",
                INTERNAL_SERVER_ERROR,
            )
        report_models = [models.LiveCommentReportModel(**row) for row in rows]

        reports = []
        for report_model in report_models:
            report = fill_livecomment_report_response(c, report_model)
            if not report:
                # app.logger.info("failed to fill livecomment report")
                raise HttpException(
                    "failed to fill livecomment report",
                    INTERNAL_SERVER_ERROR,
                )
            reports.append(asdict(report))

        return reports, OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


@app.route("/api/livestream/<int:livestream_id>/ngwords", methods=["GET"])
def get_ngwords(livestream_id: int) -> tuple[list[dict[str, Any]], int]:
    verify_user_session()

    user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
    if not user_id:
        raise HttpException("unauthorized", UNAUTHORIZED)

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM ng_words WHERE user_id = %s AND livestream_id = %s ORDER BY created_at DESC"
        c.execute(sql, [user_id, livestream_id])
        rows = c.fetchall()
        if rows is None:
            app.logger.error("failed to get ngwords")
            raise HttpException("error", INTERNAL_SERVER_ERROR)
        if len(rows) == 0:
            return [], OK
        ngwords = [asdict(models.NGWord(**row)) for row in rows]
        return ngwords, OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# ライブコメント報告
@app.route(
    "/api/livestream/<int:livestream_id>/livecomment/<int:livecomment_id>/report",
    methods=["POST"],
)
def report_livecomment_handler(
    livestream_id: int, livecomment_id: int
) -> tuple[dict[str, Any], int]:
    verify_user_session()

    user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
    if not user_id:
        raise HttpException("failed to find user-id from session", UNAUTHORIZED)

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM livestreams WHERE id = %s"
        c.execute(sql, [livestream_id])
        row = c.fetchone()
        if not row:
            raise HttpException("livestream not found", NOT_FOUND)

        now = int(datetime.now().timestamp())
        sql = "INSERT INTO livecomment_reports (user_id, livestream_id, livecomment_id, created_at) VALUES (%s, %s, %s, %s)"
        c.execute(sql, [user_id, livestream_id, livecomment_id, now])

        report_id = c.lastrowid

        report_model = models.LiveCommentReportModel(
            id=report_id,
            user_id=user_id,
            livestream_id=livestream_id,
            livecomment_id=livecomment_id,
            created_at=now,
        )

        report = fill_livecomment_report_response(c, report_model)

        return asdict(report), CREATED
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# 配信者によるモデレーション (NGワード登録)
@app.route("/api/livestream/<int:livestream_id>/moderate", methods=["POST"])
def moderate_handler(livestream_id: int) -> tuple[dict[str, Any], int]:
    verify_user_session()

    user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
    if not user_id:
        raise HttpException("failed to find user-id from session", UNAUTHORIZED)

    req = get_request_json()
    if not req or "ng_word" not in req:
        raise HttpException(
            "failed to decode the request body as json",
            BAD_REQUEST,
        )

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        # 配信者自身の配信に対するmoderateなのかを検証
        sql = "SELECT * FROM livestreams WHERE id = %s AND user_id = %s"
        c.execute(sql, [livestream_id, user_id])
        owned_livestreams = c.fetchall()
        if owned_livestreams is None or len(owned_livestreams) == 0:
            raise HttpException(
                "A streamer can't moderate livestreams that other streamers own",
                BAD_REQUEST,
            )

        sql = "INSERT INTO ng_words(user_id, livestream_id, word, created_at) VALUES (%s, %s, %s, %s)"
        c.execute(
            sql,
            [
                user_id,
                livestream_id,
                req["ng_word"],
                datetime.now().timestamp(),
            ],
        )

        word_id = c.lastrowid
        # app.logger.info(f"word_id: {word_id}, word: {req['ng_word']}")

        sql = "SELECT * FROM ng_words WHERE livestream_id = %s"
        c.execute(sql, [livestream_id])
        rows = c.fetchall()
        if rows is None:
            raise HttpException("failed to get NG words", INTERNAL_SERVER_ERROR)
        ngwords = [models.NGWord(**row) for row in rows]

        # NGワードにヒットする過去の投稿も全削除する
        for ngword in ngwords:
            sql = "SELECT * FROM livecomments"
            c.execute(sql)
            rows = c.fetchall()
            if rows is None:
                app.logger.warn("failed to get livecomments")
                raise HttpException(
                    "failed to get livecomments",
                    INTERNAL_SERVER_ERROR,
                )
            livecomments = [models.LiveCommentModel(**row) for row in rows]

            for livecomment in livecomments:
                # app.logger.info(f"delete: {livecomment}")
                sql = """
                    DELETE FROM livecomments
                    WHERE
                    id = %s AND
                    (SELECT COUNT(*)
                    FROM
                    (SELECT %s AS text) AS texts
                    INNER JOIN
                    (SELECT CONCAT('%%', %s, '%%')	AS pattern) AS patterns
                    ON texts.text LIKE patterns.pattern) >= 1;
                """
                c.execute(sql, [livecomment.id, livecomment.comment, ngword.word])
        return asdict(models.ModerateResponse(word_id=word_id)), CREATED
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# livestream_viewersにINSERTするため必要
# ユーザ視聴開始 (viewer)
@app.route("/api/livestream/<int:livestream_id>/enter", methods=["POST"])
def enter_livestream_handler(livestream_id: int) -> tuple[str, int]:
    verify_user_session()

    user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
    if not user_id:
        raise HttpException("failed to find user-id from session", UNAUTHORIZED)

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "INSERT INTO livestream_viewers_history (user_id, livestream_id, created_at) VALUES(%s, %s, %s)"
        c.execute(sql, [user_id, livestream_id, int(datetime.now().timestamp())])
        return "", OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# ユーザ視聴終了 (viewer)
@app.route("/api/livestream/<int:livestream_id>/exit", methods=["DELETE"])
def exit_livestream_handler(livestream_id: int) -> tuple[str, int]:
    verify_user_session()

    user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
    if not user_id:
        raise HttpException("failed to find user-id from session", UNAUTHORIZED)

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "DELETE FROM livestream_viewers_history WHERE user_id = %s AND livestream_id = %s"
        c.execute(sql, [user_id, livestream_id])
        return "", OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# user
@app.route("/api/register", methods=["POST"])
def register_handler() -> tuple[dict[str, Any], int]:
    req = get_request_json()
    if not req:
        raise HttpException(
            "failed to decode the request body as json",
            BAD_REQUEST,
        )

    if not all(
        key in req
        for key in ["name", "password", "display_name", "description", "theme"]
    ):
        raise HttpException(
            "failed to decode the request body as json",
            BAD_REQUEST,
        )

    if req["name"] == "pipe":
        raise HttpException("the username 'pipe' is reserved", BAD_REQUEST)

    hashed_password = bcrypt.hashpw(
        req["password"].encode(), bcrypt.gensalt(rounds=Settings.BCRYPT_DEFAULT_COST)
    )

    user_model = models.UserModel(
        id=0,  # INSERT後に確定する
        name=req["name"],
        display_name=req["display_name"],
        description=req["description"],
        password=hashed_password.decode(),
    )

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "INSERT INTO users (name, display_name, description, password) VALUES (%s,%s,%s,%s)"
        c.execute(
            sql,
            [
                user_model.name,
                user_model.display_name,
                user_model.description,
                user_model.password,
            ],
        )
        user_id = c.lastrowid
        user_model.id = user_id

        sql = "INSERT INTO themes (user_id, dark_mode) VALUES (%s,%s)"
        c.execute(sql, [user_id, req["theme"]["dark_mode"]])

        if Settings.POWERDNS_ENABLED and Settings.POWERDNS_SUBDOMAIN_ADDRESS:
            # app.logger.info("add-record")
            result = subprocess.run(
                [
                    "pdnsutil",
                    "add-record",
                    "u.isucon.dev",
                    req["name"],
                    "A",
                    "0",
                    Settings.POWERDNS_SUBDOMAIN_ADDRESS,
                ],
                capture_output=True,
                text=True,
            )
            if result.returncode != 0:
                raise HttpException(result.stdout, INTERNAL_SERVER_ERROR)

        user = fill_user_response(c, user_model)
        return asdict(user), CREATED
    except DatabaseError as err:
        conn.rollback()
        app.logger.warn("failed to insert user: %s", err)
        raise HttpException("failed to insert user", INTERNAL_SERVER_ERROR)
    finally:
        conn.commit()
        conn.close()


@app.route("/api/login", methods=["POST"])
def login_handler() -> tuple[str, int]:
    req = get_request_json()
    if not req:
        raise HttpException(
            "failed to decode the request body as json",
            BAD_REQUEST,
        )

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM users WHERE name = %s"
        c.execute(sql, [req["username"]])
        row = c.fetchone()
        if row is None:
            raise HttpException("invalid username or password", UNAUTHORIZED)
        user = models.UserModel(**row)

        if not bcrypt.checkpw(req["password"].encode(), user.password.encode()):
            raise HttpException("invalid username or password", UNAUTHORIZED)

        session_end_at = datetime.now() + timedelta(hours=1)
        session_id = str(uuid.uuid4())

        session[Settings.DEFAULT_SESSION_ID_KEY] = session_id
        session[Settings.DEFAULT_USER_ID_KEY] = user.id
        session[Settings.DEFAULT_USER_NAME_KEY] = user.name
        session[Settings.DEFAULT_SESSION_EXPIRES_KEY] = int(session_end_at.timestamp())
        return "", OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


@app.route("/api/user/me", methods=["GET"])
def get_me_handler() -> tuple[dict[str, Any], int]:
    verify_user_session()

    user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
    if not user_id:
        raise HttpException("unauthorized", UNAUTHORIZED)

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM users WHERE id = %s"
        c.execute(sql, [user_id])
        row = c.fetchone()
        if not row:
            raise HttpException("failed to get user", INTERNAL_SERVER_ERROR)
        user_model = models.UserModel(**row)
        user = fill_user_response(c, user_model)
        return asdict(user), OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# フロントエンドで、配信予約のコラボレーターを指定する際に必要
@app.route("/api/user/<string:username>", methods=["GET"])
def get_user_handler(username: str) -> tuple[dict[str, Any], int]:
    verify_user_session()

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM users WHERE name = %s"
        c.execute(sql, [username])
        row = c.fetchone()
        if row is None:
            raise HttpException("not found", NOT_FOUND)
        user_model = models.UserModel(**row)

        user = fill_user_response(c, user_model)
        return asdict(user), OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


@app.route("/api/user/<string:username>/statistics", methods=["GET"])
def get_user_statistics_handler(username: str) -> tuple[dict[str, Any], int]:
    verify_user_session()

    # ユーザごとに、紐づく配信について、累計リアクション数、累計ライブコメント数、累計売上金額を算出
    # また、現在の合計視聴者数もだす

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM users WHERE name = %s"
        c.execute(sql, [username])
        stat_target_user: dict[str, Any] = c.fetchone()
        if stat_target_user is None:
            raise HttpException("not found user that has the given username", NOT_FOUND)

        # ランク算出
        sql = "SELECT * FROM users"
        c.execute(sql)
        rows = c.fetchall()
        if rows is None:
            raise HttpException("failed to get users", INTERNAL_SERVER_ERROR)
        users = [models.UserModel(**row) for row in rows]

        ranking = []
        for user in users:
            sql = """
                SELECT COUNT(*) FROM users u
                INNER JOIN livestreams l ON l.user_id = u.id
                INNER JOIN reactions r ON r.livestream_id = l.id
                WHERE u.id = %s
                """
            c.execute(sql, [user.id])
            row = c.fetchone()
            if not row:
                raise HttpException(
                    "failed to count reactions",
                    INTERNAL_SERVER_ERROR,
                )
            reactions = int(row["COUNT(*)"])

            sql = """
                SELECT IFNULL(SUM(l2.tip), 0) FROM users u
                INNER JOIN livestreams l ON l.user_id = u.id
                INNER JOIN livecomments l2 ON l2.livestream_id = l.id
                WHERE u.id = %s
                """
            c.execute(sql, [user.id])
            row = c.fetchone()
            if not row:
                raise HttpException(
                    "failed to count tips",
                    INTERNAL_SERVER_ERROR,
                )
            tips = int(row["IFNULL(SUM(l2.tip), 0)"])

            score = reactions + tips
            ranking.append(models.UserRankingEntry(username=user.name, score=score))
        ranking = sorted(ranking, key=lambda x: (x.score, x.username))

        rank = 1
        i = len(ranking) - 1
        while i >= 0:
            entry = ranking[i]
            if entry.username == username:
                break
            rank += 1
            i -= 1

        # リアクション数
        sql = """
            SELECT COUNT(*) FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.name = %s
        """
        c.execute(sql, [username])
        row = c.fetchone()
        if not row:
            raise HttpException(
                "failed to count total reactions",
                INTERNAL_SERVER_ERROR,
            )
        total_reactions = row["COUNT(*)"]

        # ライブコメント数、チップ合計
        total_livecomments = 0
        total_tip = 0
        sql = "SELECT * FROM livestreams WHERE user_id = %s"
        c.execute(sql, [stat_target_user["id"]])
        rows = c.fetchall()
        if rows is None:
            app.logger.error("livestreams livecomments")
            raise HttpException(
                "failed to get livestreams",
                INTERNAL_SERVER_ERROR,
            )
        livestreams = [models.LiveStreamModel(**row) for row in rows]

        for livestream in livestreams:
            sql = "SELECT * FROM livecomments WHERE livestream_id = %s"
            c.execute(sql, [livestream.id])
            livecomments = c.fetchall()
            if livecomments is None:
                app.logger.error("livecomments")
                raise HttpException(
                    "failed to get livecomments",
                    INTERNAL_SERVER_ERROR,
                )
            for livecomment in livecomments:
                total_tip += livecomment["tip"]
                total_livecomments += 1

        # 合計視聴者数
        viewers_count = 0

        for livestream in livestreams:
            sql = "SELECT COUNT(*) FROM livestream_viewers_history WHERE livestream_id = %s"
            c.execute(sql, [livestream.id])
            cnt = c.fetchone()
            if not cnt:
                raise HttpException(
                    "failed to get livestream_view_history",
                    INTERNAL_SERVER_ERROR,
                )
            viewers_count += int(cnt["COUNT(*)"])

        # お気に入り絵文字
        sql = """
            SELECT r.emoji_name
            FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.name = %s
            GROUP BY emoji_name
            ORDER BY COUNT(*) DESC, emoji_name DESC
            LIMIT 1
            """
        c.execute(sql, [username])
        row = c.fetchone()
        if not row:
            favorite_emoji = ""
        else:
            favorite_emoji = row["emoji_name"]

        statistics = models.UserStatistics(
            rank=rank,
            viewers_count=viewers_count,
            total_reactions=total_reactions,
            total_livecomments=total_livecomments,
            total_tip=total_tip,
            favorite_emoji=favorite_emoji,
        )

        return asdict(statistics), OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


@app.route("/api/user/<string:username>/icon", methods=["GET"])
def get_icon_handler(username: str) -> Response:
    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM users WHERE name = %s"
        c.execute(sql, [username])

        row = c.fetchone()
        if row is None:
            raise HttpException("user not found", INTERNAL_SERVER_ERROR)
        user = models.UserModel(**row)

        sql = "SELECT image FROM icons WHERE user_id = %s"
        c.execute(sql, [user.id])

        image = c.fetchone()
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()

    if not image:
        return send_file(
            Settings.FALLBACK_IMAGE, mimetype="image/jpeg", as_attachment=True
        )
    return send_file(
        io.BytesIO(image["image"]),
        mimetype="image/jpeg",
        as_attachment=True,
        download_name="icon.jpg",
    )


@app.route("/api/icon", methods=["POST"])
def post_icon_handler() -> tuple[dict[str, Any], int]:
    verify_user_session()

    user_id = session.get(Settings.DEFAULT_USER_ID_KEY)
    if not user_id:
        raise HttpException("unauthorized", UNAUTHORIZED)

    req = get_request_json()
    if not req or "image" not in req:
        raise HttpException(
            "failed to decode the request body as json",
            BAD_REQUEST,
        )
    new_icon = b64decode(req["image"])

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "DELETE FROM icons WHERE user_id = %s"
        c.execute(sql, [user_id])

        sql = "INSERT INTO icons (user_id, image) VALUES (%s, %s)"
        c.execute(sql, [user_id, new_icon])

        icon_id = c.lastrowid
        return {"id": icon_id}, CREATED
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# stats
# ライブコメント統計情報
@app.route("/api/livestream/<int:livestream_id>/statistics", methods=["GET"])
def get_livestream_statistics_handler(livestream_id: int) -> tuple[dict[str, Any], int]:
    verify_user_session()

    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)

        sql = "SELECT * FROM livestreams WHERE id = %s"
        c.execute(sql, [livestream_id])
        row = c.fetchone()
        if row is None:
            raise HttpException("cannot get stats of not found livestream", BAD_REQUEST)
        livestream = models.LiveStreamModel(**row)

        sql = "SELECT * FROM livestreams"
        c.execute(sql)
        rows = c.fetchall()
        if rows is None:
            raise HttpException("failed to get livestreams", INTERNAL_SERVER_ERROR)
        livestreams = [models.LiveStreamModel(**row) for row in rows]

        # ランク算出
        ranking = []
        for livestream in livestreams:
            sql = "SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON l.id = r.livestream_id WHERE l.id = %s"
            c.execute(sql, [livestream.id])
            row = c.fetchone()
            if row is None:
                raise HttpException(
                    "failed to get livestream",
                    INTERNAL_SERVER_ERROR,
                )
            reactions = int(row["COUNT(*)"])

            sql = "SELECT IFNULL(SUM(l2.tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l.id = l2.livestream_id WHERE l.id = %s"
            c.execute(sql, [livestream.id])
            row = c.fetchone()
            if row is None:
                raise HttpException("failed to count tips", INTERNAL_SERVER_ERROR)
            total_tips = int(row["IFNULL(SUM(l2.tip), 0)"])

            score = reactions + total_tips
            ranking.append(
                asdict(
                    models.LiveStreamRankingEntry(
                        livestream_id=livestream.id,
                        score=score,
                    )
                )
            )
        ranking = sorted(ranking, key=lambda x: (x["score"], x["livestream_id"]))

        rank = 1
        i = len(ranking) - 1
        while i >= 0:
            entry = ranking[i]
            if entry["livestream_id"] == livestream_id:
                break
            rank += 1
            i -= 1

        # 視聴者数算出
        sql = "SELECT COUNT(*) FROM livestreams l INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id WHERE l.id = %s"
        c.execute(sql, [livestream_id])
        row = c.fetchone()
        if row is None:
            raise HttpException(
                "failed to get viewers_count",
                INTERNAL_SERVER_ERROR,
            )
        viewers_count = int(row["COUNT(*)"])

        # 最大チップ額
        sql = "SELECT IFNULL(MAX(tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l2.livestream_id = l.id WHERE l.id = %s"
        c.execute(sql, [livestream_id])
        row = c.fetchone()
        if row is None:
            raise HttpException(
                "failed to get max_tip",
                INTERNAL_SERVER_ERROR,
            )
        max_tip = int(row["IFNULL(MAX(tip), 0)"])

        # リアクション数
        sql = "SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON r.livestream_id = l.id WHERE l.id = %s"
        c.execute(sql, [livestream_id])
        row = c.fetchone()
        if row is None:
            raise HttpException(
                "failed to get total_reactions",
                INTERNAL_SERVER_ERROR,
            )
        total_reactions = int(row["COUNT(*)"])

        # スパム報告数
        sql = "SELECT COUNT(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = %s"
        c.execute(sql, [livestream_id])
        row = c.fetchone()
        if row is None:
            raise HttpException(
                "failed to get total_reports",
                INTERNAL_SERVER_ERROR,
            )
        total_reports = int(row["COUNT(*)"])

        user_statistics = models.LiveStreamStatistics(
            rank=rank,
            viewers_count=viewers_count,
            total_reactions=total_reactions,
            total_reports=total_reports,
            max_tip=max_tip,
        )
        return asdict(user_statistics), OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


# 課金情報
@app.route("/api/payment", methods=["GET"])
def get_payment_result() -> tuple[dict[str, Any], int]:
    conn = engine.raw_connection()

    try:
        conn.start_transaction()
        c = conn.cursor(dictionary=True)
        sql = "SELECT IFNULL(SUM(tip), 0) AS sum_tip FROM livecomments"
        c.execute(sql)
        row = c.fetchone()
        if not row:
            raise HttpException("failed to count total tip", INTERNAL_SERVER_ERROR)
        return asdict(models.PaymentResult(total_tip=row)), OK
    except DatabaseError as err:
        conn.rollback()
        raise err
    finally:
        conn.commit()
        conn.close()


def verify_user_session() -> None:
    sess = session.get(Settings.DEFAULT_SESSION_ID_KEY)
    if not sess:
        raise HttpException("invalid session", INTERNAL_SERVER_ERROR)

    session_expires = session.get(Settings.DEFAULT_SESSION_EXPIRES_KEY)
    if not session_expires:
        raise HttpException("forbidden", FORBIDDEN)

    now = datetime.now()
    if int(now.timestamp()) > session_expires:
        raise HttpException("session has expired", UNAUTHORIZED)

    return


def fill_livecomment_response(
    c: mysql.connector.cursor.MySQLCursorDict,
    livecomment_model: models.LiveCommentModel,
) -> models.LiveComment:
    sql = "SELECT * FROM users WHERE id = %s"
    c.execute(sql, [livecomment_model.user_id])
    row = c.fetchone()
    if not row:
        app.logger.error("failed to get comment_owner_model")
        raise HttpException("failed to get comment_owner_model", INTERNAL_SERVER_ERROR)
    comment_owner_model = models.UserModel(**row)

    comment_owner = fill_user_response(c, comment_owner_model)

    sql = "SELECT * FROM livestreams WHERE id = %s"
    c.execute(sql, [livecomment_model.livestream_id])
    row = c.fetchone()
    if not row:
        app.logger.error("failed to get livestream_model")
        raise HttpException("failed to get livestream_model", INTERNAL_SERVER_ERROR)

    livestream_model = models.LiveStreamModel(**row)

    livestream = fill_livestream_response(c, livestream_model)

    return models.LiveComment(
        id=livecomment_model.id,
        user=comment_owner,
        livestream=livestream,
        comment=livecomment_model.comment,
        tip=livecomment_model.tip,
        created_at=livecomment_model.created_at,
    )


def fill_reaction_response(
    c: mysql.connector.cursor.MySQLCursorDict,
    reaction_model: models.ReactionModel,
) -> models.Reaction:
    sql = "SELECT * FROM users WHERE id = %s"
    c.execute(sql, [reaction_model.user_id])
    row = c.fetchone()
    if not row:
        app.logger.error("failed to get user_model")
        raise HttpException("failed to get user_model", INTERNAL_SERVER_ERROR)
    user_model = models.UserModel(**row)

    user = fill_user_response(c, user_model)

    sql = "SELECT * FROM livestreams WHERE id = %s"
    c.execute(sql, [reaction_model.livestream_id])
    row = c.fetchone()
    if not row:
        app.logger.error("failed to get livestream_model")
        raise HttpException("livestream_model", INTERNAL_SERVER_ERROR)
    livestream_model = models.LiveStreamModel(**row)

    livestream = fill_livestream_response(c, livestream_model)

    return models.Reaction(
        id=reaction_model.id,
        user=user,
        livestream=livestream,
        emoji_name=reaction_model.emoji_name,
        created_at=reaction_model.created_at,
    )


def fill_livecomment_report_response(
    c: mysql.connector.cursor.MySQLCursorDict,
    report_model: models.LiveCommentReportModel,
) -> models.LiveCommentReport:
    sql = "SELECT * FROM users WHERE id = %s"
    c.execute(sql, [report_model.user_id])
    row = c.fetchone()
    if not row:
        raise HttpException("failed to get reporter user", INTERNAL_SERVER_ERROR)
    reporter_model = models.UserModel(**row)

    reporter = fill_user_response(c, reporter_model)

    sql = "SELECT * FROM livecomments WHERE id = %s"
    c.execute(sql, [report_model.livecomment_id])
    row = c.fetchone()
    if row is None:
        raise HttpException("failed to get livecomment", INTERNAL_SERVER_ERROR)
    livecomment_model = models.LiveCommentModel(**row)

    livecomment = fill_livecomment_response(c, livecomment_model)

    return models.LiveCommentReport(
        id=report_model.id,
        reporter=reporter,
        livecomment=livecomment,
        created_at=report_model.created_at,
    )


def fill_livestream_response(
    c: mysql.connector.cursor.MySQLCursorDict,
    livestream_model: models.LiveStreamModel,
) -> models.LiveStream:
    sql = "SELECT * FROM users WHERE id = %s"
    c.execute(sql, [livestream_model.user_id])
    row = c.fetchone()
    if not row:
        raise HttpException("failed to get owner_model", INTERNAL_SERVER_ERROR)
    owner_model = models.UserModel(**row)

    owner = fill_user_response(c, owner_model)

    sql = "SELECT * FROM livestream_tags WHERE livestream_id = %s"
    c.execute(sql, [livestream_model.id])
    rows = c.fetchall()

    tags = []
    for row in rows:
        livestream_tag = models.LiveStreamTagModel(**row)
        sql = "SELECT * FROM tags WHERE id = %s"
        c.execute(sql, [livestream_tag.tag_id])
        tag_row = c.fetchone()
        if not tag_row:
            raise HttpException("failed to get tags", INTERNAL_SERVER_ERROR)
        tag = models.Tag(**tag_row)
        tags.append(tag)

    livestream = models.LiveStream(
        id=livestream_model.id,
        owner=owner,
        title=livestream_model.title,
        tags=tags,
        description=livestream_model.description,
        playlist_url=livestream_model.playlist_url,
        thumbnail_url=livestream_model.thumbnail_url,
        start_at=livestream_model.start_at,
        end_at=livestream_model.end_at,
    )
    return livestream


def fill_user_response(
    c: mysql.connector.cursor.MySQLCursorDict, user_model: models.UserModel
) -> models.User:
    sql = "SELECT * FROM themes WHERE user_id = %s"
    c.execute(sql, [user_model.id])
    row = c.fetchone()
    if row is None:
        raise HttpException("not found", NOT_FOUND)
    theme_model = models.ThemeModel(**row)

    sql = "SELECT image FROM icons WHERE user_id = %s"
    c.execute(sql, [user_model.id])
    image_row = c.fetchone()
    if not image_row:
        image = open(Settings.FALLBACK_IMAGE, "rb").read()
    else:
        image = io.BytesIO(image_row["image"]).getvalue()
    icon_hash = hashlib.sha256(image).hexdigest()

    user = models.User(
        id=user_model.id,
        name=user_model.name,
        display_name=user_model.display_name,
        description=user_model.description,
        theme=models.Theme(id=theme_model.id, dark_mode=theme_model.dark_mode),
        icon_hash=icon_hash,
    )

    return user


# Content-Type付けてこないクライアントからのJSONリクエストボディを
# いい感じに受け取れるようにするやつ
def get_request_json() -> Any:
    return json.loads(request.get_data().decode())


class HttpException(Exception):
    status_code = 500

    def __init__(self, message: str, status_code: int):
        Exception.__init__(self)
        self.message = message
        self.status_code = status_code

    def get_response(self) -> tuple[dict[str, Any], int]:
        return {
            "error": f"code={self.status_code}, message={self.message}",
            "message": self.message,
        }, self.status_code


@app.errorhandler(HttpException)
def handle_http_exception(error: Any) -> Any:
    return error.get_response()


if __name__ == "__main__":
    if not Settings.POWERDNS_SUBDOMAIN_ADDRESS:
        app.logger.critical("environ POWERDNS_SUBDOMAIN_ADDRESS must be provided")
        sys.exit(1)

    global engine
    engine = create_engine(
        f"mysql+mysqlconnector://{Settings.DB_USER}:{Settings.DB_PASSWORD}@{Settings.DB_HOST}:{Settings.DB_PORT}/{Settings.DB_NAME}"
    )
    app.secret_key = Settings.SESSION_SECRET_KEY
    app.config["SESSION_COOKIE_DOMAIN"] = Settings.SESSION_COOKIE_DOMAIN
    app.config["SESSION_COOKIE_PATH"] = Settings.SESSION_COOKIE_PATH
    app.config["PERMANENT_SESSION_LIFETIME"] = Settings.PERMANENT_SESSION_LIFETIME
    app.run(host="0.0.0.0", port=Settings.LISTEN_PORT, debug=True, threaded=True)
