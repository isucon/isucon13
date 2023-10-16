package main

import (
	"context"
	"database/sql"
	"errors"
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
	Rank     int `json:"tip_rank"`
	TotalTip int `json:"total_tip"`
}

type TipRankModel struct {
	Rank     int `db:"tip_rank"`
	TotalTip int `db:"total_tip"`
}

type ReactionRank struct {
	Rank          int    `json:"reaction_rank"`
	TotalReaction int    `json:"total_reaction"`
	EmojiName     string `json:"emoji_name"`
}

type ReactionRankModel struct {
	Rank          int    `db:"reaction_rank"`
	TotalReaction int    `db:"total_reaction"`
	EmojiName     string `db:"emoji_name"`
}

func getUserStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	username := c.Param("username")

	userModel := UserModel{}
	err := dbConn.GetContext(ctx, &userModel, "SELECT * FROM users where name = ?", username)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var viewedLivestreams []*LivestreamViewerModel
	if err := dbConn.SelectContext(ctx, &viewedLivestreams, "SELECT user_id, livestream_id FROM livestream_viewers_history WHERE user_id = ?", userModel.Id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tipRankModelPerLivestreams, err := queryTotalTipRankPerViewedLivestream(ctx, viewedLivestreams)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tipRankPerLivestream := make(map[int]TipRank)
	for livestreamId, tipRankModel := range tipRankModelPerLivestreams {
		tipRankPerLivestream[livestreamId] = TipRank{
			Rank:     tipRankModel.Rank,
			TotalTip: tipRankModel.TotalTip,
		}
	}

	stats := UserStatistics{
		TipRankPerLivestreams: tipRankPerLivestream,
	}
	return c.JSON(http.StatusOK, stats)
}

func getLivestreamStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamId := c.Param("livestream_id")

	tipRankModels := []TipRankModel{}
	query := "SELECT SUM(tip) AS total_tip, RANK() OVER(ORDER BY SUM(tip) DESC) AS tip_rank " +
		"FROM livecomments GROUP BY livestream_id " +
		"HAVING livestream_id = ? ORDER BY total_tip DESC LIMIT 3"
	if err := dbConn.SelectContext(ctx, &tipRankModels, query, livestreamId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reactionRankModels := []ReactionRankModel{}
	query = "SELECT COUNT(*) AS total_reaction, emoji_name, RANK() OVER(ORDER BY COUNT(*) DESC) AS reaction_rank " +
		"FROM reactions GROUP BY livestream_id, emoji_name " +
		"HAVING livestream_id = ? ORDER BY total_reaction DESC LIMIT 3"
	if err := dbConn.SelectContext(ctx, &reactionRankModels, query, livestreamId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tipRanks := make([]TipRank, len(tipRankModels))
	for i := range tipRankModels {
		tipRanks[i] = TipRank{
			Rank:     tipRankModels[i].Rank,
			TotalTip: tipRankModels[i].TotalTip,
		}
	}

	reactionRanks := make([]ReactionRank, len(reactionRankModels))
	for i := range reactionRankModels {
		reactionRanks[i] = ReactionRank{
			Rank:          reactionRankModels[i].Rank,
			TotalReaction: reactionRankModels[i].TotalReaction,
		}
	}

	stats := LivestreamStatistics{
		MostTipRanking:            tipRanks,
		MostPostedReactionRanking: reactionRanks,
	}
	return c.JSON(http.StatusOK, stats)
}

func queryTotalTipRankPerViewedLivestream(
	ctx context.Context,
	viewedLivestreams []*LivestreamViewerModel,
) (map[int]TipRankModel, error) {
	totalTipRankPerLivestream := make(map[int]TipRankModel)
	// get total tip per viewed livestream
	for _, viewedLivestream := range viewedLivestreams {
		totalTip := 0
		if err := dbConn.GetContext(ctx, &totalTip, "SELECT SUM(tip) FROM livecomments WHERE user_id = ? AND livestream_id = ?", viewedLivestream.UserId, viewedLivestream.LivestreamId); err != nil {
			return totalTipRankPerLivestream, err
		}

		totalTipRankPerLivestream[viewedLivestream.LivestreamId] = TipRankModel{
			TotalTip: totalTip,
		}
	}

	for livestreamId, stat := range totalTipRankPerLivestream {
		query := "SELECT SUM(tip) AS total_tip, RANK() OVER(ORDER BY SUM(tip) DESC) AS tip_rank " +
			"FROM livecomments GROUP BY livestream_id " +
			"HAVING livestream_id = ? AND total_tip = ?"

		rank := TipRankModel{}
		if err := dbConn.GetContext(ctx, &rank, query, livestreamId, stat.TotalTip); err != nil {
			return totalTipRankPerLivestream, err
		}

		totalTipRankPerLivestream[livestreamId] = rank

	}

	return totalTipRankPerLivestream, nil
}
