/**
 * エラーの情報にメッセージを付与して、再度エラーを投げる関数
 * @param message エラーに付与する文字列
 *
 * @example
 * await fetch('https://example.com').catch(throwErrorWith('GET https://example.com failed'))
 * // => GET https://example.com failed
 * //    TypeError: Failed to fetch
 */
export const throwErrorWith =
  (message: string) =>
  (error: unknown): never => {
    throw `${message}\n${error}`
  }
