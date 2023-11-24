package scheduler

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnterExitStats(t *testing.T) {
	s := NewStatsScheduler()
	for i := 0; i < 5; i++ {
		livestreamID := int64(i)
		s.livestreamStats[livestreamID] = NewLivestreamStats(livestreamID)
		username := fmt.Sprintf("streamer%d", i)
		s.userStats[username] = NewUserStats(username)
	}

	for i := 0; i < 5; i++ {
		streamerName := fmt.Sprintf("streamer%d", i)
		for j := 0; j < 5; j++ {
			livestreamID := int64(j)
			err := s.EnterLivestream(streamerName, livestreamID)
			assert.NoError(t, err)
		}
	}
	for i := 0; i < 5; i++ {
		streamerName := fmt.Sprintf("streamer%d", i)
		for j := 0; j < 5; j++ {
			livestreamID := int64(j)
			assert.Equal(t, int64(5), s.userStats[streamerName].TotalViewers)
			assert.Equal(t, int64(5), s.livestreamStats[livestreamID].TotalViewers)
		}
	}
	for i := 0; i < 5; i++ {
		streamerName := fmt.Sprintf("streamer%d", i)
		for j := 0; j < 5; j++ {
			livestreamID := int64(j)
			err := s.ExitLivestream(streamerName, livestreamID)
			assert.NoError(t, err)
		}
	}
	for i := 0; i < 5; i++ {
		streamerName := fmt.Sprintf("streamer%d", i)
		for j := 0; j < 5; j++ {
			livestreamID := int64(j)
			assert.Equal(t, int64(0), s.userStats[streamerName].TotalViewers)
			assert.Equal(t, int64(0), s.livestreamStats[livestreamID].TotalViewers)
		}
	}
}

func TestReactionStats(t *testing.T) {
	s := NewStatsScheduler()
	for i := 0; i < 5; i++ {
		livestreamID := int64(i)
		s.livestreamStats[livestreamID] = NewLivestreamStats(livestreamID)
		username := fmt.Sprintf("streamer%d", i)
		s.userStats[username] = NewUserStats(username)
	}

	for i := 0; i < 5; i++ {
		streamerName := fmt.Sprintf("streamer%d", i)
		for j := 0; j < 5; j++ {
			livestreamID := int64(j)
			reaction := strconv.Itoa(j)
			err := s.AddReaction(streamerName, livestreamID, reaction)
			assert.NoError(t, err)
		}
	}
	for i := 0; i < 5; i++ {
		streamerName := fmt.Sprintf("streamer%d", i)
		for j := 0; j < 5; j++ {
			livestreamID := int64(j)
			assert.Equal(t, int64(5), s.userStats[streamerName].TotalReactions())
			assert.Equal(t, int64(5), s.livestreamStats[livestreamID].TotalReactions)
			favoriteEmoji, ok := s.userStats[streamerName].FavoriteEmoji()
			assert.True(t, ok)
			assert.Equal(t, "4", favoriteEmoji)
			assert.Equal(t, int64(5), s.userStats[streamerName].Score())
		}
	}

	startAt := time.Now()
	log.Printf("start = %s\n", startAt.String())
	userRank, err := s.GetUserRank("streamer1")
	assert.NoError(t, err)
	assert.Equal(t, int64(4), userRank)

	livestreamRank, err := s.GetLivestreamRank(3)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), livestreamRank)
	endAt := time.Now()
	log.Printf("end = %s\n", endAt.String())
	log.Printf("elapsed = %s\n", time.Since(startAt).String())
}

func TestLivecommentStats(t *testing.T) {
	s := NewStatsScheduler()
	for i := 0; i < 5; i++ {
		livestreamID := int64(i)
		s.livestreamStats[livestreamID] = NewLivestreamStats(livestreamID)
		username := fmt.Sprintf("streamer%d", i)
		s.userStats[username] = NewUserStats(username)
	}

	for i := 0; i < 5; i++ {
		streamerName := fmt.Sprintf("streamer%d", i)
		for j := 0; j < 5; j++ {
			livestreamID := int64(j)
			err := s.AddLivecomment(streamerName, livestreamID, &Tip{Tip: j})
			assert.NoError(t, err)
		}
	}
	for i := 0; i < 5; i++ {
		streamerName := fmt.Sprintf("streamer%d", i)
		for j := 0; j < 5; j++ {
			livestreamID := int64(j)
			assert.Equal(t, int64(5), s.userStats[streamerName].TotalLivecomments)
			assert.Equal(t, int64(j*5), s.livestreamStats[livestreamID].TotalTips)
			assert.Equal(t, int64(j), s.livestreamStats[livestreamID].MaxTip)
		}
	}

	startAt := time.Now()
	userRank, err := s.GetUserRank("streamer4")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), userRank)

	livestreamRank, err := s.GetLivestreamRank(1)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), livestreamRank)
	endAt := time.Now()
	log.Printf("end = %s\n", endAt.String())
	log.Printf("elapsed = %s\n", time.Since(startAt).String())
}
