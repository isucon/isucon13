package isupipe

import (
	"context"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/logger"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestValidateLivecomment(t *testing.T) {
	user := User{
		ID:          1,
		Name:        "hoge",
		DisplayName: "fuga",
		Description: "piyo",
		Theme: Theme{
			DarkMode: true,
		},
		IconHash: "abcde",
	}
	livestream := Livestream{
		ID: 1,
		// Owner: user,
		Tags:  []Tag{Tag{ID: 1, Name: ""}},
		Title: "hoge",
		// Description:  "fuga",
		PlaylistUrl:  "foo",
		ThumbnailUrl: "",
		StartAt:      0,
		EndAt:        2,
	}
	livecomment := &Livecomment{
		// ID:         1,
		User:       user,
		Livestream: livestream,
		Comment:    "isu",
		CreatedAt:  1,
	}

	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	err = ValidateResponse(req, livecomment)
	assert.Error(t, err)
	log.Println(err)

	livecomments := []*Livecomment{livecomment}
	err = ValidateSlice(req, livecomments)
	assert.Error(t, err)
	log.Println(err)
}

// スパム処理テスト

func TestModerate(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitTestLogger()
	assert.NoError(t, err)

	client, err := NewClient(
		testLogger,
		agent.WithBaseURL(config.TargetBaseURL),
		agent.WithTimeout(10*time.Minute),
	)
	assert.NoError(t, err)

	user := scheduler.UserScheduler.GetRandomStreamer()
	_, err = client.Register(ctx, &RegisterRequest{
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Description: user.Description,
		Password:    user.RawPassword,
		Theme: Theme{
			DarkMode: user.DarkMode,
		},
	})
	assert.NoError(t, err)

	err = client.Login(ctx, &LoginRequest{
		Username: user.Name,
		Password: user.RawPassword,
	})
	assert.NoError(t, err)

	var (
		startAt = time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC).Unix()
		endAt   = time.Date(2024, 6, 12, 9, 0, 0, 0, time.UTC).Unix()
	)
	livestream, err := client.ReserveLivestream(ctx, user.Name, &ReserveLivestreamRequest{
		Title:        "livestream-test1",
		Description:  "livestream-test1",
		PlaylistUrl:  "https://example.com",
		ThumbnailUrl: "https://example.com",
		StartAt:      startAt,
		EndAt:        endAt,
		Tags:         []int64{},
	})
	assert.NoError(t, err)

	err = client.Moderate(ctx, livestream.ID, livestream.Owner.Name, "test")
	assert.NoError(t, err)
}

// ref. https://github.com/isucon/isucon13/pull/141/files#r1380262831
func TestGetNgWordsBug(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitTestLogger()
	assert.NoError(t, err)

	client, err := NewClient(
		testLogger,
		agent.WithBaseURL(config.TargetBaseURL),
		// FIXME: moderateが遅い
		agent.WithTimeout(1*time.Minute),
	)
	assert.NoError(t, err)

	user := scheduler.UserScheduler.GetRandomStreamer()
	client.Register(ctx, &RegisterRequest{
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Description: user.Description,
		Password:    user.RawPassword,
		Theme: Theme{
			DarkMode: user.DarkMode,
		},
	})

	err = client.Login(ctx, &LoginRequest{
		Username: user.Name,
		Password: user.RawPassword,
	})
	assert.NoError(t, err)

	livestream, err := client.ReserveLivestream(ctx, user.Name, &ReserveLivestreamRequest{
		Title:        "ngword-test",
		Description:  "ngword-test",
		PlaylistUrl:  "https://example.com",
		ThumbnailUrl: "https://example.com",
		StartAt:      time.Date(2024, 4, 20, 0, 0, 0, 0, time.UTC).Unix(),
		EndAt:        time.Date(2024, 4, 20, 5, 0, 0, 0, time.UTC).Unix(),
		Tags:         []int64{},
	})
	assert.NoError(t, err)

	ngWords, err := client.GetNgwords(ctx, livestream.ID, livestream.Owner.Name)
	assert.NoError(t, err)
	for _, ngWord := range ngWords {
		assert.NotZero(t, ngWord.CreatedAt)
	}
}
