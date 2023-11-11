import { Hono } from 'hono'
import { RowDataPacket } from 'mysql2/promise'
import { ApplicationDeps, HonoEnvironment } from '../types'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'

export const statsHandler = (deps: ApplicationDeps) => {
  const handler = new Hono<HonoEnvironment>()

  handler.get(
    '/api/user/:username/statistics',
    verifyUserSessionMiddleware,
    async (c) => {
      const username = c.req.param('username')

      await deps.connection.beginTransaction()

      const user = await deps.connection
        .query<RowDataPacket[]>('SELECT * FROM users WHERE name = ?', [
          username,
        ])
        .then(([[result]]) => ({ ok: true, data: result }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!user.ok) {
        await deps.connection.rollback()
        return c.json('failed to get user', 500)
      }
      if (!user.data) {
        await deps.connection.rollback()
        return c.json('not found user that has the given username', 404)
      }

      // ランク算出
      const users = await deps.connection
        .query<RowDataPacket[]>('SELECT * FROM users')
        .then(([results]) => ({ ok: true, data: results }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!users.ok) {
        await deps.connection.rollback()
        return c.json('failed to get users', 500)
      }

      const ranking: { username: string; score: number }[] = []
      for (const user of users.data) {
        const reaction = await deps.connection
          .query<RowDataPacket[]>(
            `
        SELECT COUNT(*) FROM users u
        INNER JOIN livestreams l ON l.user_id = u.id
        INNER JOIN reactions r ON r.livestream_id = l.id
        WHERE u.id = ?`,
            [user.id],
          )
          .then(
            ([[result]]) => ({ ok: true, data: result['COUNT(*)'] }) as const,
          )
          .catch((error) => ({ ok: false, error }) as const)
        if (!reaction.ok) {
          await deps.connection.rollback()
          return c.json('failed to count reactions', 500)
        }
        if (reaction.data === undefined) {
          await deps.connection.rollback()
          return c.json('failed to count reactions', 500)
        }

        const tips = await deps.connection
          .query<RowDataPacket[]>(
            `
        SELECT IFNULL(SUM(l2.tip), 0) FROM users u
        INNER JOIN livestreams l ON l.user_id = u.id	
        INNER JOIN livecomments l2 ON l2.livestream_id = l.id
        WHERE u.id = ?`,
            [user.id],
          )
          .then(
            ([[result]]) =>
              ({ ok: true, data: result['IFNULL(SUM(l2.tip), 0)'] }) as const,
          )
          .catch((error) => ({ ok: false, error }) as const)
        if (!tips.ok) {
          await deps.connection.rollback()
          return c.json('failed to count tips', 500)
        }
        if (tips.data === undefined) {
          await deps.connection.rollback()
          return c.json('failed to count tips', 500)
        }

        ranking.push({
          username: user.name,
          score: reaction.data + tips.data,
        })
      }

      ranking.sort((a, b) => b.score - a.score)

      let rank = 1
      for (const r of ranking) {
        if (r.username === username) break
        rank++
      }

      // リアクション数
      const totalReactions = await deps.connection
        .query<RowDataPacket[]>(
          `SELECT COUNT(*) FROM users u 
      INNER JOIN livestreams l ON l.user_id = u.id 
      INNER JOIN reactions r ON r.livestream_id = l.id
      WHERE u.name = ?
    `,
          [username],
        )
        .then(([[result]]) => ({ ok: true, data: result['COUNT(*)'] }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!totalReactions.ok) {
        await deps.connection.rollback()
        return c.json('failed to count reactions', 500)
      }
      if (totalReactions.data === undefined) {
        await deps.connection.rollback()
        return c.json('failed to count reactions', 500)
      }

      // ライブコメント数、チップ合計
      let totalLivecomments = 0
      let totalTip = 0
      for (const user of users.data) {
        const livestreams = await deps.connection
          .query<RowDataPacket[]>(
            `SELECT * FROM livestreams WHERE user_id = ?`,
            [user.id],
          )
          .then(([results]) => ({ ok: true, data: results }) as const)
          .catch((error) => ({ ok: false, error }) as const)
        if (!livestreams.ok) {
          await deps.connection.rollback()
          return c.json('failed to get livestreams', 500)
        }

        for (const livestream of livestreams.data) {
          const livecomments = await deps.connection
            .query<RowDataPacket[]>(
              `SELECT * FROM livecomments WHERE livestream_id = ?`,
              [livestream.id],
            )
            .then(([results]) => ({ ok: true, data: results }) as const)
            .catch((error) => ({ ok: false, error }) as const)
          if (!livecomments.ok) {
            await deps.connection.rollback()
            return c.json('failed to get livecomments', 500)
          }

          for (const livecomment of livecomments.data) {
            totalLivecomments++
            totalTip += livecomment.tip
          }
        }
      }

      // 合計視聴者数
      let viewersCount = 0
      for (const user of users.data) {
        const livestreams = await deps.connection
          .query<RowDataPacket[]>(
            `SELECT * FROM livestreams WHERE user_id = ?`,
            [user.id],
          )
          .then(([results]) => ({ ok: true, data: results }) as const)
          .catch((error) => ({ ok: false, error }) as const)
        if (!livestreams.ok) {
          await deps.connection.rollback()
          return c.json('failed to get livestreams', 500)
        }

        for (const livestream of livestreams.data) {
          const livestreamViewerCount = await deps.connection
            .query<RowDataPacket[]>(
              `SELECT COUNT(*) FROM livestream_viewers_history WHERE livestream_id = ?`,
              [livestream.id],
            )
            .then(
              ([[result]]) => ({ ok: true, data: result['COUNT(*)'] }) as const,
            )
            .catch((error) => ({ ok: false, error }) as const)
          if (!livestreamViewerCount.ok) {
            await deps.connection.rollback()
            return c.json('failed to get livecomments', 500)
          }

          viewersCount += livestreamViewerCount.data
        }
      }

      // お気に入り絵文字
      const favoriteEmoji = await deps.connection
        .query<RowDataPacket[]>(
          `
            SELECT r.emoji_name
            FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.name = ?
            GROUP BY emoji_name
            ORDER BY COUNT(*) DESC
            LIMIT 1
          `,
          [username],
        )
        .then(([[result]]) => ({ ok: true, data: result }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!favoriteEmoji.ok) {
        await deps.connection.rollback()
        return c.json('failed to get favorite emoji', 500)
      }

      return c.json({
        rank,
        viewers_count: viewersCount,
        total_reactions: totalReactions.data,
        total_livecomments: totalLivecomments,
        total_tip: totalTip,
        favorite_emoji: favoriteEmoji.data?.emoji_name,
      })
    },
  )

  handler.get(
    '/api/livestream/:livestream_id/statistics',
    verifyUserSessionMiddleware,
    async (c) => {
      const livestreamId = Number.parseInt(c.req.param('livestream_id'), 10)
      if (Number.isNaN(livestreamId)) {
        return c.json('livestream_id in path must be integer', 400)
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
        return c.json('failed to get livestream', 500)
      }
      if (!livestream.data) {
        await deps.connection.rollback()
        return c.json('cannot get stats of not found livestream', 404)
      }

      const livestreams = await deps.connection
        .query<RowDataPacket[]>('SELECT * FROM livestreams')
        .then(([results]) => ({ ok: true, data: results }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!livestreams.ok) {
        await deps.connection.rollback()
        return c.json('failed to get livestreams', 500)
      }

      // ランク算出
      const ranking: { livestreamId: number; title: string; score: number }[] =
        []
      for (const livestream of livestreams.data) {
        const reactionCount = await deps.connection
          .query<RowDataPacket[]>(
            'SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON l.id = r.livestream_id WHERE l.id = ?',
            [livestream.id],
          )
          .then(
            ([[result]]) => ({ ok: true, data: result['COUNT(*)'] }) as const,
          )
          .catch((error) => ({ ok: false, error }) as const)
        if (!reactionCount.ok) {
          await deps.connection.rollback()
          return c.json('failed to count reactions', 500)
        }
        if (reactionCount.data === undefined) {
          await deps.connection.rollback()
          return c.json('failed to count reactions', 500)
        }

        const totalTip = await deps.connection
          .query<RowDataPacket[]>(
            'SELECT IFNULL(SUM(l2.tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l.id = l2.livestream_id WHERE l.id = ?',
            [livestream.id],
          )
          .then(
            ([[result]]) =>
              ({ ok: true, data: result['IFNULL(SUM(l2.tip), 0)'] }) as const,
          )
          .catch((error) => ({ ok: false, error }) as const)
        if (!totalTip.ok) {
          await deps.connection.rollback()
          return c.json('failed to count tips', 500)
        }
        if (totalTip.data === undefined) {
          await deps.connection.rollback()
          return c.json('failed to count tips', 500)
        }

        ranking.push({
          livestreamId: livestream.id,
          title: livestream.title,
          score: reactionCount.data + totalTip.data,
        })
      }
      ranking.sort((a, b) => b.score - a.score)

      let rank = 1
      for (const r of ranking) {
        if (r.livestreamId === livestreamId) break
        rank++
      }

      // 視聴者数算出
      const viewersCount = await deps.connection
        .query<RowDataPacket[]>(
          'SELECT COUNT(*) FROM livestreams l INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id WHERE l.id = ?',
          [livestreamId],
        )
        .then(([[result]]) => ({ ok: true, data: result['COUNT(*)'] }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!viewersCount.ok) {
        await deps.connection.rollback()
        return c.json('failed to count viewers', 500)
      }

      // 最大チップ額
      const maxTip = await deps.connection
        .query<RowDataPacket[]>(
          'SELECT IFNULL(MAX(tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l2.livestream_id = l.id WHERE l.id = ?',
          [livestreamId],
        )
        .then(
          ([[result]]) =>
            ({ ok: true, data: result['IFNULL(MAX(tip), 0)'] }) as const,
        )
        .catch((error) => ({ ok: false, error }) as const)
      if (!maxTip.ok) {
        await deps.connection.rollback()
        return c.json('failed to get max tip', 500)
      }

      // リアクション数
      const totalReactions = await deps.connection
        .query<RowDataPacket[]>(
          'SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON r.livestream_id = l.id WHERE l.id = ?',
          [livestreamId],
        )
        .then(([[result]]) => ({ ok: true, data: result['COUNT(*)'] }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!totalReactions.ok) {
        await deps.connection.rollback()
        return c.json('failed to count reactions', 500)
      }

      // スパム報告数
      const totalReports = await deps.connection
        .query<RowDataPacket[]>(
          'SELECT COUNT(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = ?',
          [livestreamId],
        )
        .then(([[result]]) => ({ ok: true, data: result['COUNT(*)'] }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!totalReports.ok) {
        await deps.connection.rollback()
        return c.json('failed to count reports', 500)
      }

      return c.json({
        rank,
        viewers_count: viewersCount.data,
        total_reactions: totalReactions.data,
        total_reports: totalReports.data,
        max_tip: maxTip.data,
      })
    },
  )

  return handler
}
