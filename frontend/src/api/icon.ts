export function iconUrl(username: string | undefined): string | undefined {
  return username && `/api/user/${username ?? 0}/icon`;
}
