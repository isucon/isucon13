export const throwErrorWith =
  (message: string) =>
  (error: unknown): never => {
    throw new Error(`${message}\n${error}`)
  }
