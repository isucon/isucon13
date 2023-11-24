package scheduler

import (
	"fmt"
	"log"
	"slices"
	"sort"
	"sync"
	"time"
)

var StatsSched = NewStatsScheduler()

func init() {
	startAt := time.Now()
	log.Printf("Start loading initial data ... %s\n", startAt.String())
	if err := StatsSched.loadInitialData(); err != nil {
		log.Fatalln(err)
	}
	elapsed := time.Since(startAt)
	endAt := time.Now()
	log.Printf("Finish loading initial data: %s\n", endAt.String())
	log.Printf("elapsed = %s\n", elapsed.String())
}

// NOTE: Pretestの用途で利用を想定

type UserStatsRanking []*UserStats

func (r UserStatsRanking) Len() int      { return len(r) }
func (r UserStatsRanking) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r UserStatsRanking) Less(i, j int) bool {
	var (
		leftScore  = r[i].Score()
		rightScore = r[j].Score()
	)
	if leftScore == rightScore {
		return r[i].Username < r[j].Username
	} else {
		return leftScore < rightScore
	}
}

type UserStats struct {
	Username string

	// 視聴者数
	TotalViewers int64
	// トータルリアクション数
	// お気に入り絵文字
	reactions map[string]int64

	// トータルライブコメント数
	TotalLivecomments int64
	// チップ合計金額
	TotalTips int64
}

func NewUserStats(username string) *UserStats {
	return &UserStats{
		Username:  username,
		reactions: make(map[string]int64),
	}
}

func (s *UserStats) TotalReactions() int64 {
	var total int64
	for _, count := range s.reactions {
		total += count
	}
	return total
}

func (s *UserStats) FavoriteEmoji() (string, bool) {
	var (
		favoriteEmojis     []string
		favoriteEmojiCount int64
	)
	for emoji, count := range s.reactions {
		if count > favoriteEmojiCount {
			favoriteEmojis = []string{emoji}
			favoriteEmojiCount = count
		} else if count == favoriteEmojiCount {
			favoriteEmojis = append(favoriteEmojis, emoji)
		}
	}
	if len(favoriteEmojis) == 0 {
		return "", false
	}

	// FIXME: 文字列アルファベット順の昇順になるはずだが、要チェック
	slices.Sort(favoriteEmojis)
	return favoriteEmojis[len(favoriteEmojis)-1], true
}

func (s *UserStats) Score() int64 {
	return int64(len(s.reactions)) + s.TotalTips
}

type LivestreamStatsRanking []*LivestreamStats

func (r LivestreamStatsRanking) Len() int      { return len(r) }
func (r LivestreamStatsRanking) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r LivestreamStatsRanking) Less(i, j int) bool {
	var (
		leftScore  = r[i].Score()
		rightScore = r[j].Score()
	)
	if leftScore == rightScore {
		return r[i].LivestreamID < r[j].LivestreamID
	} else {
		return leftScore < rightScore
	}
}

type LivestreamStats struct {
	LivestreamID int64

	// 視聴者数 (初期では0)
	totalViewers int64

	// トータルレポート数 (初期では0)
	totalReports int64

	// トータルリアクション数
	totalReactions int64

	// 最大チップ金額 (ライブコメント)
	totalTips int64
	maxTip    int64
}

func NewLivestreamStats(livestreamID int64) *LivestreamStats {
	return &LivestreamStats{LivestreamID: livestreamID}
}

func (s *LivestreamStats) Score() int64 {
	return s.totalReactions + s.totalTips
}

type StatsScheduler struct {
	userStatsMu       sync.Mutex
	userStats         map[string]*UserStats
	livestreamStatsMu sync.Mutex
	livestreamStats   map[int64]*LivestreamStats
}

func NewStatsScheduler() *StatsScheduler {
	return &StatsScheduler{
		userStats:       make(map[string]*UserStats),
		livestreamStats: make(map[int64]*LivestreamStats),
	}
}

