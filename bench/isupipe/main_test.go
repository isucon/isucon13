package isupipe

import (
	"testing"
)

func TestMain(m *testing.M) {
	// client, err := NewClient(
	// 	agent.WithTimeout(1 * time.Minute),
	// )
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// log.Println("main initialize ...")
	// if _, err := client.Initialize(context.Background()); err != nil {
	// 	log.Fatalln(err)
	// }
	// log.Println("main initialize done")

	m.Run()
}
