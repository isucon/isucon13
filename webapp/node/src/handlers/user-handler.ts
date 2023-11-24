import { Context } from 'hono'
import { ResultSetHeader, RowDataPacket } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
import {
  defaultUserIDKey,
  defaultUserNameKey,
  defaultSessionExpiresKey,
} from '../contants'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { fillUserResponse } from '../utils/fill-user-response'
import { throwErrorWith } from '../utils/throw-error-with'
import { IconModel, UserModel } from '../types/models'

// GET /api/user/:username/icon
export const getIconHandler = [
  async (c: Context<HonoEnvironment, '/api/user/:username/icon'>) => {
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
        await conn.rollback()
        return c.text('not found user that has the given username', 404)
      }

      const [[icon]] = await conn
        .query<(Pick<IconModel, 'image'> & RowDataPacket)[]>(
          'SELECT image FROM icons WHERE user_id = ?',
          [user.id],
        )
        .catch(throwErrorWith('failed to get icon'))
      if (!icon) {
        await conn.rollback()
        return c.body(await c.get('runtime').fallbackUserIcon(), 200, {
          'Content-Type': 'image/jpeg',
        })
      }

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.body(icon.image, 200, {
        'Content-Type': 'image/jpeg',
      })
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// POST /api/icon
export const postIconHandler = [
  verifyUserSessionMiddleware,
  async (c: Context<HonoEnvironment, '/api/icon'>) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware

    // base64 encoded image
    const body = await c.req.json<{ image: string }>()

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      await conn
        .execute('DELETE FROM icons WHERE user_id = ?', [userId])
        .catch(throwErrorWith('failed to delete old user icon'))

      const [{ insertId: iconId }] = await conn
        .query<ResultSetHeader>(
          'INSERT INTO icons (user_id, image) VALUES (?, ?)',
          [userId, Buffer.from(body.image, 'base64')],
        )
        .catch(throwErrorWith('failed to insert icon'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json({ id: iconId }, 201)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// GET /api/user/me
export const getMeHandler = [
  verifyUserSessionMiddleware,
  async (c: Context<HonoEnvironment, '/api/user/me'>) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware

    const conn = await c.get('pool').getConnection()
    await conn.beginTransaction()

    try {
      const [[user]] = await conn
        .query<(UserModel & RowDataPacket)[]>(
          'SELECT * FROM users WHERE id = ?',
          [userId],
        )
        .catch(throwErrorWith('failed to get user'))

      if (!user) {
        await conn.rollback()
        return c.text('not found user that has the userid in session', 404)
      }

      const response = await fillUserResponse(
        conn,
        user,
        c.get('runtime').fallbackUserIcon,
      ).catch(throwErrorWith('failed to fill user'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(response)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]

// ユーザ登録API
// POST /api/register
export const registerHandler = async (
  c: Context<HonoEnvironment, '/api/register'>,
) => {
  const body = await c.req.json<{
    name: string
    display_name: string
    description: string
    password: string
    theme: { dark_mode: boolean }
  }>()

  if (body.name === 'pipe') {
    return c.text("the username 'pipe' is reserved", 400)
  }

  const hashedPassword = await c
    .get('runtime')
    .hashPassword(body.password)
    .catch(throwErrorWith('failed to generate hashed password'))

  const conn = await c.get('pool').getConnection()
  await conn.beginTransaction()

  try {
    const [{ insertId: userId }] = await conn
      .execute<ResultSetHeader>(
        'INSERT INTO users (name, display_name, description, password) VALUES(?, ?, ?, ?)',
        [body.name, body.display_name, body.description, hashedPassword],
      )
      .catch(throwErrorWith('failed to insert user'))

    await conn
      .execute('INSERT INTO themes (user_id, dark_mode) VALUES(?, ?)', [
        userId,
        body.theme.dark_mode,
      ])
      .catch(throwErrorWith('failed to insert user theme'))

    await c
      .get('runtime')
      .exec([
        'pdnsutil',
        'add-record',
        'u.isucon.dev',
        body.name,
        'A',
        '0',
        c.get('runtime').powerDNSSubdomainAddress,
      ])
      .catch(throwErrorWith('failed to add record to powerdns'))

    const response = await fillUserResponse(
      conn,
      {
        id: userId,
        name: body.name,
        display_name: body.display_name,
        description: body.description,
      },
      c.get('runtime').fallbackUserIcon,
    ).catch(throwErrorWith('failed to fill user'))

    await conn.commit().catch(throwErrorWith('failed to commit'))

    return c.json(response, 201)
  } catch (error) {
    await conn.rollback()
    return c.text(`Internal Server Error\n${error}`, 500)
  } finally {
    await conn.rollback()
    conn.release()
  }
}

// ユーザログインAPI
// POST /api/login
export const loginHandler = async (
  c: Context<HonoEnvironment, '/api/login'>,
) => {
  const body = await c.req.json<{
    username: string
    password: string
  }>()

  const conn = await c.get('pool').getConnection()
  await conn.beginTransaction()

  try {
    // usernameはUNIQUEなので、whereで一意に特定できる
    const [[user]] = await conn
      .query<(UserModel & RowDataPacket)[]>(
        'SELECT * FROM users WHERE name = ?',
        [body.username],
      )
      .catch(throwErrorWith('failed to get user'))

    if (!user) {
      await conn.rollback()
      return c.text('invalid username or password', 401)
    }

    await conn.commit().catch(throwErrorWith('failed to commit'))

    const isPasswordMatch = await c
      .get('runtime')
      .comparePassword(body.password, user.password)
      .catch(throwErrorWith('failed to compare hash and password'))
    if (!isPasswordMatch) {
      return c.text('invalid username or password', 401)
    }

    // 1時間でセッションが切れるようにする
    const sessionEndAt = Date.now() + 1000 * 60 * 60

    const session = c.get('session')
    session.set(defaultUserIDKey, user.id)
    session.set(defaultUserNameKey, user.name)
    session.set(defaultSessionExpiresKey, sessionEndAt)

    // eslint-disable-next-line unicorn/no-null
    return c.body(null)
  } catch (error) {
    await conn.rollback()
    return c.text(`Internal Server Error\n${error}`, 500)
  } finally {
    await conn.rollback()
    conn.release()
  }
}

// GET /api/user/:username
export const getUserHandler = [
  verifyUserSessionMiddleware,
  async (c: Context<HonoEnvironment, '/api/user/:username'>) => {
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
        await conn.rollback()
        return c.text('not found user that has the given username', 404)
      }

      const response = await fillUserResponse(
        conn,
        user,
        c.get('runtime').fallbackUserIcon,
      ).catch(throwErrorWith('failed to fill user'))

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(response)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      await conn.rollback()
      conn.release()
    }
  },
]
