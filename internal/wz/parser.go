package wz

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
)

// Node represents a generic WZ XML node
type Node struct {
	XMLName  xml.Name
	Name     string  `xml:"name,attr"`
	Value    string  `xml:"value,attr"`
	Children []Node  `xml:",any"`
}

// GetChild returns a child node by name
func (n *Node) GetChild(name string) *Node {
	for i := range n.Children {
		if n.Children[i].Name == name {
			return &n.Children[i]
		}
	}
	return nil
}

// GetChildren returns all children with the given name
func (n *Node) GetChildren(name string) []*Node {
	var result []*Node
	for i := range n.Children {
		if n.Children[i].Name == name {
			result = append(result, &n.Children[i])
		}
	}
	return result
}

// GetAllChildren returns all direct children (imgdir nodes)
func (n *Node) GetAllChildren() []*Node {
	var result []*Node
	for i := range n.Children {
		if n.Children[i].XMLName.Local == "imgdir" {
			result = append(result, &n.Children[i])
		}
	}
	return result
}

// GetString returns a string value from a child node
func (n *Node) GetString(name string) string {
	for _, child := range n.Children {
		if child.Name == name && child.XMLName.Local == "string" {
			return child.Value
		}
	}
	return ""
}

// GetInt returns an int value from a child node
func (n *Node) GetInt(name string) int {
	for _, child := range n.Children {
		if child.Name == name && (child.XMLName.Local == "int" || child.XMLName.Local == "short") {
			val, _ := strconv.Atoi(child.Value)
			return val
		}
	}
	return 0
}

// GetInt64 returns an int64 value from a child node
func (n *Node) GetInt64(name string) int64 {
	for _, child := range n.Children {
		if child.Name == name && child.XMLName.Local == "int" {
			val, _ := strconv.ParseInt(child.Value, 10, 64)
			return val
		}
	}
	return 0
}

// GetFloat returns a float value from a child node
func (n *Node) GetFloat(name string) float64 {
	for _, child := range n.Children {
		if child.Name == name && (child.XMLName.Local == "float" || child.XMLName.Local == "double") {
			val, _ := strconv.ParseFloat(child.Value, 64)
			return val
		}
	}
	return 0
}

// GetVector returns x,y values from a vector child node
func (n *Node) GetVector(name string) (int, int) {
	for _, child := range n.Children {
		if child.Name == name && child.XMLName.Local == "vector" {
			x, _ := strconv.Atoi(child.GetString("x"))
			y, _ := strconv.Atoi(child.GetString("y"))
			return x, y
		}
	}
	return 0, 0
}

// HasChild checks if a child with the given name exists
func (n *Node) HasChild(name string) bool {
	return n.GetChild(name) != nil
}

// ParseFile parses a WZ XML file and returns the root node
func ParseFile(path string) (*Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var root Node
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse XML %s: %w", path, err)
	}

	return &root, nil
}

