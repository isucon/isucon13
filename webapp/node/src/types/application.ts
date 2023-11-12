import { Env } from 'hono'
import { Session } from 'hono-sessions'
import { Pool } from 'mysql2/promise'

export interface Deps {
  exec: (cmd: string[]) => Promise<{ stdout: string; stderr: string }>
  hashPassword: (password: string) => Promise<string>
  comparePassword: (password: string, hash: string) => Promise<boolean>
  // TODO: remove
  uuid: () => string
}

export interface ApplicationDeps extends Deps {
  pool: Pool
  powerDNSSubdomainAddress: string
}

export interface HonoEnvironment extends Env {
  Variables: {
    session: Session
    session_key_rotation: boolean
  }
}
