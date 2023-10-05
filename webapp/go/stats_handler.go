package main

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
)

// FIXME: 配信毎、ユーザごとのリアクション種別ごとの数などもだす

type LivestreamStatistics struct {
	MostTipRanking            []TipRank      `json:"most_tip_ranking"`
	MostPostedReactionRanking []ReactionRank `json:"most_posted_reaction_ranking"`
}

type UserStatistics struct {
	TipRankPerLivestreams map[int]TipRank `json:"tip_rank_by_livestream"`
}

type TipRank struct {
	Rank     int `json:"tip_rank" db:"tip_rank"`
	TotalTip int `json:"total_tip" db:"total_tip"`
}

type ReactionRank struct {
	Rank          int    `json:"reaction_rank" db:"reaction_rank"`
	TotalReaction int    `json:"total_reaction" db:"total_reaction"`
	EmojiName     string `json:"emoji_name" db:"emoji_name"`
}

func getUserStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	userId := c.Param("user_id")

	var viewedLivestreams []*LivestreamViewer
	if err := dbConn.SelectContext(ctx, &viewedLivestreams, "SELECT user_id, livestream_id FROM livestream_viewers_history WHERE user_id = ?", userId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tipRankPerLivestreams, err := queryTotalTipRankPerViewedLivestream(ctx, userId, viewedLivestreams)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	stats := UserStatistics{
		TipRankPerLivestreams: tipRankPerLivestreams,
	}
	return c.JSON(http.StatusOK, stats)
}

func getLivestreamStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamId := c.Param("livestream_id")

	tipRanks := []TipRank{}
	query := "SELECT SUM(tip) AS total_tip, RANK() OVER(ORDER BY SUM(tip) DESC) AS tip_rank " +
		"FROM livecomments GROUP BY livestream_id " +
		"HAVING livestream_id = ? ORDER BY total_tip DESC LIMIT 3"
	if err := dbConn.SelectContext(ctx, &tipRanks, query, livestreamId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reactionRanks := []ReactionRank{}
	query = "SELECT COUNT(*) AS total_reaction, emoji_name, RANK() OVER(ORDER BY COUNT(*) DESC) AS reaction_rank " +
		"FROM reactions GROUP BY livestream_id, emoji_name " +
		"HAVING livestream_id = ? ORDER BY total_reaction DESC LIMIT 3"
	if err := dbConn.SelectContext(ctx, &reactionRanks, query, livestreamId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	stats := LivestreamStatistics{
		MostTipRanking:            tipRanks,
		MostPostedReactionRanking: reactionRanks,
	}
	return c.JSON(http.StatusOK, stats)
}

func queryTotalTipRankPerViewedLivestream(
	ctx context.Context,
	userId string,
	viewedLivestreams []*LivestreamViewer,
) (map[int]TipRank, error) {
	totalTipRankPerLivestream := make(map[int]TipRank)
	// get total tip per viewed livestream
	for _, viewedLivestream := range viewedLivestreams {
		totalTip := 0
		if err := dbConn.GetContext(ctx, &totalTip, "SELECT SUM(tip) FROM livecomments WHERE user_id = ? AND livestream_id = ?", viewedLivestream.UserId, viewedLivestream.LivestreamId); err != nil {
			return totalTipRankPerLivestream, err
		}

		totalTipRankPerLivestream[viewedLivestream.LivestreamId] = TipRank{
			TotalTip: totalTip,
		}
	}

	for livestreamId, stat := range totalTipRankPerLivestream {
		query := "SELECT SUM(tip) AS total_tip, RANK() OVER(ORDER BY SUM(tip) DESC) AS tip_rank " +
			"FROM livecomments GROUP BY livestream_id " +
			"HAVING livestream_id = ? AND total_tip = ?"

		rank := TipRank{}
		if err := dbConn.GetContext(ctx, &rank, query, livestreamId, stat.TotalTip); err != nil {
			return totalTipRankPerLivestream, err
		}

		totalTipRankPerLivestream[livestreamId] = rank

	}

	return totalTipRankPerLivestream, nil
}
