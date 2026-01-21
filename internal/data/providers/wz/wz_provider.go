package wz

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type WzProvider struct {
	wzPath string

	mu    sync.RWMutex
	cache map[string]*WzImage // cache images by relative path
}

func NewWzProvider(wzPath string) *WzProvider {
	return &WzProvider{
		wzPath: wzPath,
		cache:  make(map[string]*WzImage),
	}
}

// Dir returns a directory handle rooted at wzPath/relPath
func (p *WzProvider) Dir(relPath string) *WzDirectory {
	return &WzDirectory{
		p:       p,
		relPath: filepath.Clean(relPath),
	}
}

// Image loads an image XML (e.g. "Map.wz/Map/Map1/100000000.img.xml")
func (p *WzProvider) Image(relImgXmlPath string) (*WzImage, error) {
	relImgXmlPath = filepath.Clean(relImgXmlPath)

	// cache
	p.mu.RLock()
	if img, ok := p.cache[relImgXmlPath]; ok {
		p.mu.RUnlock()
		return img, nil
	}
	p.mu.RUnlock()

	fullPath := filepath.Join(p.wzPath, relImgXmlPath)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("open wz image %s: %w", fullPath, err)
	}
	defer f.Close()

	var root ImgDir
	if err := xml.NewDecoder(f).Decode(&root); err != nil {
		return nil, fmt.Errorf("decode wz image %s: %w", fullPath, err)
	}

	img := &WzImage{
		p:       p,
		relPath: relImgXmlPath,
		root:    &root,
	}

	p.mu.Lock()
	p.cache[relImgXmlPath] = img
	p.mu.Unlock()

	return img, nil
}
