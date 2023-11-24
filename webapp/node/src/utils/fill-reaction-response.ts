import { PoolConnection, RowDataPacket } from 'mysql2/promise'
import { LivestreamsModel, ReactionsModel, UserModel } from '../types/models'
import {
  LivestreamResponse,
  fillLivestreamResponse,
} from './fill-livestream-response'
import { UserResponse, fillUserResponse } from './fill-user-response'

export interface ReactionResponse {
  id: number
  emoji_name: string
  user: UserResponse
  livestream: LivestreamResponse
  created_at: number
}

export const fillReactionResponse = async (
  conn: PoolConnection,
  reaction: ReactionsModel,
  getFallbackUserIcon: () => Promise<Readonly<ArrayBuffer>>,
) => {
  const [[user]] = await conn.query<(UserModel & RowDataPacket)[]>(
    'SELECT * FROM users WHERE id = ?',
    [reaction.user_id],
  )
  if (!user) throw new Error('not found user that has the given id')

  const userResponse = await fillUserResponse(conn, user, getFallbackUserIcon)

  const [[livestream]] = await conn.query<(LivestreamsModel & RowDataPacket)[]>(
    'SELECT * FROM livestreams WHERE id = ?',
    [reaction.livestream_id],
  )
  if (!livestream) throw new Error(`not found livestream that has the given id`)

  const livestreamResponse = await fillLivestreamResponse(
    conn,
    livestream,
    getFallbackUserIcon,
  )

  return {
    id: reaction.id,
    emoji_name: reaction.emoji_name,
    user: userResponse,
    livestream: livestreamResponse,
    created_at: reaction.created_at,
  } satisfies ReactionResponse
}
