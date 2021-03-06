package parser

import (
	"image"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/salviati/go-tmx/tmx"
)

var (
	farmCache     *Map
	farmCacheOnce sync.Once
)

var farmFileMap = path.Clean(path.Join(os.Getenv("HOME"), "Content/TMX/Farm.tmx"))

// Map represents a tile map loaded from disk and cached in memory.
type Map struct {
	TMX *tmx.Map
	// source determines where the map was loaded from, which is needed as a reference
	// for loading other assets from the TMX.
	source string

	mu           sync.Mutex
	imageSources map[string]image.Image
}

// FetchSeasonSource obtains the source image specified in source, but also replaces the
// source string to match the current season.
func (m *Map) FetchSeasonSource(source, season string) (image.Image, error) {
	// Expect the _ suffix because there's a "seasonobjects.png" that we should not replace.
	return m.FetchSource(strings.Replace(source, "spring_", season+"_", 1))
}

// FetchSource obtains the source image specified in source.
func (m *Map) FetchSource(s string) (image.Image, error) {
	s = strings.Replace(s, `\`, "/", -1)
	m.mu.Lock()
	defer m.mu.Unlock()
	img, ok := m.imageSources[s]
	if ok {
		return img, nil
	}

	f, err := os.Open(path.Join(path.Dir(m.source), s))
	if err != nil {
		return nil, err
	}
	img, _, err = image.Decode(f)
	if err == nil {
		m.imageSources[s] = img
	}
	return img, err
}

// LoadFarmMap loads the TMX farm map. Calling this function multiple times
// always returns the same content.
func LoadFarmMap() *Map {
	farmCacheOnce.Do(func() {
		f, err := os.Open(farmFileMap)
		if err != nil {
			panic(err)
		}
		m, err := tmx.Read(f)
		if err != nil {
			panic(err)
		}

		farmCache = &Map{TMX: m, source: farmFileMap, imageSources: map[string]image.Image{}}
	})
	return farmCache
}
