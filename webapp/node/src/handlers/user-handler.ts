import { Hono } from 'hono'
import { ResultSetHeader, RowDataPacket } from 'mysql2/promise'
import { ApplicationDeps } from '../types'

export const userHandler = (deps: ApplicationDeps) => {
  const handler = new Hono()

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
      .then(([{ insertId }]) => ({ ok: true, data: { insertId } }) as const)
      .catch((error) => ({ ok: false, error }) as const)
    if (!result.ok) {
      await deps.connection.rollback()
      return c.text(`failed to insert user\n${result.error}`, 500)
    }
    const userId = result.data.insertId

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
      .then(([[theme]]) => ({ ok: true, data: { theme } }) as const)
      .catch((error) => ({ ok: false, error }) as const)
    if (!theme.ok) {
      await deps.connection.rollback()
      return c.text('failed to fetch theme', 500)
    }
    console.log(theme.data)
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
          id: theme.data.theme.id,
          dark_mode: !!theme.data.theme.dark_mode,
        },
      },
      201,
    )
  })

  return handler
}
