import { Context } from 'hono'
import { RowDataPacket } from 'mysql2/promise'
import { HonoEnvironment } from '../types/application'
import { throwErrorWith } from '../utils/throw-error-with'

// GET /api/payment
export const GetPaymentResult = async (
  c: Context<HonoEnvironment, '/api/payment'>,
) => {
  const conn = await c.get('pool').getConnection()
  await conn.beginTransaction()

  try {
    const [[{ 'IFNULL(SUM(tip), 0)': totalTip }]] = await conn
      .query<({ 'IFNULL(SUM(tip), 0)': number } & RowDataPacket)[]>(
        'SELECT IFNULL(SUM(tip), 0) FROM livecomments',
      )
      .catch(throwErrorWith('failed to count total tip'))

    await conn.commit().catch(throwErrorWith('failed to commit'))

    return c.json({ totalTip: totalTip })
  } catch (error) {
    await conn.rollback()
    return c.text(`Internal Server Error\n${error}`, 500)
  } finally {
    await conn.rollback()
    conn.release()
  }
}
