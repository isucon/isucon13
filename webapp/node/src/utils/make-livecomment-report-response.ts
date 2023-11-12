import { PoolConnection, RowDataPacket } from 'mysql2/promise'
import {
  LivecommentReportsModel,
  LivecommentsModel,
  UserModel,
} from '../types/models'
import { UserResponse, makeUserResponse } from './make-user-response'
import {
  LivecommentResponse,
  makeLivecommentResponse,
} from './make-livecomment-response'
import { throwErrorWith } from './throw-error-with'

export interface LivecommentReportResponse {
  id: number
  reporter: UserResponse
  livecomment: LivecommentResponse
  created_at: number
}

export const makeLivecommentReportResponse = async (
  conn: PoolConnection,
  livecommentReport: LivecommentReportsModel,
) => {
  const [[user]] = await conn
    .query<(UserModel & RowDataPacket)[]>('SELECT * FROM users WHERE id = ?', [
      livecommentReport.user_id,
    ])
    .catch(throwErrorWith('failed to get user'))
  if (!user) throw new Error('not found user that has the given id')

  const userResponse = await makeUserResponse(conn, user)

  const [[livecomment]] = await conn
    .query<(LivecommentsModel & RowDataPacket)[]>(
      'SELECT * FROM livecomments WHERE id = ?',
      [livecommentReport.livecomment_id],
    )
    .catch(throwErrorWith('failed to get livecomment'))
  if (!livecomment)
    throw new Error('not found livecomment that has the given id')

  const livecommentResponse = await makeLivecommentResponse(conn, livecomment)

  return {
    id: livecommentReport.id,
    reporter: userResponse,
    livecomment: livecommentResponse,
    created_at: livecommentReport.created_at,
  } satisfies LivecommentReportResponse
}
