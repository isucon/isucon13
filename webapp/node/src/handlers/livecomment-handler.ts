import { Context } from 'hono'
import { RowDataPacket, ResultSetHeader } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { defaultUserIDKey } from '../contants'
import {
  LivecommentResponse,
  fillLivecommentResponse,
} from '../utils/fill-livecomment-response'
import { fillLivecommentReportResponse } from '../utils/fill-livecomment-report-response'
import {
  LivecommentsModel,
  LivestreamsModel,
  NgWordsModel,
} from '../types/models'
import { throwErrorWith } from '../utils/throw-error-with'
import { atoi } from '../utils/integer'

// GET /api/livestream/:livestream_id/livecomment
export const getLivecommentsHandler = [
  verifyUserSessionMiddleware,
  async (
    c: Context<HonoEnvironment, '/api/livestream/:livestream_id/livecomment'>,
  ) => {
    const livestreamId = atoi(c.req.param('livestream_id'))
    if (livestreamId === false) {
      return c.text('livestream_id in path must be integer', 400)
    }

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      let query =
        'SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC'
      const limit = c.req.query('limit')
      if (limit) {
        const limitNumber = atoi(limit)
        if (limitNumber === false) {
          return c.text('limit query parameter must be integer', 400)
        }
        query += ` LIMIT ${limitNumber}`
      }
      const [livecomments] = await conn
        .query<(LivecommentsModel & RowDataPacket)[]>(query, [livestreamId])
        .catch(throwErrorWith('failed to get livecomments'))

      const livecommnetResponses: LivecommentResponse[] = []
      for (const livecomment of livecomments) {
        const livecommentResponse = await fillLivecommentResponse(
          conn,
          livecomment,
          c.get('runtime').fallbackUserIcon,
        ).catch(throwErrorWith('failed to fill livecomment'))
        livecommnetResponses.push(livecommentResponse)
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))
      return c.json(livecommnetResponses)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// GET /api/livestream/:livestream_id/ngwords
export const getNgwords = [
  verifyUserSessionMiddleware,
  async (
    c: Context<HonoEnvironment, '/api/livestream/:livestream_id/ngwords'>,
  ) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = atoi(c.req.param('livestream_id'))
    if (livestreamId === false) {
      return c.text('livestream_id in path must be integer', 400)
    }

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      const [ngwords] = await conn
        .query<(NgWordsModel & RowDataPacket)[]>(
          'SELECT * FROM ng_words WHERE user_id = ? AND livestream_id = ? ORDER BY created_at DESC',
          [userId, livestreamId],
        )
        .catch(throwErrorWith('failed to get NG words'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(
        ngwords.map((ngword) => ({
          id: ngword.id,
          user_id: ngword.user_id,
          livestream_id: ngword.livestream_id,
          word: ngword.word,
          created_at: ngword.created_at,
        })),
      )
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// POST /api/livestream/:livestream_id/livecomment
export const postLivecommentHandler = [
  verifyUserSessionMiddleware,
  async (
    c: Context<HonoEnvironment, '/api/livestream/:livestream_id/livecomment'>,
  ) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = atoi(c.req.param('livestream_id'))
    if (livestreamId === false) {
      return c.text('livestream_id in path must be integer', 400)
    }

    const body = await c.req.json<{ comment: string; tip: number }>()

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()
    try {
      const [[livestream]] = await conn
        .execute<(LivestreamsModel & RowDataPacket)[]>(
          `SELECT * FROM livestreams WHERE id = ?`,
          [livestreamId],
        )
        .catch(throwErrorWith('failed to get livestream'))
      if (!livestream) {
        await conn.rollback()
        return c.text('livestream not found', 404)
      }

      // スパム判定
      const [ngwords] = await conn
        .query<
          (Pick<NgWordsModel, 'id' | 'user_id' | 'livestream_id' | 'word'> &
            RowDataPacket)[]
        >(
          'SELECT id, user_id, livestream_id, word FROM ng_words WHERE user_id = ? AND livestream_id = ?',
          [livestream.user_id, livestreamId],
        )
        .catch(throwErrorWith('failed to get NG words'))

      for (const ngword of ngwords) {
        const [[{ 'COUNT(*)': hitSpam }]] = await conn
          .query<({ 'COUNT(*)': number } & RowDataPacket)[]>(
            `
              SELECT COUNT(*)
              FROM
              (SELECT ? AS text) AS texts
              INNER JOIN
              (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
              ON texts.text LIKE patterns.pattern;
            `,
            [body.comment, ngword.word],
          )
          .catch(throwErrorWith('failed to get hitspam'))

        console.info(`[hitSpam=${hitSpam}] comment = ${body.comment}`)
        if (hitSpam >= 1) {
          await conn.rollback()
          return c.text('このコメントがスパム判定されました', 400)
        }
      }
      const now = Date.now()
      const [{ insertId: livecommentId }] = await conn
        .query<ResultSetHeader>(
          'INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?, ?, ?, ?, ?)',
          [userId, livestreamId, body.comment, body.tip, now],
        )
        .catch(throwErrorWith('failed to insert livecomment'))

      const livecommentResponse = await fillLivecommentResponse(
        conn,
        {
          id: livecommentId,
          user_id: userId,
          livestream_id: livestreamId,
          comment: body.comment,
          tip: body.tip,
          created_at: now,
        },
        c.get('runtime').fallbackUserIcon,
      ).catch(throwErrorWith('failed to fill livecomment'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(livecommentResponse, 201)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// POST /api/livestream/:livestream_id/livecomment/:livecomment_id/report
export const reportLivecommentHandler = [
  verifyUserSessionMiddleware,
  async (
    c: Context<
      HonoEnvironment,
      '/api/livestream/:livestream_id/livecomment/:livecomment_id/report'
    >,
  ) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = atoi(c.req.param('livestream_id'))
    if (livestreamId === false) {
      return c.text('livestream_id in path must be integer', 400)
    }
    const livecommentId = atoi(c.req.param('livecomment_id'))
    if (livecommentId === false) {
      return c.text('livecomment_id in path must be integer', 400)
    }

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      const now = Date.now()

      const [[livestream]] = await conn
        .execute<(LivestreamsModel & RowDataPacket)[]>(
          `SELECT * FROM livestreams WHERE id = ?`,
          [livestreamId],
        )
        .catch(throwErrorWith('failed to get livestream'))
      if (!livestream) {
        await conn.rollback()
        return c.text('livestream not found', 404)
      }

      const [[livecomment]] = await conn
        .execute<(LivecommentsModel & RowDataPacket)[]>(
          `SELECT * FROM livecomments WHERE id = ?`,
          [livecommentId],
        )
        .catch(throwErrorWith('failed to get livecomment'))
      if (!livecomment) {
        await conn.rollback()
        return c.text('livecomment not found', 404)
      }

      const [{ insertId: reportId }] = await conn
        .query<ResultSetHeader>(
          'INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id, created_at) VALUES (?, ?, ?, ?)',
          [userId, livestreamId, livecommentId, now],
        )
        .catch(throwErrorWith('failed to insert livecomment report'))

      const livecommentReportResponse = await fillLivecommentReportResponse(
        conn,
        {
          id: reportId,
          user_id: userId,
          livestream_id: livestreamId,
          livecomment_id: livecommentId,
          created_at: now,
        },
        c.get('runtime').fallbackUserIcon,
      ).catch(throwErrorWith('failed to fill livecomment report'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(livecommentReportResponse, 201)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// POST /api/livestream/:livestream_id/moderate
export const moderateHandler = [
  verifyUserSessionMiddleware,
  async (
    c: Context<HonoEnvironment, '/api/livestream/:livestream_id/moderate'>,
  ) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = atoi(c.req.param('livestream_id'))
    if (livestreamId === false) {
      return c.text('livestream_id in path must be integer', 400)
    }

    const body = await c.req.json<{ ng_word: string }>()

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      // 配信者自身の配信に対するmoderateなのかを検証
      const [ownedLivestreams] = await conn
        .query<(LivecommentsModel & RowDataPacket)[]>(
          'SELECT * FROM livestreams WHERE id = ? AND user_id = ?',
          [livestreamId, userId],
        )
        .catch(throwErrorWith('failed to get livestreams'))

      if (ownedLivestreams.length === 0) {
        await conn.rollback()
        return c.text(
          "A streamer can't moderate livestreams that other streamers own",
          400,
        )
      }

      const [{ insertId: wordId }] = await conn
        .query<ResultSetHeader>(
          'INSERT INTO ng_words(user_id, livestream_id, word, created_at) VALUES (?, ?, ?, ?)',
          [userId, livestreamId, body.ng_word, Date.now()],
        )
        .catch(throwErrorWith('failed to insert new NG word'))

      const [ngwords] = await conn
        .query<(NgWordsModel & RowDataPacket)[]>(
          'SELECT * FROM ng_words WHERE livestream_id = ?',
          [livestreamId],
        )
        .catch(throwErrorWith('failed to get NG words'))

      // NGワードにヒットする過去の投稿も全削除する
      for (const ngword of ngwords) {
        // ライブコメント一覧取得
        const [livecomments] = await conn
          .query<(LivecommentsModel & RowDataPacket)[]>(
            'SELECT * FROM livecomments',
          )
          .catch(throwErrorWith('failed to get livecomments'))

        for (const livecomment of livecomments) {
          await conn
            .query(
              `
                DELETE FROM livecomments
                WHERE
                id = ? AND
                livestream_id = ? AND
                (SELECT COUNT(*)
                FROM
                (SELECT ? AS text) AS texts
                INNER JOIN
                (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
                ON texts.text LIKE patterns.pattern) >= 1;
              `,
              [livecomment.id, livestreamId, livecomment.comment, ngword.word],
            )
            .catch(
              throwErrorWith(
                'failed to delete old livecomments that hit spams',
              ),
            )
        }
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json({ word_id: wordId }, 201)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]
