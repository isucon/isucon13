import { Hono } from 'hono'
import { logger } from 'hono/logger'
import { createConnection } from 'mysql2/promise'
import { ApplicationDeps, Deps } from './types'
import { userHandler } from './handlers/user-handler'

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

  const applicationDeps = {
    ...deps,
    connection,
    powerDNSSubdomainAddress,
  } satisfies ApplicationDeps

  const app = new Hono()
  app.use('*', logger())

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

  return app
}
