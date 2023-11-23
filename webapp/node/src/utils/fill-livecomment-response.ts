import { PoolConnection, RowDataPacket } from 'mysql2/promise'
import { LivecommentsModel, LivestreamsModel, UserModel } from '../types/models'
import { UserResponse, fillUserResponse } from './fill-user-response'
import {
  LivestreamResponse,
  fillLivestreamResponse,
} from './fill-livestream-response'

export interface LivecommentResponse {
  id: number
  user: UserResponse
  livestream: LivestreamResponse
  comment: string
  tip: number
  created_at: number
}

export const fillLivecommentResponse = async (
  conn: PoolConnection,
  livecomment: LivecommentsModel,
  getFallbackUserIcon: () => Promise<Readonly<ArrayBuffer>>,
) => {
  const [[user]] = await conn.query<(UserModel & RowDataPacket)[]>(
    'SELECT * FROM users WHERE id = ?',
    [livecomment.user_id],
  )
  if (!user) throw new Error('not found user that has the given id')

  const userResponse = await fillUserResponse(conn, user, getFallbackUserIcon)

  const [[livestream]] = await conn.query<(LivestreamsModel & RowDataPacket)[]>(
    'SELECT * FROM livestreams WHERE id = ?',
    [livecomment.livestream_id],
  )
  if (!livestream) throw new Error('not found livestream that has the given id')

  const livestreamResponse = await fillLivestreamResponse(
    conn,
    livestream,
    getFallbackUserIcon,
  )

  return {
    id: livecomment.id,
    user: userResponse,
    livestream: livestreamResponse,
    comment: livecomment.comment,
    tip: livecomment.tip,
    created_at: livecomment.created_at,
  } satisfies LivecommentResponse
}
