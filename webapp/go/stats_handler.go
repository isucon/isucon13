package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// FIXME: 配信毎、ユーザごとのリアクション種別ごとの数などもだす

type LivestreamStatistics struct {
	TotalViewers                            int `json:"total_viewers"`
	TotalTips                               int `json:"total_tips"`
	TotalLivecomments                       int `json:"total_livecomments"`
	TotalReactions                          int `json:"total_reactions"`
	PreviousLivestreamTotalViewersDiff      int `json:"previous_livestream_total_viewers_diff"`
	PreviousLivestreamTotalTipsDiff         int `json:"previous_livestream_total_tips_diff"`
	PreviousLivestreamTotalLivecommentsDiff int `json:"previous_livestream_total_livecomments_diff"`
	PreviousLivestreamTotaRlReactionsDiff   int `json:"previous_livestream_total_reactions_diff"`
}

type UserStatistics struct {
	AverageViewers      float64 `json:"average_viewers"`
	AverageTips         float64 `json:"average_tips"`
	AverageLivecomments float64 `json:"average_livecomments"`
	AverageReactions    float64 `json:"average_reactions"`
}

func getUserStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	userID := c.Param("user_id")

	rows, err := dbConn.QueryxContext(ctx, "SELECT id FROM livestreams where user_id = ?", userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	statsSequence := []LivestreamStatistics{}

	for rows.Next() {
		ls := Livestream{}
		if err := rows.Scan(&ls); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		stats, err := queryLivestreamStatistics(ctx, fmt.Sprintf("%d", ls.Id))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		statsSequence = append(statsSequence, stats)
	}

	viewersSum := float64(0)
	reactionsSum := float64(0)
	livecommentsSum := float64(0)
	tipsSum := float64(0)

	for _, stats := range statsSequence {
		viewersSum += float64(stats.TotalViewers)
		reactionsSum += float64(stats.TotalReactions)
		livecommentsSum += float64(stats.TotalLivecomments)
		tipsSum += float64(stats.TotalTips)
	}

	stats := UserStatistics{
		AverageViewers:      viewersSum / float64(len(statsSequence)),
		AverageReactions:    reactionsSum / float64(len(statsSequence)),
		AverageLivecomments: livecommentsSum / float64(len(statsSequence)),
		AverageTips:         tipsSum / float64(len(statsSequence)),
	}
	return c.JSON(http.StatusOK, stats)
}

func getLivestreamStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	livestreamID := c.Param("livestream_id")

	livestream := Livestream{}
	if err := dbConn.GetContext(ctx, &livestream, "SELECT user_id, start_at FROM livestreams WHERE id = ?", livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "livestream not found")
	}

	statistics, err := queryLivestreamStatistics(ctx, livestreamID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	prevLivestream := getPreviousLivestream(ctx, &livestream)
	prevLivestreamStatistics := LivestreamStatistics{}
	if prevLivestream != nil {
		prevLivestreamID := fmt.Sprintf("%d", prevLivestream.Id)
		prevLivestreamStatistics, err = queryLivestreamStatistics(ctx, prevLivestreamID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	statistics.PreviousLivestreamTotalViewersDiff = statistics.TotalViewers - prevLivestreamStatistics.TotalViewers
	statistics.PreviousLivestreamTotalLivecommentsDiff = statistics.TotalLivecomments - prevLivestreamStatistics.TotalLivecomments
	statistics.PreviousLivestreamTotalTipsDiff = statistics.TotalTips - prevLivestreamStatistics.TotalTips
	statistics.PreviousLivestreamTotaRlReactionsDiff = statistics.TotalReactions - prevLivestreamStatistics.TotalReactions

	return c.JSON(http.StatusOK, statistics)
}

func countTotalViewers(ctx context.Context, livestreamID string) (int, error) {
	rows, err := dbConn.QueryxContext(ctx, "SELECT * FROM livestream_viewers WHERE livestream_id = ?", livestreamID)
	if err != nil {
		return 0, err
	}

	viewerCount := 0
	for rows.Next() {
		viewerCount++
	}

	return viewerCount, nil
}

func countTotalReactions(ctx context.Context, livestreamID string) (int, error) {
	rows, err := dbConn.QueryxContext(ctx, "SELECT * FROM reactions WHERE livestream_id = ?", livestreamID)
	if err != nil {
		return 0, err
	}

	reactionCount := 0
	for rows.Next() {
		reactionCount++
	}

	return reactionCount, nil
}

func calculateLivecommentStatistics(ctx context.Context, livestreamID string) (totalLivecomments int, totalTips int, err error) {
	rows, err := dbConn.QueryxContext(ctx, "SELECT * FROM livecomments WHERE livestream_id = ?", livestreamID)
	if err != nil {
		return 0, 0, nil
	}

	totalLivecomments = 0
	totalTips = 0

	for rows.Next() {
		livecomment := Livecomment{}
		if err := rows.Scan(&livecomment); err != nil {
			return 0, 0, err
		}

		totalLivecomments++
		totalTips += livecomment.Tip
	}

	return totalLivecomments, totalTips, nil
}

func getPreviousLivestream(ctx context.Context, currentLivestream *Livestream) *Livestream {
	rows, err := dbConn.QueryxContext(ctx, "SELECT id, start_at FROM livestreams WHERE user_id = ?", currentLivestream.UserId)
	if err != nil {
		return nil
	}

	newestLivestream := &Livestream{}
	newestLivestreamStartAt := int64(0)
	for rows.Next() {
		ls := Livestream{}
		if err := rows.Scan(&ls); err != nil {
			return nil
		}

		if newestLivestreamStartAt < ls.StartAt.Unix() && ls.StartAt.Unix() < currentLivestream.StartAt.Unix() {
			*newestLivestream = ls
		}
	}

	if newestLivestreamStartAt == int64(0) {
		return nil
	}

	return newestLivestream
}

func queryLivestreamStatistics(ctx context.Context, livestreamID string) (LivestreamStatistics, error) {
	statistics := LivestreamStatistics{}

	totalLivecomments, totalTips, err := calculateLivecommentStatistics(ctx, livestreamID)
	if err != nil {
		return statistics, err
	}
	statistics.TotalLivecomments = totalLivecomments
	statistics.TotalTips = totalTips

	totalViewers, err := countTotalViewers(ctx, livestreamID)
	if err != nil {
		return statistics, err
	}
	statistics.TotalViewers = totalViewers

	totalReactions, err := countTotalReactions(ctx, livestreamID)
	if err != nil {
		return statistics, err
	}

	statistics.TotalReactions = totalReactions
	return statistics, nil
}
