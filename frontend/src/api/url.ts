export function normalizeUrl(
  path: string,
  tenant?: string | undefined,
): string {
  const hostname = tenant ? `${tenant}.u.isucon.dev` : 'pipe.u.isucon.dev';
  const port = window.location.port;
  return `https://${hostname}${port ? `:${port}` : ''}${path}`;
}
