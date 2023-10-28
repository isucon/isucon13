package isupipe

import (
	"context"
	"encoding/json"
	"io"
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

	req, err := c.agent.NewRequest(http.MethodGet, "/tag", nil)
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
