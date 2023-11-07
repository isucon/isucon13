export function iconUrl(username: string | undefined): string {
  return `/api/user/${username ?? 0}/icon`;
}
