package config

// 2024-04-01 01:00:00
// NOTE: 2024-04-01 00:00:00 ~ 2024-04-01 01:00:00は初期データで予約済み
const BaseAt = 1700874000

// 同時配信枠数
// NOTE: ベンチマーカー調整項目
const NumSlots = 5

// NOTE: 初期データ予約済みの1時間分を引く必要がある
const NumHours = (24 * 365) - 1

// この時間[h]を超えた配信枠は長時間配信とみなす
const LongHourThreshold = 10
