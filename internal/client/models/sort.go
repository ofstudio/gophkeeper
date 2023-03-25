package models

type SortDirection int

const (
	SortAsc SortDirection = iota
	SortDesc
)

type ItemMetaList []*ItemMeta

func (l ItemMetaList) Len() int {
	return len(l)
}

func (l ItemMetaList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type ByCreatedAt struct {
	ItemMetaList
	order SortDirection
}

func (l ByCreatedAt) Less(i, j int) bool {
	return l.order == SortAsc && l.ItemMetaList[i].CreatedAt < l.ItemMetaList[j].CreatedAt
}
