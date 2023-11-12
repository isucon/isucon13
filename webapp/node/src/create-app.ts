import { Hono } from 'hono'
import { logger } from 'hono/logger'
import { createConnection } from 'mysql2/promise'
import { sessionMiddleware, CookieStore } from 'hono-sessions'
import { ApplicationDeps, Deps, HonoEnvironment } from './types'
import { userHandler } from './handlers/user-handler'
import { topHandler } from './handlers/top-handler'
import { livestreamHandler } from './handlers/livestream-handler'
import { livecommentHandler } from './handlers/livecomment-handler'
import { reactionHandler } from './handlers/reaction-handler'
import { statsHandler } from './handlers/stats-handler'
import { paymentHandler } from './handlers/payment-handler'

export const createApp = async (deps: Deps) => {
  const connection = await createConnection({
    user: process.env['ISUCON13_MYSQL_DIALCONFIG_USER'] ?? 'isucon',
    password: process.env['ISUCON13_MYSQL_DIALCONFIG_PASSWORD'] ?? 'isucon',
    database: process.env['ISUCON13_MYSQL_DIALCONFIG_DATABASE'] ?? 'isupipe',
    host: process.env['ISUCON13_MYSQL_DIALCONFIG_ADDRESS'] ?? '127.0.0.1',
    port: Number(process.env['ISUCON13_MYSQL_DIALCONFIG_PORT'] ?? '3306'),
  })

  if (!process.env['ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS']) {
    throw new Error(
      'envionment variable ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS is not set',
    )
  }
  const powerDNSSubdomainAddress =
    process.env['ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS']

  const store = new CookieStore()

  const applicationDeps = {
    ...deps,
    connection,
    powerDNSSubdomainAddress,
  } satisfies ApplicationDeps

  const app = new Hono<HonoEnvironment>()
  app.use('*', logger())
  app.use(
    '*',
    sessionMiddleware({
      store,
      encryptionKey: '24553845-c33d-4a87-b0c3-f7a0e17fd82f',
      cookieOptions: {
        path: '/',
        domain: 'u.isucon.dev',
        maxAge: 60_000 /* 10 seconds */, // FIXME: 600
      },
    }),
  )
  app.use('*', async (c, next) => {
    await next()
    if (c.res.status >= 500) {
      console.error(c.res.status, await c.res.clone().text())
    }
  })

  app.post('/api/initialize', async (c) => {
    try {
      await deps.exec(['../sql/init.sh'])
      return c.json({ advertise_level: 10, advertise_name: 'node' })
    } catch (error) {
      console.log('init.sh failed with')
      console.log(error)
      return c.text('failed to initialize', 500)
    }
  })

  app.route('/', userHandler(applicationDeps))
  app.route('/', topHandler(applicationDeps))
  app.route('/', livestreamHandler(applicationDeps))
  app.route('/', livecommentHandler(applicationDeps))
  app.route('/', reactionHandler(applicationDeps))
  app.route('/', statsHandler(applicationDeps))
  app.route('/', paymentHandler(applicationDeps))

  return app
}
