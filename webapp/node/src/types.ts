import { Connection } from 'mysql2/promise'

export interface Deps {
  exec: (cmd: string[]) => Promise<{ stdout: string; stderr: string }>
  hashPassword: (password: string) => Promise<string>
}

export interface ApplicationDeps extends Deps {
  connection: Connection
  powerDNSSubdomainAddress: string
}
