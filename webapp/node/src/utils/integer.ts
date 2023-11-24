/**
 * Goのstrconv.Atoiを模した関数
 * Integerを変換した値を返す
 * Integerに変換できない値が来たときはfalseを返す
 */
export const atoi = (string_: string): number | false => {
  const number_ = Number(string_)
  const isInt = Number.isInteger(number_)
  if (isInt) return number_
  return false
}
