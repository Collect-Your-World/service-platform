package pagingUtil

type SortBy string

const (
	ASC  SortBy = "ASC"
	DESC SortBy = "DESC"
)

type Page struct {
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	SortBy  SortBy `json:"sort_by"`
	OrderBy string `json:"order_by"`
}

func (p *Page) LoadDefault() {
	if p.Limit == 0 {
		p.Limit = 100
	}
	if p.SortBy == "" {
		p.SortBy = ASC
	}
	if p.OrderBy == "" {
		p.OrderBy = "id"
	}
}

type PageCursor[T comparable] struct {
	Limit   int    `json:"limit"`
	LastID  T      `json:"last_id"`
	SortBy  SortBy `json:"sort_by"`
	OrderBy string `json:"order_by"`
}
