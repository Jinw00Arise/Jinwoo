package wz

type WzImage struct {
	p       *WzProvider
	relPath string
	root    *ImgDir
}

func (i *WzImage) Path() string  { return i.relPath }
func (i *WzImage) Root() *ImgDir { return i.root }
func (i *WzImage) Get(name string) *ImgDir { // convenience: root.Get(...)
	if i.root == nil {
		return nil
	}
	return i.root.Get(name)
}
