import { RowDataPacket } from 'mysql2'
import { ApplicationDeps } from '../types'
import { UserResponse, makeUserResponse } from './make-user-response'
import {
  LivestreamResponse,
  makeLivestreamResponse,
} from './make-livestream-response'

export interface LivecommentResponse {
  id: number
  user: UserResponse
  livestream: LivestreamResponse
  comment: string
  tip: number
  created_at: number
}

export const makeLivecommentResponse = async (
  deps: ApplicationDeps,
  livecomment: {
    id: number
    user_id: number
    livestream_id: number
    comment: string
    tip: number
    created_at: number
  },
) => {
  const user = await deps.connection
    .query<RowDataPacket[]>('SELECT * FROM users WHERE id = ?', [
      livecomment.user_id,
    ])
    .then(([[user]]) => ({ ok: true, data: user }) as const)
    .catch((error) => ({ ok: false, error }) as const)
  if (!user.ok) {
    return { ok: false, error: 'failed to fetch user' } as const
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
      livecomment.livestream_id,
    ])
    .then(([[result]]) => ({ ok: true, data: result }) as const)
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
    playlist_url: livestream.data.playlist_url,
    thumbnail_url: livestream.data.thumbnail_url,
    start_at: livestream.data.start_at,
    end_at: livestream.data.end_at,
  })
  if (!livestreamResponse.ok) {
    return livestreamResponse
  }

  return {
    ok: true,
    data: {
      id: livecomment.id,
      user: userResponse.data,
      livestream: livestreamResponse.data,
      comment: livecomment.comment,
      tip: livecomment.tip,
      created_at: livecomment.created_at,
    } satisfies LivecommentResponse,
  } as const
}
