package assets

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

func ValidateStaticAssets(contestantLogger *zap.Logger, targetBaseURL string) error {
	assets, err := Load()
	if err != nil {
		return err
	}
	for _, asset := range assets {
		url, err := url.JoinPath(targetBaseURL, asset.Path)
		if err != nil {
			return err
		}
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("ファイルのダウンロードに失敗しました %s", asset.Path)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read")
		}
		actualAssetHash := sha256.Sum256(b)
		if asset.Hash != actualAssetHash {
			return fmt.Errorf("ファイル %s のハッシュが一致しません actual: %x expected: %x", asset.Path, actualAssetHash, asset.Hash)
		}
	}
	return nil
}
