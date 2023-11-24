import { Context } from 'hono'
import { RowDataPacket, ResultSetHeader } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { defaultUserIDKey } from '../contants'
import {
  LivestreamResponse,
  fillLivestreamResponse,
} from '../utils/fill-livestream-response'
import {
  LivecommentReportResponse,
  fillLivecommentReportResponse,
} from '../utils/fill-livecomment-report-response'
import {
  LivecommentReportsModel,
  LivestreamTagsModel,
  LivestreamsModel,
  ReservationSlotsModel,
  TagsModel,
  UserModel,
} from '../types/models'
import { throwErrorWith } from '../utils/throw-error-with'
import { atoi } from '../utils/integer'

// POST /api/livestream/reservation
export const reserveLivestreamHandler = [
  verifyUserSessionMiddleware,
  async (c: Context<HonoEnvironment, '/api/livestream/reservation'>) => {
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
      // 2023/11/25 10:00からの１年間の期間内であるかチェック
      const termStartAt = Date.UTC(2023, 10, 25, 1)
      const termEndAt = Date.UTC(2024, 10, 25, 1)
      const reserveStartAt = body.start_at * 1000
      const reserveEndAt = body.end_at * 1000

      if (reserveStartAt >= termEndAt || reserveEndAt <= termStartAt) {
        await conn.rollback()
        return c.text('bad reservation time range', 400)
      }

      // 予約枠をみて、予約が可能か調べる
      // NOTE: 並列な予約のoverbooking防止にFOR UPDATEが必要
      const [slots] = await conn
        .query<(ReservationSlotsModel & RowDataPacket)[]>(
          'SELECT * FROM reservation_slots WHERE start_at >= ? AND end_at <= ? FOR UPDATE',
          [body.start_at, body.end_at],
        )
        .catch((error) => {
          console.warn(`予約枠一覧取得でエラー発生: ${error}`)
          return throwErrorWith('failed to get reservation_slots')(error)
        })

      for (const slot of slots) {
        const [[count]] = await conn
          .query<(Pick<ReservationSlotsModel, 'slot'> & RowDataPacket)[]>(
            'SELECT slot FROM reservation_slots WHERE start_at = ? AND end_at = ?',
            [slot.start_at, slot.end_at],
          )
          .catch(throwErrorWith('failed to get reservation_slots'))

        console.info(
          `${slot.start_at} ~ ${slot.end_at} 予約枠の残数 = ${count.slot}`,
        )
        if (count.slot < 1) {
          return c.text(
            `予約期間 ${Math.floor(termStartAt / 1000)} ~ ${Math.floor(
              termEndAt / 1000,
            )}に対して、予約区間 ${body.start_at} ~ ${
              body.end_at
            }が予約できません`,
            400,
          )
        }
      }

      await conn
        .query(
          'UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= ? AND end_at <= ?',
          [body.start_at, body.end_at],
        )
        .catch(throwErrorWith('failed to update reservation_slot'))
      const [{ insertId: livestreamId }] = await conn
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
            [livestreamId, tagId],
          )
          .catch(throwErrorWith('failed to insert livestream tag'))
      }

      const response = await fillLivestreamResponse(
        conn,
        {
          id: livestreamId,
          user_id: userId,
          title: body.title,
          description: body.description,
          playlist_url: body.playlist_url,
          thumbnail_url: body.thumbnail_url,
          start_at: body.start_at,
          end_at: body.end_at,
        },
        c.get('runtime').fallbackUserIcon,
      ).catch(throwErrorWith('failed to fill livestream'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(response, 201)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// GET /api/livestream/search
export const searchLivestreamsHandler = async (
  c: Context<HonoEnvironment, '/api/livestream/search'>,
) => {
  const keyTagName = c.req.query('tag')

  const conn = await c.get('pool').getConnection()
  await conn.beginTransaction()

  try {
    const livestreams: (LivestreamsModel & RowDataPacket)[] = []

    if (keyTagName) {
      // タグによる取得
      const [tagIds] = await conn
        .query<(Pick<TagsModel, 'id'> & RowDataPacket)[]>(
          'SELECT id FROM tags WHERE name = ?',
          [keyTagName],
        )
        .catch(throwErrorWith('failed to get tag'))

      const [livestreamTags] = await conn
        .query<(LivestreamTagsModel & RowDataPacket)[]>(
          'SELECT * FROM livestream_tags WHERE tag_id IN (?) ORDER BY livestream_id DESC',
          [tagIds.map((tag) => tag.id)],
        )
        .catch(throwErrorWith('failed to get keyTaggedLivestreams'))

      for (const livestreamTag of livestreamTags) {
        const [[livestream]] = await conn
          .query<(LivestreamsModel & RowDataPacket)[]>(
            'SELECT * FROM livestreams WHERE id = ?',
            [livestreamTag.livestream_id],
          )
          .catch(throwErrorWith('failed to get livestreams'))

        livestreams.push(livestream)
      }
    } else {
      // 検索条件なし
      let query = `SELECT * FROM livestreams ORDER BY id DESC`
      const limit = c.req.query('limit')
      if (limit) {
        const limitNumber = atoi(limit)
        if (limitNumber === false) {
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
      const livestreamResponse = await fillLivestreamResponse(
        conn,
        livestream,
        c.get('runtime').fallbackUserIcon,
      ).catch(throwErrorWith('failed to fill livestream'))
      livestreamResponses.push(livestreamResponse)
    }

    await conn.commit().catch(throwErrorWith('failed to commit'))

    return c.json(livestreamResponses)
  } catch (error) {
    await conn.rollback()
    return c.text(`Internal Server Error\n${error}`, 500)
  } finally {
    await conn.rollback()
    conn.release()
  }
}

// GET /api/livestream
export const getMyLivestreamsHandler = [
  verifyUserSessionMiddleware,
  async (c: Context<HonoEnvironment, '/api/livestream'>) => {
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
        const livestreamResponse = await fillLivestreamResponse(
          conn,
          livestream,
          c.get('runtime').fallbackUserIcon,
        ).catch(throwErrorWith('failed to fill livestream'))

        livestreamResponses.push(livestreamResponse)
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(livestreamResponses)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// GET /api/user/:username/livestream
export const getUserLivestreamsHandler = [
  verifyUserSessionMiddleware,
  async (c: Context<HonoEnvironment, '/api/user/:username/livestream'>) => {
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
        return c.text('user not found', 404)
      }

      const [livestreams] = await conn
        .query<(LivestreamsModel & RowDataPacket)[]>(
          'SELECT * FROM livestreams WHERE user_id = ?',
          [user.id],
        )
        .catch(throwErrorWith('failed to get livestreams'))

      const livestreamResponses: LivestreamResponse[] = []
      for (const livestream of livestreams) {
        const livestreamResponse = await fillLivestreamResponse(
          conn,
          livestream,
          c.get('runtime').fallbackUserIcon,
        ).catch(throwErrorWith('failed to fill livestream'))

        livestreamResponses.push(livestreamResponse)
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(livestreamResponses)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// POST /api/livestream/:livestream_id/enter
export const enterLivestreamHandler = [
  verifyUserSessionMiddleware,
  async (
    c: Context<HonoEnvironment, '/api/livestream/:livestream_id/enter'>,
  ) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = atoi(c.req.param('livestream_id'))
    if (livestreamId === false) {
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
        .catch(throwErrorWith('failed to insert livestream_view_history'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      // eslint-disable-next-line unicorn/no-null
      return c.body(null)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// DELETE /api/livestream/:livestream_id/exit
export const exitLivestreamHandler = [
  verifyUserSessionMiddleware,
  async (
    c: Context<HonoEnvironment, '/api/livestream/:livestream_id/exit'>,
  ) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = atoi(c.req.param('livestream_id'))
    if (livestreamId === false) {
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
        .catch(throwErrorWith('failed to delete livestream_view_history'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      // eslint-disable-next-line unicorn/no-null
      return c.body(null)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// GET /api/livestream/:livestream_id
export const getLivestreamHandler = [
  verifyUserSessionMiddleware,
  async (c: Context<HonoEnvironment, '/api/livestream/:livestream_id'>) => {
    const livestreamId = atoi(c.req.param('livestream_id'))
    if (livestreamId === false) {
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

      const livestreamResponse = await fillLivestreamResponse(
        conn,
        livestream,
        c.get('runtime').fallbackUserIcon,
      ).catch(throwErrorWith('failed to fill livestream'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(livestreamResponse)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// GET /api/livestream/:livestream_id/report
export const getLivecommentReportsHandler = [
  verifyUserSessionMiddleware,
  async (
    c: Context<HonoEnvironment, '/api/livestream/:livestream_id/report'>,
  ) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware
    const livestreamId = atoi(c.req.param('livestream_id'))
    if (livestreamId === false) {
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
        const report = await fillLivecommentReportResponse(
          conn,
          livecommentReport,
          c.get('runtime').fallbackUserIcon,
        ).catch(throwErrorWith('failed to fill livecomment report'))
        reportResponses.push(report)
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(reportResponses)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]
