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
	"github.com/isucon/isucon13/bench/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestClient_Timeout(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitTestLogger()
	assert.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		fmt.Fprintln(w, `{"tags": []}`)
	})
	ts := httptest.NewServer(h)
	defer ts.Close()

	client, err := NewClient(testLogger, agent.WithBaseURL(ts.URL), agent.WithTimeout(1*time.Microsecond))
	assert.NoError(t, err)

	// NOTE: 呼び出すエンドポイントは何でも良い
	// タグ取得がパラメータがなく簡単であるためこうしている
	_, err = client.GetTags(ctx)
	assert.True(t, errors.Is(err, bencherror.ErrTimeout))
}
