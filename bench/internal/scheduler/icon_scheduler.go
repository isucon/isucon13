package scheduler

import (
	"crypto/sha256"
	"embed"
	"log"
	"math/rand"
	"path/filepath"
)

//go:embed images/*
var images embed.FS

var IconSched = mustNewIconScheduler()

type Image struct {
	Image []byte
	Hash  [32]byte
}

type IconScheduler struct {
	images []*Image
}

func mustNewIconScheduler() *IconScheduler {
	sched := new(IconScheduler)
	const dir = "images"

	entries, err := images.ReadDir(dir)
	if err != nil {
		log.Fatalln(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if filepath.Ext(filename) != ".jpg" {
			continue
		}

		path := filepath.Join(dir, filename)
		b, err := images.ReadFile(path)
		if err != nil {
			log.Fatalln(err)
		}

		sched.images = append(sched.images, &Image{
			Image: b,
			Hash:  sha256.Sum256(b),
		})
	}

	return sched
}

func (s *IconScheduler) GetRandomIcon() *Image {
	idx := rand.Intn(len(s.images))
	return s.images[idx]
}
