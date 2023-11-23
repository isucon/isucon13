import { createHash } from 'node:crypto'
import { PoolConnection, RowDataPacket } from 'mysql2/promise'
import { IconModel, ThemeModel, UserModel } from '../types/models'
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
  icon_hash: string
}

export const fillUserResponse = async (
  conn: PoolConnection,
  user: Omit<UserModel, 'password'>,
  fallbackUserIcon: Readonly<ArrayBuffer>,
) => {
  const [[theme]] = await conn
    .query<(ThemeModel & RowDataPacket)[]>(
      'SELECT * FROM themes WHERE user_id = ?',
      [user.id],
    )
    .catch(throwErrorWith('failed to get theme'))

  const [[icon]] = await conn.query<
    (Pick<IconModel, 'image'> & RowDataPacket)[]
  >('SELECT image FROM icons WHERE user_id = ?', [user.id])

  let image = icon?.image

  if (!image) {
    image = fallbackUserIcon
  }

  return {
    id: user.id,
    name: user.name,
    display_name: user.display_name,
    description: user.description,
    theme: {
      id: theme.id,
      dark_mode: !!theme.dark_mode,
    },
    icon_hash: createHash('sha256').update(new Uint8Array(image)).digest('hex'),
  } satisfies UserResponse
}
