package assets

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
)

type Asset struct {
	Path string
	Hash [32]byte
}

// Load は静的ファイルのハッシュ値などをローカルファイルから読み出します
func Load(assetDir string) ([]*Asset, error) {
	assetDir = filepath.Clean(assetDir)
	var assets []*Asset

	err := filepath.Walk(assetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		assets = append(assets, &Asset{
			Path: strings.TrimPrefix(path, assetDir),
			Hash: sha256.Sum256(b),
		})

		return nil
	})
	if err != nil {
		return []*Asset{}, err
	}

	return assets, nil
}
