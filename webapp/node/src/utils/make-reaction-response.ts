import { RowDataPacket } from 'mysql2'
import { ApplicationDeps } from '../types'
import {
  LivestreamResponse,
  makeLivestreamResponse,
} from './make-livestream-response'
import { UserResponse, makeUserResponse } from './make-user-response'

export interface ReactionResponse {
  id: number
  emoji_name: string
  user: UserResponse
  livestream: LivestreamResponse
  created_at: number
}

export const makeReactionResponse = async (
  deps: ApplicationDeps,
  reaction: {
    id: number
    emoji_name: string
    user_id: number
    livestream_id: number
    created_at: number
  },
) => {
  const user = await deps.connection
    .query<RowDataPacket[]>('SELECT * FROM users WHERE id = ?', [
      reaction.user_id,
    ])
    .then(([[user]]) => ({ ok: true, data: user }) as const)
    .catch((error) => ({ ok: false, error }) as const)
  if (!user.ok) {
    return { ok: false, error: 'failed to get user' } as const
  }
  if (!user.data) {
    return { ok: false, error: 'not found user that has the given id' } as const
  }
  const userResponse = await makeUserResponse(deps, {
    id: user.data.id,
    name: user.data.name,
    display_name: user.data.display_name,
    description: user.data.description,
  })
  if (!userResponse.ok) {
    return userResponse
  }

  const livestream = await deps.connection
    .query<RowDataPacket[]>('SELECT * FROM livestreams WHERE id = ?', [
      reaction.livestream_id,
    ])
    .then(([[livestream]]) => ({ ok: true, data: livestream }) as const)
    .catch((error) => ({ ok: false, error }) as const)
  if (!livestream.ok) {
    return { ok: false, error: 'failed to get livestream' } as const
  }
  if (!livestream.data) {
    return {
      ok: false,
      error: 'not found livestream that has the given id',
    } as const
  }
  const livestreamResponse = await makeLivestreamResponse(deps, {
    id: livestream.data.id,
    user_id: livestream.data.user_id,
    title: livestream.data.title,
    description: livestream.data.description,
    thumbnail_url: livestream.data.thumbnail_url,
    playlist_url: livestream.data.playlist_url,
    start_at: livestream.data.start_at,
    end_at: livestream.data.end_at,
  })
  if (!livestreamResponse.ok) {
    return livestreamResponse
  }

  return {
    ok: true,
    data: {
      id: reaction.id,
      emoji_name: reaction.emoji_name,
      user: userResponse.data,
      livestream: livestreamResponse.data,
      created_at: reaction.created_at,
    } satisfies ReactionResponse,
  } as const
}
