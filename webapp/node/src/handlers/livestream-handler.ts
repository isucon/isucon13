import { Hono } from 'hono'
import { RowDataPacket, ResultSetHeader } from 'mysql2/promise'
import { ApplicationDeps, HonoEnvironment } from '../types'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { defaultUserIDKey } from '../contants'
import { makeLivestreamResponse } from '../utils/make-livestream-response'

export const livestreamHandler = (deps: ApplicationDeps) => {
  const handler = new Hono<HonoEnvironment>()

  handler.post(
    '/api/livestream/reservation',
    verifyUserSessionMiddleware,
    async (c) => {
      const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
      const body = await c.req.json<{
        tags: number[]
        title: string
        description: string
        playlist_url: string
        thumbnail_url: string
        start_at: number
        end_at: number
      }>()

      await deps.connection.beginTransaction()

      // 2024/04/01からの１年間の期間内であるかチェック
      const termStartAt = Date.UTC(2024, 3, 1) // NOTE: month is 0-indexed
      const termEndAt = Date.UTC(2025, 3, 1) // NOTE: month is 0-indexed
      const reserveStartAt = body.start_at
      const reserveEndAt = body.end_at
      if (reserveStartAt >= termEndAt && reserveEndAt <= termStartAt) {
        await deps.connection.rollback()
        return c.text('bad reservation time range', 400)
      }

      const slots = await deps.connection
        .query<RowDataPacket[]>(
          'SELECT * FROM reservation_slots WHERE start_at >= ? AND end_at <= ?',
          [reserveStartAt, reserveEndAt],
        )
        .then(([results]) => ({ ok: true, data: results }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!slots.ok) {
        await deps.connection.rollback()
        return c.text('failed to get reservation_slots', 500)
      }

      for (const slot of slots.data) {
        const count = await deps.connection
          .query<RowDataPacket[]>(
            'SELECT slot FROM reservation_slots WHERE start_at = ? AND end_at = ?',
            [slot.start_at, slot.end_at],
          )
          .then(([result]) => ({ ok: true, data: result.length }) as const)
          .catch((error) => ({ ok: false, error }) as const)
        if (!count.ok) {
          await deps.connection.rollback()
          return c.text('failed to get reservation_slots', 500)
        }
        console.log(
          `${new Date(slot.start_at).toISOString()} ~ ${new Date(
            slot.end_at,
          ).toISOString()} 予約枠の残数 - ${slot.slot}`,
        )
        if (count.data < 1) {
          return c.text(
            `予約区間 ${new Date(slot.start_at).toISOString()} ~ ${new Date(
              slot.end_at,
            ).toISOString()} が予約できません`,
            400,
          )
        }
      }

      await deps.connection.query(
        'UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= ? AND end_at <= ?',
        [reserveStartAt, reserveEndAt],
      )
      const livestreamsResult = await deps.connection
        .execute<ResultSetHeader>(
          'INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(?, ?, ?, ?, ?, ?, ?)',
          [
            userId,
            body.title,
            body.description,
            body.playlist_url,
            body.thumbnail_url,
            body.start_at,
            body.end_at,
          ],
        )
        .then(([result]) => ({ ok: true, data: result }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!livestreamsResult.ok) {
        await deps.connection.rollback()
        return c.text('failed to insert livestream', 500)
      }

      const livestreamID = livestreamsResult.data.insertId

      // タグ追加
      for (const tagId of body.tags) {
        try {
          await deps.connection.execute(
            'INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (?, ?)',
            [livestreamID, tagId],
          )
        } catch {
          await deps.connection.rollback()
          return c.text(`failed to insert livestream tag`, 500)
        }
      }

      const response = await makeLivestreamResponse(deps, {
        id: livestreamID,
        user_id: userId,
        title: body.title,
        description: body.description,
        playlist_url: body.playlist_url,
        thumbnail_url: body.thumbnail_url,
        start_at: body.start_at,
        end_at: body.end_at,
      })
      if (!response.ok) {
        await deps.connection.rollback()
        return c.text(response.error, 500)
      }

      try {
        await deps.connection.commit()
      } catch {
        await deps.connection.rollback()
        return c.text('failed to commit', 500)
      }

      return c.json(response.data, 201)
    },
  )

  return handler
}
