import { PoolConnection, RowDataPacket } from 'mysql2/promise'
import { LivecommentsModel, LivestreamsModel, UserModel } from '../types/models'
import { UserResponse, makeUserResponse } from './make-user-response'
import {
  LivestreamResponse,
  makeLivestreamResponse,
} from './make-livestream-response'
import { throwErrorWith } from './throw-error-with'

export interface LivecommentResponse {
  id: number
  user: UserResponse
  livestream: LivestreamResponse
  comment: string
  tip: number
  created_at: number
}

export const makeLivecommentResponse = async (
  conn: PoolConnection,
  livecomment: LivecommentsModel,
) => {
  const [[user]] = await conn
    .query<(UserModel & RowDataPacket)[]>('SELECT * FROM users WHERE id = ?', [
      livecomment.user_id,
    ])
    .catch(throwErrorWith('failed to get user'))
  if (!user) throw new Error('not found user that has the given id')

  const userResponse = await makeUserResponse(conn, user)

  const [[livestream]] = await conn
    .query<(LivestreamsModel & RowDataPacket)[]>(
      'SELECT * FROM livestreams WHERE id = ?',
      [livecomment.livestream_id],
    )
    .catch(throwErrorWith('failed to get livestream'))
  if (!livestream) throw new Error('not found livestream that has the given id')

  const livestreamResponse = await makeLivestreamResponse(conn, livestream)

  return {
    id: livecomment.id,
    user: userResponse,
    livestream: livestreamResponse,
    comment: livecomment.comment,
    tip: livecomment.tip,
    created_at: livecomment.created_at,
  } satisfies LivecommentResponse
}
