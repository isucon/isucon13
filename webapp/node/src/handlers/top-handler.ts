import { Hono } from 'hono'
import { RowDataPacket } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { TagsModel, ThemeModel, UserModel } from '../types/models'
import { throwErrorWith } from '../utils/throw-error-with'

export const topHandler = new Hono<HonoEnvironment>()

topHandler.get(
  '/api/user/:username/theme',
  verifyUserSessionMiddleware,
  async (c) => {
    const username = c.req.param('username')

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      const [[result]] = await conn
        .execute<(Pick<UserModel, 'id'> & RowDataPacket)[]>(
          'SELECT id FROM users WHERE name = ?',
          [username],
        )
        .catch(throwErrorWith('failed to get user'))

      if (!result) {
        await conn.rollback()
        return c.text('not found user that has the given username', 404)
      }

      const [[theme]] = await conn
        .execute<(ThemeModel & RowDataPacket)[]>(
          'SELECT * FROM themes WHERE user_id = ?',
          [result.id],
        )
        .catch(throwErrorWith('failed to get user theme'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json({ id: theme.id, dark_mode: !!theme.dark_mode })
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

topHandler.get('/api/tag', async (c) => {
  const conn = await c.get('pool').getConnection()
  await conn.beginTransaction()

  try {
    const [tags] = await conn
      .execute<(TagsModel & RowDataPacket)[]>('SELECT * FROM tags')
      .catch(throwErrorWith('failed to get tags'))

    await conn.commit().catch(throwErrorWith('failed to commit'))

    return c.json({
      tags: tags.map((tag) => ({
        id: tag.id,
        name: tag.name,
      })),
    })
  } catch (error) {
    await conn.rollback()
    return c.text(`Internal Server Error\n${error}`, 500)
  } finally {
    conn.release()
  }
})
