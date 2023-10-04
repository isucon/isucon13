package main

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
)

// FIXME: 配信毎、ユーザごとのリアクション種別ごとの数などもだす

type LivestreamStatistics struct {
	MostTipLivecommentRanking []Livecomment
}

type UserStatistics struct {
	ViewedLivestreamStatistics map[int]ViewedLivestreamStatistic `json:"viewed_livestream_statistics"`
}

type ViewedLivestreamStatistic struct {
	TipRank  int `json:"tip_rank"`
	TotalTip int `json:"total_tip"`
}

func getUserStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	userId := c.Param("user_id")

	var viewedLivestreams []*LivestreamViewer
	if err := dbConn.SelectContext(ctx, &viewedLivestreams, "SELECT * FROM livestream_viewers_history WHERE user_id = ?", userId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	viewedLivestreamStatistics, err := queryTotalTipRankPerLivestream(ctx, userId, viewedLivestreams)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	stats := UserStatistics{
		ViewedLivestreamStatistics: viewedLivestreamStatistics,
	}
	return c.JSON(http.StatusOK, stats)
}

func getLivestreamStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamId := c.Param("livestream_id")

	livestream := Livestream{}
	if err := dbConn.GetContext(ctx, &livestream, "SELECT user_id, start_at FROM livestreams WHERE id = ?", livestreamId); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "livestream not found")
	}

	return c.JSON(http.StatusOK, nil)
}

func queryTotalTipRankPerLivestream(
	ctx context.Context,
	userId string,
	viewedLivestreams []*LivestreamViewer,
) (map[int]ViewedLivestreamStatistic, error) {
	totalTipRankPerLivestream := make(map[int]ViewedLivestreamStatistic)
	// get total tip per viewed livestream
	for _, viewedLivestream := range viewedLivestreams {
		totalTip := 0
		if err := dbConn.SelectContext(ctx, &totalTip, "SELECT SUM(tip) FROM livecomments WHERE user_id = ? AND livestream_id = ?", viewedLivestream.UserId, viewedLivestream.LivestreamId); err != nil {
			return totalTipRankPerLivestream, err
		}

		totalTipRankPerLivestream[viewedLivestream.LivestreamId] = ViewedLivestreamStatistic{
			TotalTip: totalTip,
		}
	}

	for livestreamId, stat := range totalTipRankPerLivestream {
		type tipRank struct {
			rank     int `db:"tip_rank"`
			totalTip int `db:"total_tip"`
		}
		query := "SELECT SUM(tip) AS total_tip, RANK() OVER(ORDER BY SUM(tip) DESC) AS tip_rank" +
			"FROM livecomments GROUP BY livestream_id " +
			"HAVING livestream_id = ? AND total_tip = ?"

		var rank *tipRank
		if err := dbConn.SelectContext(ctx, &rank, query, livestreamId, stat.TotalTip); err != nil {
			return totalTipRankPerLivestream, err
		}

		totalTipRankPerLivestream[livestreamId] = ViewedLivestreamStatistic{
			TotalTip: rank.totalTip,
			TipRank:  rank.rank,
		}

	}

	return totalTipRankPerLivestream, nil
}
