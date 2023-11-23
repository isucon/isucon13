import { createHash } from 'node:crypto'
import { readFile } from 'node:fs/promises'
import { join } from 'node:path'
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
    // eslint-disable-next-line unicorn/prefer-module
    const result = await readFile(join(__dirname, '../../../img/NoImage.jpg'))
    const buf = result.buffer
    if (buf instanceof ArrayBuffer) {
      image = buf
    } else {
      throw new TypeError(`NoImage.jpg should be ArrayBuffer, but ${buf}`)
    }
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
