package data

// AreaCode description
type AreaCode struct {
	Code       int
	Name       string
	Path       string
	ParentCode int
	Level      int
}

// AreaCodeTree description
type AreaCodeTree struct {
	Code       int
	Name       string
	Path       string
	ParentCode int
	Level      int
	Children   []*AreaCodeTree
}
