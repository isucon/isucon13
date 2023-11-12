import { PoolConnection, RowDataPacket } from 'mysql2/promise'
import { LivestreamsModel, ReactionsModel, UserModel } from '../types/models'
import {
  LivestreamResponse,
  makeLivestreamResponse,
} from './make-livestream-response'
import { UserResponse, makeUserResponse } from './make-user-response'
import { throwErrorWith } from './throw-error-with'

export interface ReactionResponse {
  id: number
  emoji_name: string
  user: UserResponse
  livestream: LivestreamResponse
  created_at: number
}

export const makeReactionResponse = async (
  conn: PoolConnection,
  reaction: ReactionsModel,
) => {
  const [[user]] = await conn
    .query<(UserModel & RowDataPacket)[]>('SELECT * FROM users WHERE id = ?', [
      reaction.user_id,
    ])
    .catch(throwErrorWith('failed to get user'))
  if (!user) throw new Error('not found user that has the given id')

  const userResponse = await makeUserResponse(conn, user)

  const [[livestream]] = await conn
    .query<(LivestreamsModel & RowDataPacket)[]>(
      'SELECT * FROM livestreams WHERE id = ?',
      [reaction.livestream_id],
    )
    .catch(throwErrorWith('failed to get livestream'))
  if (!livestream) throw new Error(`not found livestream that has the given id`)

  const livestreamResponse = await makeLivestreamResponse(conn, livestream)

  return {
    id: reaction.id,
    emoji_name: reaction.emoji_name,
    user: userResponse,
    livestream: livestreamResponse,
    created_at: reaction.created_at,
  } satisfies ReactionResponse
}
