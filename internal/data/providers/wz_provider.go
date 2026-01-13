package providers

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// WzProvider loads and provides access to WZ data
type WzProvider struct {
	wzPath string
	cache  map[string]*ImgDir
	mu     sync.RWMutex
}

func NewWzProvider(wzPath string) *WzProvider {
	return &WzProvider{
		wzPath: wzPath,
		cache:  make(map[string]*ImgDir),
	}
}

// GetImg loads an IMG file (e.g., "Map.wz/Map/Map1/100000000.img.xml")
func (p *WzProvider) GetImg(path string) (*ImgDir, error) {
	// Check cache
	p.mu.RLock()
	if cached, exists := p.cache[path]; exists {
		p.mu.RUnlock()
		return cached, nil
	}
	p.mu.RUnlock()

	// Load from disk
	fullPath := filepath.Join(p.wzPath, path)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open WZ file %s: %w", fullPath, err)
	}
	defer file.Close()

	var root ImgDir
	if err := xml.NewDecoder(file).Decode(&root); err != nil {
		return nil, fmt.Errorf("failed to parse WZ XML %s: %w", fullPath, err)
	}

	// Cache it
	p.mu.Lock()
	p.cache[path] = &root
	p.mu.Unlock()

	return &root, nil
}

// ImgDir represents a WZ imgdir node
type ImgDir struct {
	XMLName  xml.Name     `xml:"imgdir"`
	Name     string       `xml:"name,attr"`
	ImgDirs  []ImgDir     `xml:"imgdir"`
	Ints     []IntNode    `xml:"int"`
	Floats   []FloatNode  `xml:"float"`
	Strings  []StrNode    `xml:"string"`
	Vectors  []VectorNode `xml:"vector"`
	Canvases []Canvas     `xml:"canvas"`
}

type IntNode struct {
	Name  string `xml:"name,attr"`
	Value int32  `xml:"value,attr"`
}

type FloatNode struct {
	Name  string  `xml:"name,attr"`
	Value float64 `xml:"value,attr"`
}

type StrNode struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type VectorNode struct {
	Name string `xml:"name,attr"`
	X    int32  `xml:"x,attr"`
	Y    int32  `xml:"y,attr"`
}

type Canvas struct {
	Name    string       `xml:"name,attr"`
	Width   int32        `xml:"width,attr"`
	Height  int32        `xml:"height,attr"`
	ImgDirs []ImgDir     `xml:"imgdir"`
	Vectors []VectorNode `xml:"vector"`
}

// Get finds a child imgdir by name
func (d *ImgDir) Get(name string) *ImgDir {
	for i := range d.ImgDirs {
		if d.ImgDirs[i].Name == name {
			return &d.ImgDirs[i]
		}
	}
	return nil
}

// GetInt extracts an integer value
func (d *ImgDir) GetInt(name string) (int32, error) {
	for _, intNode := range d.Ints {
		if intNode.Name == name {
			return intNode.Value, nil
		}
	}
	// Check if it's a string that can be parsed
	for _, strNode := range d.Strings {
		if strNode.Name == name {
			val, err := strconv.ParseInt(strNode.Value, 10, 32)
			if err != nil {
				return 0, fmt.Errorf("failed to parse int from string '%s': %w", strNode.Value, err)
			}
			return int32(val), nil
		}
	}
	return 0, fmt.Errorf("int value '%s' not found", name)
}

// GetIntOrDefault extracts an integer value with a default fallback
func (d *ImgDir) GetIntOrDefault(name string, defaultValue int32) int32 {
	val, err := d.GetInt(name)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetFloat extracts a float value
func (d *ImgDir) GetFloat(name string) (float64, error) {
	for _, floatNode := range d.Floats {
		if floatNode.Name == name {
			return floatNode.Value, nil
		}
	}
	return 0, fmt.Errorf("float value '%s' not found", name)
}

// GetFloatOrDefault extracts a float value with a default fallback
func (d *ImgDir) GetFloatOrDefault(name string, defaultValue float64) float64 {
	val, err := d.GetFloat(name)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetString extracts a string value
func (d *ImgDir) GetString(name string) (string, error) {
	for _, strNode := range d.Strings {
		if strNode.Name == name {
			return strNode.Value, nil
		}
	}
	// Check if it's an int that should be converted
	for _, intNode := range d.Ints {
		if intNode.Name == name {
			return strconv.FormatInt(int64(intNode.Value), 10), nil
		}
	}
	return "", fmt.Errorf("string value '%s' not found", name)
}

// GetStringOrDefault extracts a string value with a default fallback
func (d *ImgDir) GetStringOrDefault(name string, defaultValue string) string {
	val, err := d.GetString(name)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetVector extracts a vector value
func (d *ImgDir) GetVector(name string) (Vector, error) {
	for _, vecNode := range d.Vectors {
		if vecNode.Name == name {
			return Vector{X: vecNode.X, Y: vecNode.Y}, nil
		}
	}
	return Vector{}, fmt.Errorf("vector value '%s' not found", name)
}

// Vector represents a 2D point
type Vector struct {
	X int32
	Y int32
}

// Rect represents a rectangle
type Rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// GetRect extracts a rectangle from lt/rb vectors
func (d *ImgDir) GetRect() (Rect, error) {
	lt, err := d.GetVector("lt")
	if err != nil {
		return Rect{}, fmt.Errorf("failed to get 'lt' vector: %w", err)
	}
	rb, err := d.GetVector("rb")
	if err != nil {
		return Rect{}, fmt.Errorf("failed to get 'rb' vector: %w", err)
	}
	return Rect{
		Left:   lt.X,
		Top:    lt.Y,
		Right:  rb.X,
		Bottom: rb.Y,
	}, nil
}
