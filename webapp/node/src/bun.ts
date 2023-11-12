import { createApp } from './create-app'
import { Deps } from './types'

const deps = {
  exec: async (cmd: string[]) =>
    new Promise((resolve, reject) => {
      Bun.spawn(cmd, {
        async onExit(proc, exitCode, _signalCode, error) {
          if (
            typeof proc.stdout === 'number' ||
            typeof proc.stderr === 'number'
          ) {
            return reject(new Error('stdout/stderr is not a stream'))
          }
          const stdout = await new Response(proc.stdout).text()
          const stderr = await new Response(proc.stderr).text()

          if (exitCode === 0) {
            resolve({ stdout, stderr })
          } else {
            reject(error)
          }
        },
      })
    }),
  hashPassword: async (password: string) =>
    Bun.password.hash(password, { algorithm: 'bcrypt', cost: 4 }),
  comparePassword: async (password: string, hash: string) =>
    Bun.password.verify(password, hash, 'bcrypt'),
  uuid: () => crypto.randomUUID(),
} satisfies Deps

const main = async () => {
  const app = await createApp(deps)

  Bun.serve({ ...app, port: 8080 })
  console.log(`Listening on http://localhost:8080`)
}

main()