func (s *StatsScheduler) loadInitialData() error {
	// 配信者初期化
	for _, user := range initialUserPool {
		s.userStats[user.Name] = NewUserStats(user.Name)
	}
	// ライブ配信初期化
	var i int64 = 1
	for ; i <= int64(len(initialReservationPool)); i++ {
		s.livestreamStats[i] = new(LivestreamStats)
	}

	// リアクション追加
	for _, reaction := range initialReactionPool {
		userIdx := reaction.UserID - 1
		user := initialUserPool[userIdx]
		livestreamID := reaction.LivestreamID

		if err := s.addReactionForUser(user.Name, livestreamID, reaction.EmojiName); err != nil {
			return err
		}
		if err := s.addReactionForLivestream(user.Name, livestreamID, reaction.EmojiName); err != nil {
			return err
		}
	}
	// ライブコメント追加
	for _, livecomment := range initialLivecommentPool {
		userIdx := livecomment.UserID - 1
		user := initialUserPool[userIdx]
		livestreamID := livecomment.LivestreamID

		if err := s.addLivecommentForUser(user.Name, livestreamID, &Tip{}); err != nil {
			return err
		}
		if err := s.addLivecommentForLivestream(user.Name, livestreamID, &Tip{}); err != nil {
			return err
		}
	}

	return nil
}

func (s *StatsScheduler) GetUserStats(username string) (*UserStats, error) {
	s.userStatsMu.Lock()
	defer s.userStatsMu.Unlock()

	userStats, ok := s.userStats[username]
	if !ok {
		return nil, fmt.Errorf("存在しないユーザ名の統計情報取得似失敗しました: %s", username)
	}

	return userStats, nil
}

func (s *StatsScheduler) GetLivestreamStats(livestreamID int64) (*LivestreamStats, error) {
	s.livestreamStatsMu.Lock()
	defer s.livestreamStatsMu.Unlock()

	livestreamStats, ok := s.livestreamStats[livestreamID]
	if !ok {
		return nil, fmt.Errorf("存在しないライブ配信の統計情報取得似失敗しました: %d", livestreamID)
	}

	return livestreamStats, nil
}

func (s *StatsScheduler) GetUserRank(username string) (int64, error) {
	s.userStatsMu.Lock()
	defer s.userStatsMu.Unlock()

	stats := make(UserStatsRanking, len(s.userStats))
	var idx int
	for _, stat := range s.userStats {
		stats[idx] = stat
		idx++
	}
	sort.Sort(stats)

	var rank int64 = 1
	for _, stat := range stats {
		if stat.Username == username {
			return rank, nil
		}
		rank++
	}

	return 0, fmt.Errorf("存在しないユーザが指定されました: %s", username)
}

func (s *StatsScheduler) GetLivestreamRank(livestreamID int64) (int64, error) {
	s.livestreamStatsMu.Lock()
	defer s.livestreamStatsMu.Unlock()

	stats := make(LivestreamStatsRanking, len(s.userStats))
	var idx int
	for _, stat := range s.livestreamStats {
		stats[idx] = stat
		idx++
	}
	sort.Sort(stats)

	var rank int64 = 1
	for _, stat := range stats {
		if stat.LivestreamID == livestreamID {
			return rank, nil
		}
		rank++
	}

	return 0, fmt.Errorf("存在しないライブ配信が指定されました: %d", livestreamID)
}

// 視聴開始/終了
// ユーザ単位の視聴者数、ライブ配信単位の視聴者数を更新する必要がある
func (s *StatsScheduler) EnterLivestream(streamerName string, livestreamID int64) error {
	// ユーザ
	s.userStatsMu.Lock()
	userStats, ok := s.userStats[streamerName]
	if !ok {
		s.userStatsMu.Unlock()
		return fmt.Errorf("統計情報の更新に失敗(EnterLivestream.userStats): user=%s, livestream=%d", streamerName, livestreamID)
	}
	userStats.TotalViewers++
	s.userStatsMu.Unlock()

	// ライブ配信
	s.livestreamStatsMu.Lock()
	livestreamStats, ok := s.livestreamStats[livestreamID]
	if !ok {
		s.livestreamStatsMu.Unlock()
		return fmt.Errorf("統計情報の更新に失敗(EnterLivestream.livestreamStats): user=%s, livestream=%d", streamerName, livestreamID)
	}
	livestreamStats.totalViewers++
	s.livestreamStatsMu.Unlock()

	return nil
}
func (s *StatsScheduler) ExitLivestream(streamerName string, livestreamID int64) error {
	// ユーザ
	s.userStatsMu.Lock()
	defer s.userStatsMu.Unlock()
	userStats, ok := s.userStats[streamerName]
	if !ok {
		return fmt.Errorf("統計情報の更新に失敗(ExitLivestream.userStats): user=%s, livestream=%d", streamerName, livestreamID)
	}
	if userStats.TotalViewers <= 0 {
		return fmt.Errorf("ExitLivestreamの呼び出し回数が不正です: streamer=%s viewers_count=%d", streamerName, userStats.TotalViewers)
	}
	userStats.TotalViewers--

	// ライブ配信
	s.livestreamStatsMu.Lock()
	defer s.livestreamStatsMu.Unlock()
	livestreamStats, ok := s.livestreamStats[livestreamID]
	if !ok {
		return fmt.Errorf("統計情報の更新に失敗(ExitLivestream.livestreamStats): user=%s, livestream=%d", streamerName, livestreamID)
	}
	if livestreamStats.totalViewers <= 0 {
		return fmt.Errorf("ExitLivestreamの呼び出し回数が不正です: streamer=%s viewers_count=%d", streamerName, livestreamStats.totalViewers)
	}
	livestreamStats.totalViewers--

	return nil
}

