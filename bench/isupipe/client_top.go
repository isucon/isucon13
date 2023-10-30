package isupipe

import (
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/benchscore"
)

type Tag struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	// CreatedAt is the created timestamp that forms an UNIX time.
	CreatedAt int `json:"created_at"`
}

type TagsResponse struct {
	Tags []*Tag `json:"tags"`
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
	}

	benchscore.AddScore(benchscore.SuccessGetTags)
	return tags, nil
}

func (c *Client) GetRandomTags(ctx context.Context, n int) ([]int, error) {
	resp, err := c.GetTags(ctx)
	if err != nil {
		return nil, err
	}
	rand.Shuffle(len(resp.Tags), func(i, j int) {
		resp.Tags[i], resp.Tags[j] = resp.Tags[j], resp.Tags[i]
	})

	var tags []int
	for i := 0; i < n; i++ {
		tags = append(tags, resp.Tags[i].Id)
	}

	return tags, nil
}
