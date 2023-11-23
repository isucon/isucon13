import { Env } from 'hono'
import { Session } from 'hono-sessions'
import { Pool } from 'mysql2/promise'

export interface Runtime {
  exec: (cmd: string[]) => Promise<{ stdout: string; stderr: string }>
  hashPassword: (password: string) => Promise<string>
  comparePassword: (password: string, hash: string) => Promise<boolean>
  fallbackUserIcon: () => Promise<Readonly<ArrayBuffer>>
}

export interface ApplicationRuntime extends Runtime {
  powerDNSSubdomainAddress: string
}

export interface HonoEnvironment extends Env {
  Variables: {
    session: Session
    session_key_rotation: boolean
    pool: Pool
    runtime: ApplicationRuntime
  }
}
