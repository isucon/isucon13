package isupipe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/bencherror"
)

type Tag struct {
	ID   int64  `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type TagsResponse struct {
	Tags []*Tag `json:"tags" validate:"required,dive,required"`
}

func (c *Client) GetTagsWithUser(ctx context.Context, streamerName string, opts ...ClientOption) (*TagsResponse, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	if err := c.setStreamerURL(streamerName); err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	req, err := c.agent.NewRequest(http.MethodGet, "/api/tag", nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var tags *TagsResponse
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
			return nil, err
		}

		if err := ValidateResponse(req, tags); err != nil {
			return nil, err
		}
	}

	return tags, nil
}

func (c *Client) GetTags(ctx context.Context, opts ...ClientOption) (*TagsResponse, error) {
	var (
		defaultStatusCode = http.StatusOK
		o                 = newClientOptions(defaultStatusCode, opts...)
	)

	req, err := c.agent.NewRequest(http.MethodGet, "/api/tag", nil)
	if err != nil {
		return nil, bencherror.NewInternalError(err)
	}

	resp, err := sendRequest(ctx, c.agent, req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != o.wantStatusCode {
		return nil, bencherror.NewHttpStatusError(req, o.wantStatusCode, resp.StatusCode)
	}

	var tags *TagsResponse
	if resp.StatusCode == defaultStatusCode {
		if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
			return nil, err
		}

		if err := ValidateResponse(req, tags); err != nil {
			return nil, err
		}
	}

	return tags, nil
}

func (c *Client) getRandomTags(ctx context.Context, n int) ([]*Tag, error) {
	resp, err := c.GetTags(ctx)
	if err != nil {
		return nil, err
	}
	rand.Shuffle(len(resp.Tags), func(i, j int) {
		resp.Tags[i], resp.Tags[j] = resp.Tags[j], resp.Tags[i]
	})
	if len(resp.Tags) < n {
		return nil, fmt.Errorf("タグが取得できませんでした")
	}

	return resp.Tags[:n], nil
}

func (c *Client) GetRandomLivestreamTags(ctx context.Context, n int) ([]int64, error) {
	tags, err := c.getRandomTags(ctx, n)
	if err != nil {
		return nil, err
	}

	livestreamTags := []int64{}
	for _, tag := range tags {
		livestreamTags = append(livestreamTags, tag.ID)
	}

	return livestreamTags, nil
}

func (c *Client) GetRandomSearchTags(ctx context.Context, n int) ([]string, error) {
	tags, err := c.getRandomTags(ctx, n)
	if err != nil {
		return nil, err
	}

	searchTags := []string{}
	for _, tag := range tags {
		searchTags = append(searchTags, tag.Name)
	}

	return searchTags, nil
}
