import { RowDataPacket } from 'mysql2'
import { ApplicationDeps } from '../types'
import { UserResponse, makeUserResponse } from './make-user-response'
import {
  LivecommentResponse,
  makeLivecommentResponse,
} from './make-livecomment-response'

export interface LivecommentReportResponse {
  id: number
  reporter: UserResponse
  livecomment: LivecommentResponse
  created_at: number
}

export const makeLivecommentReportResponse = async (
  deps: ApplicationDeps,
  livecommentReport: {
    id: number
    user_id: number
    livestream_id: number
    livecomment_id: number
    created_at: number
  },
) => {
  const user = await deps.connection
    .query<RowDataPacket[]>('SELECT * FROM users WHERE id = ?', [
      livecommentReport.user_id,
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

  const livecomment = await deps.connection
    .query<RowDataPacket[]>('SELECT * FROM livecomments WHERE id = ?', [
      livecommentReport.livecomment_id,
    ])
    .then(([[livecomment]]) => ({ ok: true, data: livecomment }) as const)
    .catch((error) => ({ ok: false, error }) as const)
  if (!livecomment.ok) {
    return { ok: false, error: 'failed to get livecomment' } as const
  }
  if (!livecomment.data) {
    return {
      ok: false,
      error: 'not found livecomment that has the given id',
    } as const
  }
  const livecommentResponse = await makeLivecommentResponse(deps, {
    id: livecomment.data.id,
    user_id: livecomment.data.user_id,
    livestream_id: livecomment.data.livestream_id,
    comment: livecomment.data.comment,
    tip: livecomment.data.tip,
    created_at: livecomment.data.created_at,
  })
  if (!livecommentResponse.ok) {
    return livecommentResponse
  }

  return {
    ok: true,
    data: {
      id: livecommentReport.id,
      reporter: userResponse.data,
      livecomment: livecommentResponse.data,
      created_at: livecommentReport.created_at,
    } satisfies LivecommentReportResponse,
  } as const
}
