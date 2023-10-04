package isupipe

import (
	"log"
	"os"
	"testing"

	"github.com/isucon/isucon13/bench/internal/benchtest"
)

// NOTE: パッケージ内では並列にテストを実行しないため、この変数で競合が起きない
var webappIPAddress string

func TestMain(m *testing.M) {
	testResource, err := benchtest.Setup("isupipe")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	webappIPAddress = testResource.WebappIPAddress()
	code := m.Run()

	if err := benchtest.Teardown(testResource); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	os.Exit(code)
}
