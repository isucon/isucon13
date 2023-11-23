import { Context } from 'hono'
import { RowDataPacket } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { TagsModel, ThemeModel, UserModel } from '../types/models'
import { throwErrorWith } from '../utils/throw-error-with'

// GET /api/tag
export const getTagHandler = async (
  c: Context<HonoEnvironment, '/api/tag'>,
) => {
  const conn = await c.get('pool').getConnection()
  await conn.beginTransaction()

  try {
    const [tags] = await conn
      .execute<(TagsModel & RowDataPacket)[]>('SELECT * FROM tags')
      .catch(throwErrorWith('failed to get tags'))

    await conn.commit().catch(throwErrorWith('failed to commit'))

    const tagResponses = []
    for (const tag of tags) {
      tagResponses.push({
        id: tag.id,
        name: tag.name,
      })
    }

    return c.json({ tags: tagResponses })
  } catch (error) {
    await conn.rollback()
    return c.text(`Internal Server Error\n${error}`, 500)
  } finally {
    await conn.rollback()
    conn.release()
  }
}

// 配信者のテーマ取得API
// GET /api/user/:username/theme
export const getStreamerThemeHandler = [
  verifyUserSessionMiddleware,
  async (c: Context<HonoEnvironment, '/api/user/:username/theme'>) => {
    const username = c.req.param('username')

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      const [[user]] = await conn
        .execute<(Pick<UserModel, 'id'> & RowDataPacket)[]>(
          'SELECT id FROM users WHERE name = ?',
          [username],
        )
        .catch(throwErrorWith('failed to get user'))

      if (!user) {
        await conn.rollback()
        return c.text('not found user that has the given username', 404)
      }

      const [[theme]] = await conn
        .execute<(ThemeModel & RowDataPacket)[]>(
          'SELECT * FROM themes WHERE user_id = ?',
          [user.id],
        )
        .catch(throwErrorWith('failed to get user theme'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      const themeResponse = {
        id: theme.id,
        dark_mode: !!theme.dark_mode,
      }

      return c.json(themeResponse)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]
