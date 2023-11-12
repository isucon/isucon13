import { Hono } from 'hono'
import { RowDataPacket } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
import { throwErrorWith } from '../utils/throw-error-with'

export const paymentHandler = new Hono<HonoEnvironment>()

paymentHandler.get('/api/payment', async (c) => {
  const conn = await c.get('pool').getConnection()
  await conn.beginTransaction()

  try {
    const [[{ 'IFNULL(SUM(tip), 0)': totalTip }]] = await conn
      .query<({ 'IFNULL(SUM(tip), 0)': number } & RowDataPacket)[]>(
        'SELECT IFNULL(SUM(tip), 0) FROM livecomments',
      )
      .catch(throwErrorWith('failed to count total tip'))

    return c.json({ totalTip: totalTip })
  } catch (error) {
    await conn.rollback()
    return c.text(`Internal Server Error\n${error}`, 500)
  } finally {
    conn.release()
  }
})
