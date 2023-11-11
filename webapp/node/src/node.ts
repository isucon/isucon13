import { exec } from 'node:child_process'
import { promisify } from 'node:util'
import { serve } from '@hono/node-server'
import { Deps, createApp } from './create-app'

const execAsync = promisify(exec)

const deps = {
  exec: async (cmd: string) => execAsync(cmd),
} satisfies Deps

serve({ ...createApp(deps), port: 8080 }, (add) =>
  console.log(`Listening on http://localhost:${add.port}`),
)
