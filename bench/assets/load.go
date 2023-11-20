package assets

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"strings"
)

//go:embed data/hash.txt
var hashList string

type Asset struct {
	Path string
	Hash [32]byte
}

// Load は静的ファイルのハッシュ値などをローカルファイルから読み出します
func Load() ([]*Asset, error) {
	return load(hashList)
}

func load(hashList string) ([]*Asset, error) {
	buff := bytes.NewBufferString(hashList)

	var assets []*Asset
	for {
		line, err := buff.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			continue
		}

		pathname := strings.TrimPrefix(parts[1], ".")

		// hex to byte
		hash, err := hex.DecodeString(parts[0])
		if err != nil {
			continue
		}
		fixed := [32]byte{}
		copy(fixed[:], hash)

		assets = append(assets, &Asset{
			Path: pathname,
			Hash: fixed,
		})
	}

	return assets, nil
}
