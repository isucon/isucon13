import { Hono } from 'hono'
import { RowDataPacket } from 'mysql2/promise'
import { ApplicationDeps, HonoEnvironment } from '../types'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'

export const topHandler = (deps: ApplicationDeps) => {
  const handler = new Hono<HonoEnvironment>()

  handler.get(
    '/api/user/:username/theme',
    verifyUserSessionMiddleware,
    async (c) => {
      const username = c.req.param('username')

      await deps.connection.beginTransaction()

      const result = await deps.connection
        .execute<RowDataPacket[]>('SELECT id FROM users WHERE name = ?', [
          username,
        ])
        .then(([[data]]) => ({ ok: true, data }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!result.ok) {
        await deps.connection.rollback()
        return c.text('failed to get user', 500)
      }
      if (!result.data) {
        await deps.connection.rollback()
        return c.text('not found user that has the given username', 404)
      }

      const theme = await deps.connection
        .execute<RowDataPacket[]>('SELECT * FROM themes WHERE user_id = ?', [
          result.data.id,
        ])
        .then(([[data]]) => ({ ok: true, data }) as const)
        .catch((error) => ({ ok: false, error }) as const)
      if (!theme.ok) {
        await deps.connection.rollback()
        return c.text('failed to get user theme', 500)
      }

      try {
        await deps.connection.commit()
      } catch {
        await deps.connection.rollback()
        return c.text('failed to commit', 500)
      }

      return c.json(
        {
          id: theme.data.id,
          dark_mode: !!theme.data.dark_mode,
        },
        200,
      )
    },
  )

  return handler
}
