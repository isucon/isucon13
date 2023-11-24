import { Context } from 'hono'
import { ResultSetHeader, RowDataPacket } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { defaultUserIDKey } from '../contants'
import {
  ReactionResponse,
  fillReactionResponse,
} from '../utils/fill-reaction-response'
import { throwErrorWith } from '../utils/throw-error-with'
import { ReactionsModel } from '../types/models'
import { atoi } from '../utils/integer'

// GET /api/livestream/:livestream_id/reaction
export const getReactionsHandler = [
  verifyUserSessionMiddleware,
  async (
    c: Context<HonoEnvironment, '/api/livestream/:livestream_id/reaction'>,
  ) => {
    const livestreamId = atoi(c.req.param('livestream_id'))
    if (livestreamId === false) {
      return c.text('livestream_id in path must be integer', 400)
    }

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      let query =
        'SELECT * FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC'
      const limit = c.req.query('limit')
      if (limit) {
        const limitNumber = atoi(limit)
        if (limitNumber === false) {
          return c.text('limit query parameter must be integer', 400)
        }
        query += ` LIMIT ${limitNumber}`
      }

      const [reactions] = await conn
        .query<(ReactionsModel & RowDataPacket)[]>(query, [livestreamId])
        .catch(throwErrorWith('failed to get reactions'))

      const reactionResponses: ReactionResponse[] = []
      for (const reaction of reactions) {
        const reactionResponse = await fillReactionResponse(
          conn,
          reaction,
          c.get('runtime').fallbackUserIcon,
        ).catch(throwErrorWith('failed to fill reaction'))

        reactionResponses.push(reactionResponse)
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(reactionResponses)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// POST /api/livestream/:livestream_id/reaction
export const postReactionHandler = [
  verifyUserSessionMiddleware,
  async (
    c: Context<HonoEnvironment, '/api/livestream/:livestream_id/reaction'>,
  ) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = atoi(c.req.param('livestream_id'))
    if (livestreamId === false) {
      return c.text('livestream_id in path must be integer', 400)
    }

    const body = await c.req.json<{ emoji_name: string }>()

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      const now = Date.now()
      const [{ insertId: reactionId }] = await conn
        .query<ResultSetHeader>(
          'INSERT INTO reactions (user_id, livestream_id, emoji_name, created_at) VALUES (?, ?, ?, ?)',
          [userId, livestreamId, body.emoji_name, now],
        )
        .catch(throwErrorWith('failed to insert reaction'))

      const reactionResponse = await fillReactionResponse(
        conn,
        {
          id: reactionId,
          emoji_name: body.emoji_name,
          user_id: userId,
          livestream_id: livestreamId,
          created_at: now,
        },
        c.get('runtime').fallbackUserIcon,
      ).catch(throwErrorWith('failed to fill reaction'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(reactionResponse, 201)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]
