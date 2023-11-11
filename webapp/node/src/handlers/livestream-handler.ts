import { Hono } from 'hono'
import { RowDataPacket, ResultSetHeader } from 'mysql2/promise'
import { ApplicationDeps, HonoEnvironment } from '../types'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { defaultUserIDKey } from '../contants'
import {
  LivestreamResponse,
  makeLivestreamResponse,
} from '../utils/make-livestream-response'
import {
  LivecommentReportResponse,
  makeLivecommentReportResponse,
} from '../utils/make-livecomment-report-response'

export const livestreamHandler = (deps: ApplicationDeps) => {
  const handler = new Hono<HonoEnvironment>()

  handler.get('/api/livestream/search', async (c) => {
    const tagName = c.req.query('tag')

    await deps.connection.beginTransaction()

    const livestreams: RowDataPacket[] = []
    if (tagName) {
      // タグによる取得
      const tagIds = await deps.connection
        .query<RowDataPacket[]>('SELECT id FROM tags WHERE name = ?', [tagName])
        .then(([results]) => ({ ok: true, data: results }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!tagIds.ok) {
        await deps.connection.rollback()
        return c.text('failed to get tags', 500)
      }

      const livestreamTags = await deps.connection
        .query<RowDataPacket[]>(
          'SELECT * FROM livestream_tags WHERE tag_id IN (?)',
          [tagIds.data.map((tag) => tag.id)],
        )
        .then(([results]) => ({ ok: true, data: results }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!livestreamTags.ok) {
        await deps.connection.rollback()
        return c.text('failed to get keyTaggedLivestreams', 500)
      }

      for (const livestreamTag of livestreamTags.data) {
        const livestream = await deps.connection
          .query<RowDataPacket[]>('SELECT * FROM livestreams WHERE id = ?', [
            livestreamTag.livestream_id,
          ])
          .then(([[result]]) => ({ ok: true, data: result }) as const)
          .catch((error) => ({ ok: false, error }) as const)
        if (!livestream.ok) {
          await deps.connection.rollback()
          return c.text('failed to get livestream', 500)
        }
        livestreams.push(livestream.data)
      }
    } else {
      // 検索条件なし
      let query = `SELECT * FROM livestreams`
      const limit = c.req.query('limit')
      if (limit) {
        const limitNumber = Number.parseInt(limit, 10)
        if (Number.isNaN(limitNumber)) {
          console.log('limitNumber', limitNumber, 'limit', limit)
          return c.text('limit query parameter must be integer', 400)
        }
        query += ` LIMIT ${limitNumber}`
      }

      const results = await deps.connection
        .query<RowDataPacket[]>(query)
        .then(([results]) => ({ ok: true, data: results }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!results.ok) {
        await deps.connection.rollback()
        return c.text('failed to get livestreams', 500)
      }
      livestreams.push(...results.data)
    }

    const livestreamResponses: LivestreamResponse[] = []
    for (const livestream of livestreams) {
      const livestreamResponse = await makeLivestreamResponse(deps, {
        id: livestream.id,
        user_id: livestream.user_id,
        title: livestream.title,
        description: livestream.description,
        playlist_url: livestream.playlist_url,
        thumbnail_url: livestream.thumbnail_url,
        start_at: livestream.start_at,
        end_at: livestream.end_at,
      })
      if (!livestreamResponse.ok) {
        await deps.connection.rollback()
        return c.text(livestreamResponse.error, 500)
      }
      livestreamResponses.push(livestreamResponse.data)
    }

    try {
      await deps.connection.commit()
    } catch {
      await deps.connection.rollback()
      return c.text('failed to commit', 500)
    }

    return c.json(livestreamResponses, 200)
  })

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

  handler.get(
    '/api/livestream/:livestream_id/report',
    verifyUserSessionMiddleware,
    async (c) => {
      const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
      const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
      if (Number.isNaN(livestreamId)) {
        return c.text('livestream_id in path must be integer', 400)
      }

      await deps.connection.beginTransaction()

      const livestream = await deps.connection
        .query<RowDataPacket[]>('SELECT * FROM livestreams WHERE id = ?', [
          livestreamId,
        ])
        .then(([[result]]) => ({ ok: true, data: result }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!livestream.ok) {
        await deps.connection.rollback()
        return c.text('failed to get livestream', 500)
      }
      if (livestream.data.user_id !== userId) {
        return c.text("can't get other streamer's livecomment reports", 403)
      }

      const livecommentReports = await deps.connection
        .query<RowDataPacket[]>(
          'SELECT * FROM livecomment_reports WHERE livestream_id = ?',
          [livestreamId],
        )
        .then(([results]) => ({ ok: true, data: results }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!livecommentReports.ok) {
        await deps.connection.rollback()
        return c.text('failed to get livecomment reports', 500)
      }

      const reportResponses: LivecommentReportResponse[] = []
      for (const livecommentReport of livecommentReports.data) {
        const report = await makeLivecommentReportResponse(deps, {
          id: livecommentReport.id,
          user_id: livecommentReport.user_id,
          livestream_id: livecommentReport.livestream_id,
          livecomment_id: livecommentReport.livecomment_id,
          created_at: livecommentReport.created_at,
        })
        if (!report.ok) {
          await deps.connection.rollback()
          return c.text(report.error, 500)
        }
        reportResponses.push(report.data)
      }

      try {
        await deps.connection.commit()
      } catch {
        await deps.connection.rollback()
        return c.text('failed to commit', 500)
      }

      return c.json(reportResponses, 200)
    },
  )

  handler.get(
    '/api/livestream/:livestream_id',
    verifyUserSessionMiddleware,
    async (c) => {
      const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
      if (Number.isNaN(livestreamId)) {
        return c.text('livestream_id in path must be integer', 400)
      }

      await deps.connection.beginTransaction()

      const livestream = await deps.connection
        .query<RowDataPacket[]>('SELECT * FROM livestreams WHERE id = ?', [
          livestreamId,
        ])
        .then(([[result]]) => ({ ok: true, data: result }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!livestream.ok) {
        await deps.connection.rollback()
        return c.text('failed to get livestream', 500)
      }
      if (!livestream.data) {
        return c.text('not found livestream that has the given id', 404)
      }

      const livestreamResponse = await makeLivestreamResponse(deps, {
        id: livestream.data.id,
        user_id: livestream.data.user_id,
        title: livestream.data.title,
        description: livestream.data.description,
        playlist_url: livestream.data.playlist_url,
        thumbnail_url: livestream.data.thumbnail_url,
        start_at: livestream.data.start_at,
        end_at: livestream.data.end_at,
      })
      if (!livestreamResponse.ok) {
        await deps.connection.rollback()
        return c.text(livestreamResponse.error, 500)
      }

      try {
        await deps.connection.commit()
      } catch {
        await deps.connection.rollback()
        return c.text('failed to commit', 500)
      }

      return c.json(livestreamResponse.data, 200)
    },
  )

  handler.post(
    '/api/livestream/:livestream_id/enter',
    verifyUserSessionMiddleware,
    async (c) => {
      const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
      const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
      if (Number.isNaN(livestreamId)) {
        return c.text('livestream_id in path must be integer', 400)
      }

      await deps.connection.beginTransaction()

      try {
        await deps.connection.query(
          'INSERT INTO livestream_viewers_history (user_id, livestream_id, created_at) VALUES(?, ?, ?)',
          [userId, livestreamId, Date.now()],
        )
      } catch {
        await deps.connection.rollback()
        return c.text('failed to insert livestream_view_history', 500)
      }

      try {
        await deps.connection.commit()
      } catch {
        await deps.connection.rollback()
        return c.text('failed to commit', 500)
      }

      // eslint-disable-next-line unicorn/no-null
      return c.body(null, 200)
    },
  )

  return handler
}
