import { RowDataPacket } from 'mysql2'
import { ApplicationDeps } from '../types'

export interface UserResponse {
  id: number
  name: string
  display_name: string
  description: string
  theme: {
    id: number
    dark_mode: boolean
  }
}

export const makeUserResponse = async (
  deps: ApplicationDeps,
  user: { id: number; name: string; display_name: string; description: string },
) => {
  const theme = await deps.connection
    .query<RowDataPacket[]>('SELECT * FROM themes WHERE user_id = ?', [user.id])
    .then(([[theme]]) => ({ ok: true, data: theme }) as const)
    .catch((error) => ({ ok: false, error }) as const)
  if (!theme.ok) {
    return { ok: false, error: 'failed to fetch theme' } as const
  }

  return {
    ok: true,
    data: {
      id: user.id,
      name: user.name,
      display_name: user.display_name,
      description: user.description,
      theme: {
        id: theme.data.id,
        dark_mode: !!theme.data.dark_mode,
      },
    } satisfies UserResponse,
  } as const
}
