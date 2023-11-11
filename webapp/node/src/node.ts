import { spawn } from 'node:child_process'
import { randomUUID } from 'node:crypto'
import { serve } from '@hono/node-server'
import { hash, compare } from 'bcrypt'
import { createApp } from './create-app'
import { Deps } from './types'

const deps = {
  exec: async (cmd: string[]) =>
    new Promise((resolve, reject) => {
      const proc = spawn(cmd[0], cmd.slice(1))
      let stdout = ''
      let stderr = ''
      proc.stdout.on('data', (data) => (stdout += data))
      proc.stderr.on('data', (data) => (stderr += data))
      proc.on('close', (code) => {
        if (code === 0) {
          resolve({ stdout, stderr })
        } else {
          reject(new Error(`command failed with code ${code}`))
        }
      })
    }),
  hashPassword: async (password: string) => hash(password, 4),
  comparePassword: async (password: string, hash: string) =>
    compare(password, hash),
  uuid: () => randomUUID(),
} satisfies Deps

const main = async () => {
  serve({ ...(await createApp(deps)), port: 8080 }, (add) =>
    console.log(`Listening on http://localhost:${add.port}`),
  )
}

main()
