import { Hono } from 'hono'
import { ResultSetHeader, RowDataPacket } from 'mysql2/promise'
import { ApplicationDeps, HonoEnvironment } from '../types'
import {
  defaultUserIDKey,
  defaultUserNameKey,
  defaultSessionExpiresKey,
} from '../contants'
import { verifyUserSessionMiddleware } from '../middlewares/verify-user-session-middleare'

export const userHandler = (deps: ApplicationDeps) => {
  const handler = new Hono<HonoEnvironment>()

  handler.post('/api/register', async (c) => {
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

    const hashedPassword = await deps.hashPassword(body.password)

    await deps.connection.beginTransaction()
    const result = await deps.connection
      .execute<ResultSetHeader>(
        'INSERT INTO users (name, display_name, description, password) VALUES(?, ?, ?, ?)',
        [body.name, body.display_name, body.description, hashedPassword],
      )
      .then(([{ insertId }]) => ({ ok: true, data: insertId }) as const)
      .catch((error) => ({ ok: false, error }) as const)
    if (!result.ok) {
      await deps.connection.rollback()
      return c.text(`failed to insert user\n${result.error}`, 500)
    }
    const userId = result.data

    try {
      await deps.connection.execute(
        'INSERT INTO themes (user_id, dark_mode) VALUES(?, ?)',
        [userId, body.theme.dark_mode],
      )
    } catch {
      await deps.connection.rollback()
      return c.text('failed to insert user theme', 500)
    }

    try {
      await deps.exec([
        'pdnsutil',
        'add-record',
        'u.isucon.dev',
        body.name,
        'A',
        '30',
        deps.powerDNSSubdomainAddress,
      ])
    } catch (error) {
      await deps.connection.rollback()
      return c.text(String(error), 500)
    }

    const theme = await deps.connection
      .query<RowDataPacket[]>('SELECT * FROM themes WHERE user_id = ?', [
        userId,
      ])
      .then(([[theme]]) => ({ ok: true, data: theme }) as const)
      .catch((error) => ({ ok: false, error }) as const)
    if (!theme.ok) {
      await deps.connection.rollback()
      return c.text('failed to fetch theme', 500)
    }
    try {
      await deps.connection.commit()
    } catch {
      await deps.connection.rollback()
      return c.text('failed to commit', 500)
    }

    return c.json(
      {
        id: userId,
        name: body.name,
        display_name: body.display_name,
        description: body.description,
        theme: {
          id: theme.data.id,
          dark_mode: !!theme.data.dark_mode,
        },
      },
      201,
    )
  })

  handler.post('/api/login', async (c) => {
    const body = await c.req.json<{
      username: string
      password: string
    }>()

    await deps.connection.beginTransaction()
    const user = await deps.connection
      .query<RowDataPacket[]>('SELECT * FROM users WHERE name = ?', [
        body.username,
      ])
      .then(([[user]]) => ({ ok: true, data: user }) as const)
      .catch((error) => ({ ok: false, error }) as const)
    if (!user.ok) {
      await deps.connection.rollback()
      return c.text('invalid username or password', 401)
    }
    try {
      await deps.connection.commit()
    } catch {
      await deps.connection.rollback()
      return c.text('failed to commit', 500)
    }
    const isPasswordMatch = await deps.comparePassword(
      body.password,
      user.data.password,
    )
    if (!isPasswordMatch) {
      return c.text('invalid username or password', 401)
    }

    // 1時間でセッションが切れるようにする
    const sessionEndAt = Date.now() + 1000 * 60 * 60

    const session = c.get('session')
    session.set(defaultUserIDKey, user.data.id)
    session.set(defaultUserNameKey, user.data.name)
    session.set(defaultSessionExpiresKey, sessionEndAt)

    // eslint-disable-next-line unicorn/no-null
    return c.body(null, 200)
  })

  handler.get('/api/user/me', verifyUserSessionMiddleware, async (c) => {
    const userId = c.get('session').get(defaultUserIDKey) as number // userId is verified by verifyUserSessionMiddleware

    await deps.connection.beginTransaction()
    const user = await deps.connection
      .query<RowDataPacket[]>('SELECT * FROM users WHERE id = ?', [userId])
      .then(([[user]]) => ({ ok: true, data: user }) as const)
      .catch((error) => ({ ok: false, error }) as const)
    if (!user.ok) {
      await deps.connection.rollback()
      return c.text('failed to get user', 500)
    }
    if (!user.data) {
      await deps.connection.rollback()
      return c.text('not found user that has the userid in session', 404)
    }

    const theme = await deps.connection
      .query<RowDataPacket[]>('SELECT * FROM themes WHERE user_id = ?', [
        userId,
      ])
      .then(([[theme]]) => ({ ok: true, data: theme }) as const)
      .catch((error) => ({ ok: false, error }) as const)
    if (!theme.ok) {
      await deps.connection.rollback()
      return c.text('failed to fetch theme', 500)
    }

    try {
      await deps.connection.commit()
    } catch {
      await deps.connection.rollback()
      return c.text('failed to commit', 500)
    }

    return c.json(
      {
        id: userId,
        name: user.data.name,
        display_name: user.data.display_name,
        description: user.data.description,
        theme: {
          id: theme.data.id,
          dark_mode: !!theme.data.dark_mode,
        },
      },
      200,
    )
  })

  handler.get('/api/user/:username', verifyUserSessionMiddleware, async (c) => {
    const username = c.req.param('username')

    await deps.connection.beginTransaction()

    const user = await deps.connection
      .query<RowDataPacket[]>('SELECT * FROM users WHERE name = ?', [username])
      .then(([[user]]) => ({ ok: true, data: user }) as const)
      .catch((error) => ({ ok: false, error }) as const)
    if (!user.ok) {
      await deps.connection.rollback()
      return c.text('failed to get user', 500)
    }
    if (!user.data) {
      await deps.connection.rollback()
      return c.text('not found user that has the given username', 404)
    }

    const theme = await deps.connection
      .query<RowDataPacket[]>('SELECT * FROM themes WHERE user_id = ?', [
        user.data.id,
      ])
      .then(([[theme]]) => ({ ok: true, data: theme }) as const)
      .catch((error) => ({ ok: false, error }) as const)
    if (!theme.ok) {
      await deps.connection.rollback()
      return c.text('failed to fetch theme', 500)
    }

    try {
      await deps.connection.commit()
    } catch {
      await deps.connection.rollback()
      return c.text('failed to commit', 500)
    }

    return c.json(
      {
        id: user.data.id,
        name: user.data.name,
        display_name: user.data.display_name,
        description: user.data.description,
        theme: {
          id: theme.data.id,
          dark_mode: !!theme.data.dark_mode,
        },
      },
      200,
    )
  })

  return handler
}
