package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// FIXME: 配信毎、ユーザごとのリアクション種別ごとの数などもだす

type LivestreamStatistics struct {
	MostTipRanking            []TipRank      `json:"most_tip_ranking"`
	MostPostedReactionRanking []ReactionRank `json:"most_posted_reaction_ranking"`
}

type UserStatistics struct {
	TipRankPerLivestreams map[int64]TipRank `json:"tip_rank_by_livestream"`
}

type TipRank struct {
	Rank     int64 `json:"tip_rank"`
	TotalTip int64 `json:"total_tip"`
}

type TipRankModel struct {
	Rank     int64 `db:"tip_rank"`
	TotalTip int64 `db:"total_tip"`
}

type ReactionRank struct {
	Rank          int64  `json:"reaction_rank"`
	TotalReaction int64  `json:"total_reaction"`
	EmojiName     string `json:"emoji_name"`
}

type ReactionRankModel struct {
	Rank          int64  `db:"reaction_rank"`
	TotalReaction int64  `db:"total_reaction"`
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
	if err := dbConn.SelectContext(ctx, &viewedLivestreams, "SELECT user_id, livestream_id FROM livestream_viewers_history WHERE user_id = ?", userModel.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tipRankModelPerLivestreams, err := queryTotalTipRankPerViewedLivestream(ctx, viewedLivestreams)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tipRankPerLivestream := make(map[int64]TipRank)
	for livestreamID, tipRankModel := range tipRankModelPerLivestreams {
		tipRankPerLivestream[livestreamID] = TipRank{
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

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tipRankModels := []TipRankModel{}
	query := "SELECT SUM(tip) AS total_tip, RANK() OVER(ORDER BY SUM(tip) DESC) AS tip_rank " +
		"FROM livecomments GROUP BY livestream_id " +
		"HAVING livestream_id = ? ORDER BY total_tip DESC LIMIT 3"
	if err := dbConn.SelectContext(ctx, &tipRankModels, query, livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reactionRankModels := []ReactionRankModel{}
	query = "SELECT COUNT(*) AS total_reaction, emoji_name, RANK() OVER(ORDER BY COUNT(*) DESC) AS reaction_rank " +
		"FROM reactions GROUP BY livestream_id, emoji_name " +
		"HAVING livestream_id = ? ORDER BY total_reaction DESC LIMIT 3"
	if err := dbConn.SelectContext(ctx, &reactionRankModels, query, livestreamID); err != nil {
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
			EmojiName:     reactionRankModels[i].EmojiName,
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
) (map[int64]TipRankModel, error) {
	totalTipRankPerLivestream := make(map[int64]TipRankModel)
	// get total tip per viewed livestream
	for _, viewedLivestream := range viewedLivestreams {
		totalTip := int64(0)
		if err := dbConn.GetContext(ctx, &totalTip, "SELECT SUM(tip) FROM livecomments WHERE user_id = ? AND livestream_id = ?", viewedLivestream.UserID, viewedLivestream.LivestreamID); err != nil {
			return totalTipRankPerLivestream, err
		}

		totalTipRankPerLivestream[viewedLivestream.LivestreamID] = TipRankModel{
			TotalTip: totalTip,
		}
	}

	for livestreamID, stat := range totalTipRankPerLivestream {
		query := "SELECT SUM(tip) AS total_tip, RANK() OVER(ORDER BY SUM(tip) DESC) AS tip_rank " +
			"FROM livecomments GROUP BY livestream_id " +
			"HAVING livestream_id = ? AND total_tip = ?"

		rank := TipRankModel{}
		if err := dbConn.GetContext(ctx, &rank, query, livestreamID, stat.TotalTip); err != nil {
			return totalTipRankPerLivestream, err
		}

		totalTipRankPerLivestream[livestreamID] = rank

	}

	return totalTipRankPerLivestream, nil
}
