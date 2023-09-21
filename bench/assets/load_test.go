package assets

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	assets, err := Load("../assets/testdata")
	assert.NoError(t, err)
	assert.Len(t, assets, 6)

	for _, asset := range assets {
		path := filepath.Join("testdata", asset.Path)
		_, err := os.Stat(path)
		assert.NoError(t, err)
		assert.False(t, os.IsNotExist(err))

		b, err := os.ReadFile(path)
		assert.NoError(t, err)

		hash := sha256.Sum256(b)
		assert.Equal(t, hash, asset.Hash)
	}
}
