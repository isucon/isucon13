import { Hono } from 'hono'
import { RowDataPacket, ResultSetHeader } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
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
import {
  LivecommentReportsModel,
  LivestreamTagsModel,
  LivestreamsModel,
  ReservationSlotsModel,
  TagsModel,
  UserModel,
} from '../types/models'
import { throwErrorWith } from '../utils/throw-error-with'

export const livestreamHandler = new Hono<HonoEnvironment>()

livestreamHandler.get('/api/livestream/search', async (c) => {
  const tagName = c.req.query('tag')

  const conn = await c.get('pool').getConnection()
  await conn.beginTransaction()

  try {
    const livestreams: (LivestreamsModel & RowDataPacket)[] = []

    if (tagName) {
      // タグによる取得
      const [tagIds] = await conn
        .query<(Pick<TagsModel, 'id'> & RowDataPacket)[]>(
          'SELECT id FROM tags WHERE name = ?',
          [tagName],
        )
        .catch(throwErrorWith('failed to get tag'))

      const [livestreamTags] = await conn
        .query<(LivestreamTagsModel & RowDataPacket)[]>(
          'SELECT * FROM livestream_tags WHERE tag_id IN (?)',
          [tagIds.map((tag) => tag.id)],
        )
        .catch(throwErrorWith('failed to get keyTaggedLivestreams'))

      for (const livestreamTag of livestreamTags) {
        const [[livestream]] = await conn
          .query<(LivestreamsModel & RowDataPacket)[]>(
            'SELECT * FROM livestreams WHERE id = ?',
            [livestreamTag.livestream_id],
          )
          .catch(throwErrorWith('failed to get livestream'))

        livestreams.push(livestream)
      }
    } else {
      // 検索条件なし
      let query = `SELECT * FROM livestreams`
      const limit = c.req.query('limit')
      if (limit) {
        const limitNumber = Number.parseInt(limit, 10)
        if (Number.isNaN(limitNumber)) {
          return c.text('limit query parameter must be integer', 400)
        }
        query += ` LIMIT ${limitNumber}`
      }

      const [results] = await conn
        .query<(LivestreamsModel & RowDataPacket)[]>(query)
        .catch(throwErrorWith('failed to get livestreams'))

      livestreams.push(...results)
    }

    const livestreamResponses: LivestreamResponse[] = []
    for (const livestream of livestreams) {
      const livestreamResponse = await makeLivestreamResponse(conn, {
        id: livestream.id,
        user_id: livestream.user_id,
        title: livestream.title,
        description: livestream.description,
        playlist_url: livestream.playlist_url,
        thumbnail_url: livestream.thumbnail_url,
        start_at: livestream.start_at,
        end_at: livestream.end_at,
      })
      livestreamResponses.push(livestreamResponse)
    }

    await conn.commit().catch(throwErrorWith('failed to commit'))

    return c.json(livestreamResponses)
  } catch (error) {
    await conn.rollback()
    return c.text(`Internal Server Error\n${error}`, 500)
  } finally {
    conn.release()
  }
})

