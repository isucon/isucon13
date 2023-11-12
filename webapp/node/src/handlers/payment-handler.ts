import { Hono } from 'hono'
import { RowDataPacket } from 'mysql2/promise'
import { ApplicationDeps, HonoEnvironment } from '../types'

export const paymentHandler = (deps: ApplicationDeps) => {
  const handler = new Hono<HonoEnvironment>()

  handler.get('/api/payment', async (c) => {
    await deps.connection.beginTransaction()

    const totalTip = await deps.connection
      .query<RowDataPacket[]>('SELECT IFNULL(SUM(tip), 0) FROM livecomments')
      .then(
        ([[row]]) => ({ ok: true, data: row['IFNULL(SUM(tip), 0)'] }) as const,
      )
      .catch((error) => ({ ok: false, error: error }) as const)
    if (!totalTip.ok) {
      await deps.connection.rollback()
      return c.text('failed to count total tip')
    }

    return c.json({ totalTip: totalTip.data })
  })

  return handler
}
