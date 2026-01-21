package wz

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

type ImgDir struct {
	XMLName xml.Name `xml:"imgdir"`
	Name    string   `xml:"name,attr"`

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

func (d *ImgDir) Get(name string) *ImgDir {
	for i := range d.ImgDirs {
		if d.ImgDirs[i].Name == name {
			return &d.ImgDirs[i]
		}
	}
	return nil
}

func (d *ImgDir) GetInt(name string) (int32, error) {
	for _, n := range d.Ints {
		if n.Name == name {
			return n.Value, nil
		}
	}
	for _, s := range d.Strings {
		if s.Name == name {
			v, err := strconv.ParseInt(s.Value, 10, 32)
			if err != nil {
				return 0, err
			}
			return int32(v), nil
		}
	}
	return 0, fmt.Errorf("int '%s' not found", name)
}

func (d *ImgDir) GetString(name string) (string, error) {
	for _, s := range d.Strings {
		if s.Name == name {
			return s.Value, nil
		}
	}
	return "", fmt.Errorf("string '%s' not found", name)
}

func (d *ImgDir) GetFloat(name string) (float64, error) {
	for _, f := range d.Floats {
		if f.Name == name {
			return f.Value, nil
		}
	}
	return 0, fmt.Errorf("float '%s' not found", name)
}

func (d *ImgDir) GetVector(name string) (x, y int32, err error) {
	for _, v := range d.Vectors {
		if v.Name == name {
			return v.X, v.Y, nil
		}
	}
	return 0, 0, fmt.Errorf("vector '%s' not found", name)
}
