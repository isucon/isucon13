import { Hono } from 'hono'
import { RowDataPacket, ResultSetHeader } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { defaultUserIDKey } from '../contants'
import {
  LivecommentResponse,
  makeLivecommentResponse,
} from '../utils/make-livecomment-response'
import { makeLivecommentReportResponse } from '../utils/make-livecomment-report-response'
import { LivecommentsModel, NgWordsModel } from '../types/models'
import { throwErrorWith } from '../utils/throw-error-with'

export const livecommentHandler = new Hono<HonoEnvironment>()

livecommentHandler.post(
  '/api/livestream/:livestream_id/livecomment',
  verifyUserSessionMiddleware,
  async (c) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
    if (Number.isNaN(livestreamId)) {
      return c.text('livestream_id in path must be integer', 400)
    }

    const body = await c.req.json<{ comment: string; tip: number }>()

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()
    try {
      // スパム判定
      const [ngwords] = await conn
        .query<
          (Pick<NgWordsModel, 'id' | 'user_id' | 'livestream_id' | 'word'> &
            RowDataPacket)[]
        >('SELECT id, user_id, livestream_id, word FROM ng_words')
        .catch(throwErrorWith('failed to get ngwords'))

      for (const ngword of ngwords) {
        const query = `
            SELECT COUNT(*)
            FROM
            (SELECT ? AS text) AS texts
            INNER JOIN
            (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
            ON texts.text LIKE patterns.pattern;
          `
        const [[{ 'COUNT(*)': hitSpam }]] = await conn
          .query<({ 'COUNT(*)': number } & RowDataPacket)[]>(query, [
            body.comment,
            ngword.word,
          ])
          .catch(throwErrorWith('failed to get hitspam'))

        console.log(`[hitSpam=${hitSpam}] comment = ${body.comment}`)
        if (hitSpam > 0) {
          await conn.rollback()
          return c.text('コメントがスパム判定されました', 400)
        }
      }
      const [{ insertId }] = await conn
        .query<ResultSetHeader>(
          'INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?, ?, ?, ?, ?)',
          [userId, livestreamId, body.comment, body.tip, Date.now()],
        )
        .catch(throwErrorWith('failed to insert livecomment'))

      const livecommentResponse = await makeLivecommentResponse(conn, {
        id: insertId,
        user_id: userId,
        livestream_id: livestreamId,
        comment: body.comment,
        tip: body.tip,
        created_at: Date.now(),
      })

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(livecommentResponse, 201)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

livecommentHandler.get(
  '/api/livestream/:livestream_id/livecomment',
  verifyUserSessionMiddleware,
  async (c) => {
    const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
    if (Number.isNaN(livestreamId)) {
      return c.text('livestream_id in path must be integer', 400)
    }

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      let query =
        'SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC'
      const limit = c.req.query('limit')
      if (limit) {
        const limitNumber = Number.parseInt(limit, 10)
        if (Number.isNaN(limitNumber)) {
          return c.text('limit query must be integer', 400)
        }
        query += ` LIMIT ${limitNumber}`
      }
      const [livecomments] = await conn
        .query<(LivecommentsModel & RowDataPacket)[]>(query, [livestreamId])
        .catch(throwErrorWith('failed to get livecomments'))

      const livecommnetResponses: LivecommentResponse[] = []
      for (const livecomment of livecomments) {
        const livecommentResponse = await makeLivecommentResponse(conn, {
          id: livecomment.id,
          user_id: livecomment.user_id,
          livestream_id: livecomment.livestream_id,
          comment: livecomment.comment,
          tip: livecomment.tip,
          created_at: livecomment.created_at,
        })
        livecommnetResponses.push(livecommentResponse)
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))
      return c.json(livecommnetResponses)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

livecommentHandler.post(
  '/api/livestream/:livestream_id/livecomment/:livecomment_id/report',
  verifyUserSessionMiddleware,
  async (c) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
    if (Number.isNaN(livestreamId)) {
      return c.text('livestream_id in path must be integer', 400)
    }
    const livecommentId = Number.parseInt(c.req.param('livecomment_id'), 10)
    if (Number.isNaN(livecommentId)) {
      return c.text('livecomment_id in path must be integer', 400)
    }

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      const now = Date.now()

      const [{ insertId }] = await conn
        .query<ResultSetHeader>(
          'INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id, created_at) VALUES (?, ?, ?, ?)',
          [userId, livestreamId, livecommentId, now],
        )
        .catch(throwErrorWith('failed to insert livecomment report'))

      const livecommentReportResponse = await makeLivecommentReportResponse(
        conn,
        {
          id: insertId,
          user_id: userId,
          livestream_id: livestreamId,
          livecomment_id: livecommentId,
          created_at: now,
        },
      )

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(livecommentReportResponse, 201)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

livecommentHandler.post(
  '/api/livestream/:livestream_id/moderate',
  verifyUserSessionMiddleware,
  async (c) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
    if (Number.isNaN(livestreamId)) {
      return c.text('livestream_id in path must be integer', 400)
    }

    const body = await c.req.json<{ ng_word: string }>()

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      // 配信者自身の配信に対するmoderateなのかを検証
      const [[livestream]] = await conn
        .query<(LivecommentsModel & RowDataPacket)[]>(
          'SELECT * FROM livestreams WHERE id = ? AND user_id = ?',
          [livestreamId, userId],
        )
        .catch(throwErrorWith('failed to get livestream'))

      if (!livestream) {
        await conn.rollback()
        return c.text(
          "A streamer can't moderate livestreams that other streamers own",
          404,
        )
      }

      const [{ insertId }] = await conn
        .query<ResultSetHeader>(
          'INSERT INTO ng_words(user_id, livestream_id, word, created_at) VALUES (?, ?, ?, ?)',
          [userId, livestreamId, body.ng_word, Date.now()],
        )
        .catch(throwErrorWith('failed to insert ngword'))

      const [ngwords] = await conn
        .query<(NgWordsModel & RowDataPacket)[]>('SELECT * FROM ng_words')
        .catch(throwErrorWith('failed to get ngwords'))

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
                  (SELECT COUNT(*)
                  FROM
                  (SELECT ? AS text) AS texts
                  INNER JOIN
                  (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
                  ON texts.text LIKE patterns.pattern) >= 1;
                `,
              [livecomment.comment, ngword.word],
            )
            .catch(throwErrorWith('failed to delete livecomment'))
        }
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json({ word_id: insertId }, 201)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

livecommentHandler.get(
  '/api/livestream/:livestream_id/ngwords',
  verifyUserSessionMiddleware,
  async (c) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
    if (Number.isNaN(livestreamId)) {
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
        .catch(throwErrorWith('failed to get ngwords'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(
        ngwords.map((ngword) => ({
          ng_word: ngword.word,
        })),
      )
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)
