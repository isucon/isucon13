import { createHash } from 'node:crypto'
import { PoolConnection, RowDataPacket } from 'mysql2/promise'
import { IconModel, ThemeModel, UserModel } from '../types/models'

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
  getFallbackUserIcon: () => Promise<Readonly<ArrayBuffer>>,
) => {
  const [[theme]] = await conn.query<(ThemeModel & RowDataPacket)[]>(
    'SELECT * FROM themes WHERE user_id = ?',
    [user.id],
  )

  const [[icon]] = await conn.query<
    (Pick<IconModel, 'image'> & RowDataPacket)[]
  >('SELECT image FROM icons WHERE user_id = ?', [user.id])

  let image = icon?.image

  if (!image) {
    image = await getFallbackUserIcon()
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
