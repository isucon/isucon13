package isupipe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/bencherror"
)

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

type ReactionRank struct {
	Rank      int    `json:"reaction_rank"`
	EmojiName string `json:"emoji_name"`
}

func (c *Client) GetUserStatistics(ctx context.Context, username string, opts ...ClientOption) (*UserStatistics, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/user/%s/statistics", username)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var stats *UserStatistics
	if resp.StatusCode == defaultStatusCode {
		if json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return nil, err
		}
	}

	return stats, nil
}

func (c *Client) GetLivestreamStatistics(ctx context.Context, livestreamId int, opts ...ClientOption) (*LivestreamStatistics, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	urlPath := fmt.Sprintf("/livestream/%d/statistics", livestreamId)
	req, err := c.agent.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var stats *LivestreamStatistics
	if resp.StatusCode == defaultStatusCode {
		if json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return nil, err
		}
	}

	return stats, nil
}
