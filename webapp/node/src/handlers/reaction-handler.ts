import { Hono } from 'hono'
import { ResultSetHeader, RowDataPacket } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { defaultUserIDKey } from '../contants'
import {
  ReactionResponse,
  makeReactionResponse,
} from '../utils/make-reaction-response'
import { throwErrorWith } from '../utils/throw-error-with'
import { ReactionsModel } from '../types/models'

export const reactionHandler = new Hono<HonoEnvironment>()

reactionHandler.post(
  '/api/livestream/:livestream_id/reaction',
  verifyUserSessionMiddleware,
  async (c) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    if (Number.isInteger(c.req.param('livestream_id'))) {
      return c.text('livestream_id in path must be integer', 400)
    }
    const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)

    const body = await c.req.json<{ emoji_name: string }>()

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      const now = Date.now()
      const [{ insertId }] = await conn
        .query<ResultSetHeader>(
          'INSERT INTO reactions (user_id, livestream_id, emoji_name, created_at) VALUES (?, ?, ?, ?)',
          [userId, livestreamId, body.emoji_name, now],
        )
        .catch(throwErrorWith('failed to insert reaction'))

      const reactionResponse = await makeReactionResponse(conn, {
        id: insertId,
        emoji_name: body.emoji_name,
        user_id: userId,
        livestream_id: livestreamId,
        created_at: now,
      })

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(reactionResponse, 201)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

reactionHandler.get(
  '/api/livestream/:livestream_id/reaction',
  verifyUserSessionMiddleware,
  async (c) => {
    if (Number.isInteger(c.req.param('livestream_id'))) {
      return c.text('livestream_id in path must be integer', 400)
    }
    const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      let query =
        'SELECT * FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC'
      const limit = c.req.query('limit')
      if (limit) {
        if (Number.isInteger(limit)) {
          return c.text('limit query parameter must be integer', 400)
        }
        const limitNumber = Number.parseInt(limit, 10)
        query += ` LIMIT ${limitNumber}`
      }

      const [reactions] = await conn
        .query<(ReactionsModel & RowDataPacket)[]>(query, [livestreamId])
        .catch(throwErrorWith('failed to get reactions'))

      const reactionResponses: ReactionResponse[] = []
      for (const reaction of reactions) {
        const reactionResponse = await makeReactionResponse(conn, reaction)

        reactionResponses.push(reactionResponse)
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(reactionResponses)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)
