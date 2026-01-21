package wz

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type WzDirectory struct {
	p       *WzProvider
	relPath string

	mu    sync.RWMutex
	items map[string]any
	order []string
}

func (d *WzDirectory) Path() string { return d.relPath }

func (d *WzDirectory) Dir(name string) *WzDirectory {
	return &WzDirectory{
		p:       d.p,
		relPath: filepath.Join(d.relPath, name),
	}
}

func (d *WzDirectory) Image(name string) (*WzImage, error) {
	if !strings.HasSuffix(name, ".img.xml") {
		name += ".img.xml"
	}
	return d.p.Image(filepath.Join(d.relPath, name))
}

func (d *WzDirectory) ensureItemsLoaded() error {
	d.mu.RLock()
	if d.items != nil {
		d.mu.RUnlock()
		return nil
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.items != nil {
		return nil
	}

	absDir := filepath.Join(d.p.wzPath, d.relPath)
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	items := make(map[string]any)
	var order []string

	for _, e := range entries {
		name := e.Name()

		if e.IsDir() {
			child := &WzDirectory{
				p:       d.p,
				relPath: filepath.Join(d.relPath, name),
			}
			items[name] = child
			order = append(order, name)
			continue
		}

		if strings.HasSuffix(name, ".img.xml") {
			relImg := filepath.Join(d.relPath, name)
			img, err := d.p.Image(relImg)
			if err != nil {
				return err
			}

			key := strings.TrimSuffix(name, ".img.xml")

			items[key] = img
			order = append(order, key)
		}
	}

	d.items = items
	d.order = order
	return nil
}

type OrderedImages struct {
	Order []string
	Items map[string]*WzImage
}

func (d *WzDirectory) GetAllImages() (OrderedImages, error) {
	if err := d.ensureItemsLoaded(); err != nil {
		return OrderedImages{}, err
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	out := make(map[string]*WzImage)
	var ord []string

	for _, k := range d.order {
		if img, ok := d.items[k].(*WzImage); ok {
			out[k] = img
			ord = append(ord, k)
		}
	}
	
	return OrderedImages{Order: ord, Items: out}, nil
}
