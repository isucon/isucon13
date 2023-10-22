package benchstats

// FIXME: スケジューラから払い出すTipを当該ネガティブスコアに基づいて減額

// ネガティブ値が大きくなるに連れ、サイコロを毎ユーザ行動毎に振らされ、悪い目が出たら何もしない動きをされるようになる
// 減点はしないが、得点源を失う

// // 特定のユーザがトラブルメーカーとして振る舞うべきか判定する
// func (u *userScheduler) BehaveTroubleMaker(viewer *User) bool {
// 	u.negativeCountsMu.RLock()
// 	defer u.negativeCountsMu.RUnlock()

// 	const maxNegativeCount = 100

// 	if viewer.UserId <= 0 || viewer.UserId >= len(u.negativeCounts) {
// 		return false
// 	}
// 	negativeCount := u.negativeCounts[viewer.UserId]

// 	// 100程度のリクエスト失敗以降は同等に扱う
// 	// 0 ~ 10の値を取るようになるので、負数を除いて2割程度は最低限正常な振る舞いをするように残しておく
// 	negativeCount = int(math.Min(float64(negativeCount), maxNegativeCount))
// 	negativeValue := math.Sqrt(float64(negativeCount))

// 	r := rand.Intn(int(math.Sqrt(maxNegativeCount)))
// 	return r >= int(math.Max(negativeValue-2, 0))
// }
