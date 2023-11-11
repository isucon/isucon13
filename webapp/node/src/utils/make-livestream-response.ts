import { RowDataPacket } from 'mysql2'
import { ApplicationDeps } from '../types'
import { makeUserResponse } from './make-user-response'

export const makeLivestreamResponse = async (
  deps: ApplicationDeps,
  livestream: {
    id: number
    user_id: number
    title: string
    description: string
    playlist_url: string
    thumbnail_url: string
    start_at: number
    end_at: number
  },
) => {
  const user = await deps.connection
    .query<RowDataPacket[]>('SELECT * FROM users WHERE id = ?', [
      livestream.user_id,
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

  const livestreamTags = await deps.connection
    .query<RowDataPacket[]>(
      'SELECT * FROM livestream_tags WHERE livestream_id = ?',
      [livestream.id],
    )
    .then(([results]) => ({ ok: true, data: results }) as const)
    .catch((error) => ({ ok: false, error }) as const)
  if (!livestreamTags.ok) {
    return { ok: false, error: 'failed to fetch livestream tags' } as const
  }

  const tags: RowDataPacket[] = []
  for (const livestreamTag of livestreamTags.data) {
    const tag = await deps.connection
      .query<RowDataPacket[]>('SELECT * FROM tags WHERE id = ?', [
        livestreamTag.tag_id,
      ])
      .then(([[result]]) => ({ ok: true, data: result }) as const)
      .catch((error) => ({ ok: false, error }) as const)
    if (!tag.ok) {
      return { ok: false, error: 'failed to fetch tag' } as const
    }
    tags.push(tag.data)
  }

  return {
    ok: true,
    data: {
      id: livestream.id,
      owner: userResponse.data,
      title: livestream.title,
      tags: tags.map((tag) => ({ id: tag.id, name: tag.name })),
      description: livestream.description,
      playlist_url: livestream.playlist_url,
      thumbnail_url: livestream.thumbnail_url,
      start_at: livestream.start_at,
      end_at: livestream.end_at,
    },
  } as const
}
