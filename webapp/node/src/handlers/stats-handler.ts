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
  return handler
}
