import { Hono } from 'hono'

export interface Deps {
  exec: (cmd: string) => Promise<{ stdout: string; stderr: string }>
}

export const createApp = (deps: Deps) => {
  const app = new Hono()

  app.post('/api/initialize', async (c) => {
    try {
      await deps.exec('../sql/init.sh')
      return c.json({ advertise_level: 10, advertise_name: 'node' })
    } catch (error) {
      console.log('init.sh failed with')
      console.log(error)
      return c.text('failed to initialize', 500)
    }
  })

  return app
}
