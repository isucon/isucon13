import { spawn } from 'node:child_process'
import { readFile } from 'node:fs/promises'
import { join } from 'node:path'
import { serve } from '@hono/node-server'
import { hash, compare } from 'bcrypt'
import { createApp } from './create-app'
import { Deps } from './types/application'

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
          reject(
            new Error(`command failed with code ${code}\n${stderr}\n${stdout}`),
          )
        }
      })
    }),
  hashPassword: async (password: string) => hash(password, 4),
  comparePassword: async (password: string, hash: string) =>
    compare(password, hash),
  // eslint-disable-next-line unicorn/prefer-module, unicorn/prefer-top-level-await
  fallbackUserIcon: readFile(join(__dirname, '../../img/NoImage.jpg')).then(
    (v) => {
      const buf = v.buffer
      if (buf instanceof ArrayBuffer) {
        return buf
      } else {
        throw new TypeError(`NoImage.jpg should be ArrayBuffer, but ${buf}`)
      }
    },
  ),
} satisfies Deps

const app = createApp(deps)

serve({ ...app, port: 8080 }, (add) =>
  console.log(`Listening on http://localhost:${add.port}`),
)