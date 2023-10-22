package isupipe

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestClient_Timeout(t *testing.T) {
	ctx := context.Background()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		fmt.Fprintln(w, `{"tags": []}`)
	})
	ts := httptest.NewServer(h)
	defer ts.Close()

	client, err := NewClient(agent.WithBaseURL(ts.URL), agent.WithTimeout(1*time.Microsecond))
	assert.NoError(t, err)

	// NOTE: 呼び出すエンドポイントは何でも良い
	// タグ取得がパラメータがなく簡単であるためこうしている
	_, err = client.GetTags(ctx)
	assert.True(t, errors.Is(err, bencherror.ErrTimeout))
}

func TestClient_Spam(t *testing.T) {
	ctx := context.Background()

	client, err := NewClient(
		agent.WithBaseURL(webappIPAddress),
	)
	assert.NoError(t, err)

	streamer, err := scheduler.UserScheduler.PrepareStreamer()
	assert.NoError(t, err)
	err = client.Login(ctx, &LoginRequest{
		UserName: streamer.Name,
		Password: streamer.RawPassword,
	})
	assert.NoError(t, err)

	livestreams, err := client.GetLivestreams(ctx)
	assert.NoError(t, err)

	myLivestreamId := -1
	for _, ls := range livestreams {
		if ls.Owner.Id == streamer.UserId {
			myLivestreamId = ls.Id
		}
	}

	assert.NotEqual(t, -1, myLivestreamId, "the streamer doesn't have own livestream")

	// FIXME: livecomment schedulerから取得するように変更
	_, err = client.PostLivecomment(ctx, myLivestreamId, &PostLivecommentRequest{
		Comment: "test is greaaaaaaaaaaaaaaaaaaaaat!",
		Tip:     0,
	})
	assert.NoError(t, err)

	// NGワードを追加(moderate)し、再度投稿してスパム判定されることをチェック
	err = client.Moderate(ctx, myLivestreamId, "test")
	assert.NoError(t, err)

	_, err = client.PostLivecomment(ctx, myLivestreamId, &PostLivecommentRequest{
		Comment: "test is greaaaaaaaaaaaaaaaaaaaaat!",
		Tip:     0,
	}, WithStatusCode(http.StatusBadRequest))
	assert.NoError(t, err)

}