livestreamHandler.post(
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

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      // 2024/04/01からの１年間の期間内であるかチェック
      const termStartAt = Date.UTC(2024, 3, 1) // NOTE: month is 0-indexed
      const termEndAt = Date.UTC(2025, 3, 1) // NOTE: month is 0-indexed
      const reserveStartAt = body.start_at // NOTE: body.start_at is unixtime, so it is in seconds
      const reserveEndAt = body.end_at // NOTE: body.end_at is unixtime, so it is in seconds
      if (
        reserveStartAt * 1000 >= termEndAt ||
        reserveEndAt * 1000 <= termStartAt
      ) {
        await conn.rollback()
        console.log({
          termStartAt,
          termEndAt,
          reserveStartAt: reserveStartAt * 1000,
          reserveEndAt: reserveEndAt * 1000,
        })
        return c.text('bad reservation time range', 400)
      }

      // 予約枠をみて、予約が可能か調べる
      const [slots] = await conn
        .query<(ReservationSlotsModel & RowDataPacket)[]>(
          'SELECT * FROM reservation_slots WHERE start_at >= ? AND end_at <= ?',
          [body.start_at, body.end_at],
        )
        .catch(throwErrorWith('failed to get reservation_slots'))

      for (const slot of slots) {
        const [[count]] = await conn
          .query<(Pick<ReservationSlotsModel, 'slot'> & RowDataPacket)[]>(
            'SELECT slot FROM reservation_slots WHERE start_at = ? AND end_at = ?',
            [slot.start_at, slot.end_at],
          )
          .catch(throwErrorWith('failed to get reservation_slots'))

        console.log(
          `${new Date(slot.start_at * 1000).toISOString()} ~ ${new Date(
            slot.end_at * 1000,
          ).toISOString()} 予約枠の残数 - ${count.slot}`,
        )
        if (count.slot < 1) {
          return c.text(
            `予約区間 ${new Date(slot.start_at).toISOString()} ~ ${new Date(
              slot.end_at,
            ).toISOString()} が予約できません`,
            400,
          )
        }
      }

      await conn
        .query(
          'UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= ? AND end_at <= ?',
          [reserveStartAt, reserveEndAt],
        )
        .catch(throwErrorWith('failed to update reservation_slots'))
      const [{ insertId }] = await conn
        .query<ResultSetHeader>(
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
        .catch(throwErrorWith('failed to insert livestream'))

      // タグ追加
      for (const tagId of body.tags) {
        await conn
          .execute(
            'INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (?, ?)',
            [insertId, tagId],
          )
          .catch(throwErrorWith('failed to insert livestream tag'))
      }

      const response = await makeLivestreamResponse(conn, {
        id: insertId,
        user_id: userId,
        title: body.title,
        description: body.description,
        playlist_url: body.playlist_url,
        thumbnail_url: body.thumbnail_url,
        start_at: body.start_at,
        end_at: body.end_at,
      })

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(response, 201)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

livestreamHandler.get(
  '/api/livestream/:livestream_id/report',
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
      const [[livestream]] = await conn
        .query<(LivestreamsModel & RowDataPacket)[]>(
          'SELECT * FROM livestreams WHERE id = ?',
          [livestreamId],
        )
        .catch(throwErrorWith('failed to get livestream'))

      if (livestream.user_id !== userId) {
        return c.text("can't get other streamer's livecomment reports", 403)
      }

      const [livecommentReports] = await conn
        .query<(LivecommentReportsModel & RowDataPacket)[]>(
          'SELECT * FROM livecomment_reports WHERE livestream_id = ?',
          [livestreamId],
        )
        .catch(throwErrorWith('failed to get livecomment reports'))

      const reportResponses: LivecommentReportResponse[] = []
      for (const livecommentReport of livecommentReports) {
        const report = await makeLivecommentReportResponse(conn, {
          id: livecommentReport.id,
          user_id: livecommentReport.user_id,
          livestream_id: livecommentReport.livestream_id,
          livecomment_id: livecommentReport.livecomment_id,
          created_at: livecommentReport.created_at,
        })
        reportResponses.push(report)
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(reportResponses)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

livestreamHandler.get(
  '/api/livestream/:livestream_id',
  verifyUserSessionMiddleware,
  async (c) => {
    const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
    if (Number.isNaN(livestreamId)) {
      return c.text('livestream_id in path must be integer', 400)
    }

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      const [[livestream]] = await conn
        .query<(LivestreamsModel & RowDataPacket)[]>(
          'SELECT * FROM livestreams WHERE id = ?',
          [livestreamId],
        )
        .catch(throwErrorWith('failed to get livestream'))

      if (!livestream) {
        return c.text('not found livestream that has the given id', 404)
      }

      const livestreamResponse = await makeLivestreamResponse(conn, livestream)

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(livestreamResponse)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

livestreamHandler.post(
  '/api/livestream/:livestream_id/enter',
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
      await conn
        .query(
          'INSERT INTO livestream_viewers_history (user_id, livestream_id, created_at) VALUES(?, ?, ?)',
          [userId, livestreamId, Date.now()],
        )
        .catch(throwErrorWith('failed to insert livestream_viewers_history'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      // eslint-disable-next-line unicorn/no-null
      return c.body(null)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

livestreamHandler.delete(
  '/api/livestream/:livestream_id/exit',
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
      await conn
        .query(
          'DELETE FROM livestream_viewers_history WHERE user_id = ? AND livestream_id = ?',
          [userId, livestreamId],
        )
        .catch(throwErrorWith('failed to delete livestream_viewers_history'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      // eslint-disable-next-line unicorn/no-null
      return c.body(null)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

livestreamHandler.get(
  '/api/livestream',
  verifyUserSessionMiddleware,
  async (c) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      const [livestreams] = await conn
        .query<(LivestreamsModel & RowDataPacket)[]>(
          'SELECT * FROM livestreams WHERE user_id = ?',
          [userId],
        )
        .catch(throwErrorWith('failed to get livestreams'))

      const livestreamResponses: LivestreamResponse[] = []
      for (const livestream of livestreams) {
        const livestreamResponse = await makeLivestreamResponse(
          conn,
          livestream,
        )

        livestreamResponses.push(livestreamResponse)
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(livestreamResponses)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

livestreamHandler.get(
  '/api/user/:username/livestream',
  verifyUserSessionMiddleware,
  async (c) => {
    const username = c.req.param('username')

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      const [[user]] = await conn
        .query<(UserModel & RowDataPacket)[]>(
          'SELECT * FROM users WHERE name = ?',
          [username],
        )
        .catch(throwErrorWith('failed to get user'))

      if (!user) {
        return c.text('not found user that has the given username', 404)
      }

      const [livestreams] = await conn
        .query<(LivestreamsModel & RowDataPacket)[]>(
          'SELECT * FROM livestreams WHERE user_id = ?',
          [user.id],
        )
        .catch(throwErrorWith('failed to get livestreams'))

      const livestreamResponses: LivestreamResponse[] = []
      for (const livestream of livestreams) {
        const livestreamResponse = await makeLivestreamResponse(
          conn,
          livestream,
        )

        livestreamResponses.push(livestreamResponse)
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(livestreamResponses)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)
