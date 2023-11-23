import { PoolConnection, RowDataPacket } from 'mysql2/promise'
import {
  LivecommentReportsModel,
  LivecommentsModel,
  UserModel,
} from '../types/models'
import { UserResponse, fillUserResponse } from './fill-user-response'
import {
  LivecommentResponse,
  fillLivecommentResponse,
} from './fill-livecomment-response'

export interface LivecommentReportResponse {
  id: number
  reporter: UserResponse
  livecomment: LivecommentResponse
  created_at: number
}

export const fillLivecommentReportResponse = async (
  conn: PoolConnection,
  livecommentReport: LivecommentReportsModel,
  getFallbackUserIcon: () => Promise<Readonly<ArrayBuffer>>,
) => {
  const [[user]] = await conn.query<(UserModel & RowDataPacket)[]>(
    'SELECT * FROM users WHERE id = ?',
    [livecommentReport.user_id],
  )
  if (!user) throw new Error('not found user that has the given id')

  const userResponse = await fillUserResponse(conn, user, getFallbackUserIcon)

  const [[livecomment]] = await conn.query<
    (LivecommentsModel & RowDataPacket)[]
  >('SELECT * FROM livecomments WHERE id = ?', [
    livecommentReport.livecomment_id,
  ])
  if (!livecomment)
    throw new Error('not found livecomment that has the given id')

  const livecommentResponse = await fillLivecommentResponse(
    conn,
    livecomment,
    getFallbackUserIcon,
  )

  return {
    id: livecommentReport.id,
    reporter: userResponse,
    livecomment: livecommentResponse,
    created_at: livecommentReport.created_at,
  } satisfies LivecommentReportResponse
}
