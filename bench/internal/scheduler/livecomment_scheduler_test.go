package scheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNgLivecomment(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    bool
	}{
		{name: "basic", comment: "視覚態度分析、本当にこれでいいの？", want: true},
		{name: "basic2", comment: "この視覚選択性、信頼できる？", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LivecommentScheduler.IsNgLivecomment(tt.comment)
			assert.Equal(t, tt.want, got)
		})
	}
}
