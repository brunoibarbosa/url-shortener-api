package domain

type SortKind uint8

const (
	SortNone SortKind = iota
	SortAsc
	SortDesc
)
