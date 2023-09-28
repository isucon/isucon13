package isupipe

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/isucon/isucandar/agent"
	"github.com/stretchr/testify/assert"
)

func TestCliet_NetworkErrors(t *testing.T) {
	ctx := context.Background()

	// Arrange
	// ServeMuxオブジェクトなどを用意してルーティングしてもよい
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	})
	// 別goroutine上でリッスンが開始される
	ts := httptest.NewServer(h)
	defer ts.Close()

	client, err := NewClient(agent.WithBaseURL(ts.URL))
	assert.NoError(t, err)

	// NOTE: 呼び出すエンドポイントは何でも良い
	// タグ取得がパラメータがなく簡単であるためこうしている
	client.GetTags(ctx)
	FIXMe

	cli := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", ts.URL, strings.NewReader(""))
	if err != nil {
		t.Errorf("NewRequest failed: %v", err)
	}

	// Act
	resp, err := cli.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// Assertion
	got, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	want := "Hello, client\n"
	if string(got) != want {
		t.Errorf("want %q, but %q", want, got)
	}
}
