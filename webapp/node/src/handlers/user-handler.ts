import { Hono } from 'hono'
import { ResultSetHeader, RowDataPacket } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
import {
  defaultUserIDKey,
  defaultUserNameKey,
  defaultSessionExpiresKey,
} from '../contants'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'
import { makeUserResponse } from '../utils/make-user-response'
import { throwErrorWith } from '../utils/throw-error-with'
import { IconModel, UserModel } from '../types/models'

export const userHandler = new Hono<HonoEnvironment>()

userHandler.post('/api/register', async (c) => {
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

  const hashedPassword = await c.get('deps').hashPassword(body.password)

  const conn = await c.get('pool').getConnection()
  await conn.beginTransaction()

  try {
    const [{ insertId }] = await conn
      .execute<ResultSetHeader>(
        'INSERT INTO users (name, display_name, description, password) VALUES(?, ?, ?, ?)',
        [body.name, body.display_name, body.description, hashedPassword],
      )
      .catch(throwErrorWith('failed to insert user'))

    await conn
      .execute('INSERT INTO themes (user_id, dark_mode) VALUES(?, ?)', [
        insertId,
        body.theme.dark_mode,
      ])
      .catch(throwErrorWith('failed to insert user theme'))

    await c
      .get('deps')
      .exec([
        'pdnsutil',
        'add-record',
        'u.isucon.dev',
        body.name,
        'A',
        '30',
        c.get('deps').powerDNSSubdomainAddress,
      ])
      .catch(throwErrorWith('failed to add record to powerdns'))

    const response = await makeUserResponse(conn, {
      id: insertId,
      name: body.name,
      display_name: body.display_name,
      description: body.description,
    })

    await conn.commit().catch(throwErrorWith('failed to commit'))

    return c.json(response, 201)
  } catch (error) {
    await conn.rollback()
    return c.text(`Internal Server Error\n${error}`, 500)
  } finally {
    conn.release()
  }
})

userHandler.post('/api/login', async (c) => {
  const body = await c.req.json<{
    username: string
    password: string
  }>()

  const conn = await c.get('pool').getConnection()
  await conn.beginTransaction()

  try {
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
      .get('deps')
      .comparePassword(body.password, user.password)
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
    conn.release()
  }
})

userHandler.get('/api/user/me', verifyUserSessionMiddleware, async (c) => {
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

    const response = await makeUserResponse(conn, user)

    await conn.commit().catch(throwErrorWith('failed to commit'))

    return c.json(response)
  } catch (error) {
    await conn.rollback()
    return c.text(`Internal Server Error\n${error}`, 500)
  } finally {
    conn.release()
  }
})

userHandler.get(
  '/api/user/:username',
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
        await conn.rollback()
        return c.text('not found user that has the given username', 404)
      }

      const response = await makeUserResponse(conn, user)

      await conn.commit().catch(throwErrorWith('failed to commit'))

      return c.json(response)
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)

userHandler.post('/api/icon', verifyUserSessionMiddleware, async (c) => {
  const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware

  // base64 encoded image
  const body = await c.req.json<{ image: string }>()

  const conn = await c.get('pool').getConnection()
  await conn.beginTransaction()

  try {
    await conn
      .execute('DELETE FROM icons WHERE user_id = ?', [userId])
      .catch(throwErrorWith('failed to delete icon'))

    const [{ insertId }] = await conn
      .query<ResultSetHeader>(
        'INSERT INTO icons (user_id, image) VALUES (?, ?)',
        [userId, Buffer.from(body.image, 'base64')],
      )
      .catch(throwErrorWith('failed to insert icon'))

    await conn.commit().catch(throwErrorWith('failed to commit'))

    return c.json({ id: insertId }, 201)
  } catch (error) {
    await conn.rollback()
    return c.text(`Internal Server Error\n${error}`, 500)
  } finally {
    conn.release()
  }
})

userHandler.get(
  '/api/user/:username/icon',
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
        return c.body(await c.get('deps').fallbackUserIcon, 200, {
          'Content-Type': 'image/jpeg',
        })
        await conn.rollback()
        return c.text('not found icon', 404)
      }

      return c.body(icon.image, 200, {
        'Content-Type': 'image/jpeg',
      })
    } catch (error) {
      await conn.rollback()
      return c.text(`Internal Server Error\n${error}`, 500)
    } finally {
      conn.release()
    }
  },
)
