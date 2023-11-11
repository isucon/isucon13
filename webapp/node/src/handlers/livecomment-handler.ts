import { Hono } from 'hono'
import { RowDataPacket, ResultSetHeader } from 'mysql2/promise'
import { ApplicationDeps, HonoEnvironment } from '../types'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { defaultUserIDKey } from '../contants'
import {
  LivecommentResponse,
  makeLivecommentResponse,
} from '../utils/make-livecomment-response'
import { makeLivecommentReportResponse } from '../utils/make-livecomment-report-response'

export const livecommentHandler = (deps: ApplicationDeps) => {
  const handler = new Hono<HonoEnvironment>()

  handler.post(
    '/api/livestream/:livestream_id/livecomment',
    verifyUserSessionMiddleware,
    async (c) => {
      const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
      const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
      if (Number.isNaN(livestreamId)) {
        return c.json({
          ok: false,
          error: 'livestream_id in path must be integer',
        })
      }

      const body = await c.req.json<{ comment: string; tip: number }>()

      await deps.connection.beginTransaction()

      // スパム判定
      const ngwords = await deps.connection
        .query<RowDataPacket[]>(
          'SELECT id, user_id, livestream_id, word FROM ng_words',
        )
        .then(([results]) => ({ ok: true, data: results }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!ngwords.ok) {
        await deps.connection.rollback()
        return c.text('failed to get NG words', 500)
      }

      for (const ngword of ngwords.data) {
        const query = `
          SELECT COUNT(*)
          FROM
          (SELECT ? AS text) AS texts
          INNER JOIN
          (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
          ON texts.text LIKE patterns.pattern;
        `
        const result = await deps.connection
          .query<RowDataPacket[]>(query, [body.comment, ngword.word])
          .then(([[result]]) => ({ ok: true, data: result }) as const)
          .catch((error) => ({ ok: false, error }) as const)
        if (!result.ok) {
          await deps.connection.rollback()
          return c.text('failed to get hitspam', 500)
        }
        console.log(result.data)
        console.log(
          `[hitSpam=${result.data['COUNT(*)']}] comment = ${body.comment}`,
        )
        if (result.data['COUNT(*)'] > 0) {
          await deps.connection.rollback()
          return c.text('コメントがスパム判定されました', 400)
        }
      }
      const livecommentResult = await deps.connection
        .query<ResultSetHeader>(
          'INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?, ?, ?, ?, ?)',
          [userId, livestreamId, body.comment, body.tip, Date.now()],
        )
        .then(([result]) => ({ ok: true, data: result.insertId }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!livecommentResult.ok) {
        await deps.connection.rollback()
        return c.text('failed to insert livecomment', 500)
      }

      const livecommentResponse = await makeLivecommentResponse(deps, {
        id: livecommentResult.data,
        user_id: userId,
        livestream_id: livestreamId,
        comment: body.comment,
        tip: body.tip,
        created_at: Date.now(),
      })
      if (!livecommentResponse.ok) {
        await deps.connection.rollback()
        return c.text(livecommentResponse.error, 500)
      }

      try {
        await deps.connection.commit()
      } catch {
        await deps.connection.rollback()
        return c.text('failed to commit', 500)
      }

      return c.json(livecommentResponse.data, 201)
    },
  )

  handler.get(
    '/api/livestream/:livestream_id/livecomment',
    verifyUserSessionMiddleware,
    async (c) => {
      const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
      if (Number.isNaN(livestreamId)) {
        return c.text('livestream_id in path must be integer', 400)
      }

      await deps.connection.beginTransaction()

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
      const livecomments = await deps.connection
        .query<RowDataPacket[]>(query, [livestreamId])
        .then(([results]) => ({ ok: true, data: results }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!livecomments.ok) {
        await deps.connection.rollback()
        return c.text('failed to get livecomments', 500)
      }

      const livecommnetResponses: LivecommentResponse[] = []
      for (const livecomment of livecomments.data) {
        const livecommentResponse = await makeLivecommentResponse(deps, {
          id: livecomment.id,
          user_id: livecomment.user_id,
          livestream_id: livecomment.livestream_id,
          comment: livecomment.comment,
          tip: livecomment.tip,
          created_at: livecomment.created_at,
        })
        if (!livecommentResponse.ok) {
          await deps.connection.rollback()
          return c.text(livecommentResponse.error, 500)
        }
        livecommnetResponses.push(livecommentResponse.data)
      }

      try {
        await deps.connection.commit()
      } catch {
        await deps.connection.rollback()
        return c.text('failed to commit', 500)
      }

      return c.json(livecommnetResponses)
    },
  )

  handler.post(
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

      await deps.connection.beginTransaction()

      const now = Date.now()

      const livecommentReportResult = await deps.connection
        .query<ResultSetHeader>(
          'INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id, created_at) VALUES (?, ?, ?, ?)',
          [userId, livestreamId, livecommentId, now],
        )
        .then(([result]) => ({ ok: true, data: result.insertId }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!livecommentReportResult.ok) {
        await deps.connection.rollback()
        return c.text('failed to insert livecomment report', 500)
      }

      const livecommentReportResponse = await makeLivecommentReportResponse(
        deps,
        {
          id: livecommentReportResult.data,
          user_id: userId,
          livestream_id: livestreamId,
          livecomment_id: livecommentId,
          created_at: now,
        },
      )
      if (!livecommentReportResponse.ok) {
        await deps.connection.rollback()
        return c.text(livecommentReportResponse.error, 500)
      }

      try {
        await deps.connection.commit()
      } catch {
        await deps.connection.rollback()
        return c.text('failed to commit', 500)
      }

      return c.json(livecommentReportResponse.data, 201)
    },
  )

  return handler
}
