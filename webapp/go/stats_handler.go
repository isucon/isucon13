package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type LivestreamStatistics struct {
	TotalViewers                         int `json:"total_viewers"`
	TotalTips                            int `json:"total_tips"`
	TotalSuperchats                      int `json:"total_superchats"`
	TotalReactions                       int `json:"total_reactions"`
	PreviousLivesteamTotalViewersDiff    int `json:"previous_livesteam_total_viewers_diff"`
	PreviousLivesteamTotalTipsDiff       int `json:"previous_livesteam_total_tips_diff"`
	PreviousLivesteamTotalSuperchatsDiff int `json:"previous_livesteam_total_superchats_diff"`
}

func getLivestreamStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	livestreamID := c.Param("livestream_id")

	livestream := Livestream{}
	if err := dbConn.GetContext(ctx, &livestream, "SELECT user_id, start_at FROM livestreams WHERE id = ?", livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "livestream not found")
	}

	statistics := LivestreamStatistics{}
	if err := calculateSuperchatStatistics(ctx, livestreamID, &statistics); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	totalViewers, err := countTotalViewers(ctx, livestreamID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	statistics.TotalViewers = totalViewers

	totalReactions, err := countTotalReactions(ctx, livestreamID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	statistics.TotalReactions = totalReactions

	prevLivestream := getPreviousLivestream(ctx, &livestream)
	prevLivestreamStatistics := LivestreamStatistics{}
	if prevLivestream != nil {
		prevLivestreamID := fmt.Sprintf("%d", prevLivestream.Id)
		if err := calculateSuperchatStatistics(ctx, prevLivestreamID, &prevLivestreamStatistics); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		totalViewers, err := countTotalViewers(ctx, prevLivestreamID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		prevLivestreamStatistics.TotalViewers = totalViewers

		totalReactions, err := countTotalReactions(ctx, prevLivestreamID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		statistics.TotalReactions = totalReactions
	}

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

func calculateSuperchatStatistics(ctx context.Context, livestreamID string, stats *LivestreamStatistics) error {
	rows, err := dbConn.QueryxContext(ctx, "SELECT * FROM superchats WHERE livestream_id = ?", livestreamID)
	if err != nil {
		return err
	}

	for rows.Next() {
		superchat := Superchat{}
		if err := rows.Scan(&superchat); err != nil {
			return err
		}

		stats.TotalSuperchats++
		stats.TotalTips += superchat.Tip
	}

	return nil
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
