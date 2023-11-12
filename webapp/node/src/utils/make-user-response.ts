import { PoolConnection, RowDataPacket } from 'mysql2/promise'
import { ThemeModel, UserModel } from '../types/models'
import { throwErrorWith } from './throw-error-with'

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
  conn: PoolConnection,
  user: Omit<UserModel, 'password'>,
) => {
  const [[theme]] = await conn
    .query<(ThemeModel & RowDataPacket)[]>(
      'SELECT * FROM themes WHERE user_id = ?',
      [user.id],
    )
    .catch(throwErrorWith('failed to get theme'))

  return {
    id: user.id,
    name: user.name,
    display_name: user.display_name,
    description: user.description,
    theme: {
      id: theme.id,
      dark_mode: !!theme.dark_mode,
    },
  } satisfies UserResponse
}