// リアクション追加 (ユーザの配信に対して)
func (s *StatsScheduler) addReactionForUser(streamerName string, livestreamID int64, reaction string) error {
	userStats, ok := s.userStats[streamerName]
	if !ok {
		return fmt.Errorf("統計情報の更新に失敗(AddReaction.userStats): user=%s, livestream=%d", streamerName, livestreamID)
	}
	userStats.reactions[reaction]++

	return nil
}
func (s *StatsScheduler) addReactionForLivestream(streamerName string, livestreamID int64, reaction string) error {
	livestreamStats, ok := s.livestreamStats[livestreamID]
	if !ok {
		return fmt.Errorf("統計情報の更新に失敗(AddReaction.livestreamStats): user=%s, livestream=%d", streamerName, livestreamID)
	}
	livestreamStats.totalReactions++

	return nil
}
func (s *StatsScheduler) AddReaction(streamerName string, livestreamID int64, reaction string) error {
	s.userStatsMu.Lock()
	defer s.userStatsMu.Unlock()
	if err := s.addReactionForUser(streamerName, livestreamID, reaction); err != nil {
		return err
	}
	s.livestreamStatsMu.Lock()
	defer s.livestreamStatsMu.Unlock()
	if err := s.addReactionForLivestream(streamerName, livestreamID, reaction); err != nil {
		return err
	}
	return nil
}

// スパム報告追加 (ユーザの配信に対して)
func (s *StatsScheduler) AddReport(streamerName string, livestreamID int64) error {
	// ライブ配信
	s.livestreamStatsMu.Lock()
	defer s.livestreamStatsMu.Unlock()
	livestreamStats, ok := s.livestreamStats[livestreamID]
	if !ok {
		s.livestreamStatsMu.Unlock()
		return fmt.Errorf("統計情報の更新に失敗(AddReaction.livestreamStats): user=%s, livestream=%d", streamerName, livestreamID)
	}
	livestreamStats.totalReports++

	return nil
}

// ライブコメント追加 (ユーザの配信に対して)
func (s *StatsScheduler) addLivecommentForUser(streamerName string, livestreamID int64, tip *Tip) error {
	userStats, ok := s.userStats[streamerName]
	if !ok {
		return fmt.Errorf("統計情報の更新に失敗(AddLivecomment.userStats): user=%s, livestream=%d", streamerName, livestreamID)
	}
	userStats.TotalLivecomments++
	return nil
}
func (s *StatsScheduler) addLivecommentForLivestream(streamerName string, livestreamID int64, tip *Tip) error {
	livestreamStats, ok := s.livestreamStats[livestreamID]
	if !ok {
		return fmt.Errorf("統計情報の更新に失敗(AddLivecomment.livestreamStats): user=%s, livestream=%d", streamerName, livestreamID)
	}
	livestreamStats.totalTips += int64(tip.Tip)
	livestreamStats.maxTip = max(livestreamStats.maxTip, int64(tip.Tip))
	return nil
}
func (s *StatsScheduler) AddLivecomment(streamerName string, livestreamID int64, tip *Tip) error {
	s.userStatsMu.Lock()
	defer s.userStatsMu.Unlock()
	if err := s.addLivecommentForUser(streamerName, livestreamID, tip); err != nil {
		return err
	}
	s.livestreamStatsMu.Lock()
	defer s.livestreamStatsMu.Unlock()
	if err := s.addLivecommentForLivestream(streamerName, livestreamID, tip); err != nil {
		return err
	}
	return nil
}
