package request

type SortBy string

const (
	ASC  SortBy = "ASC"
	DESC SortBy = "DESC"
)

type PaginationRequest struct {
	Page    int    `json:"page" query:"page" form:"page"`
	Size    int    `json:"size" query:"size" form:"size"`
	SortBy  SortBy `json:"sort_by" query:"sort_by" form:"sort_by"`
	OrderBy string `json:"order_by" query:"order_by" form:"order_by"`
}

func (p *PaginationRequest) LoadDefaultValues(desireSize ...int) {
	if p.Page < 1 {
		p.Page = 1
	}

	size := 10
	if len(desireSize) > 0 && desireSize[0] > 0 {
		size = desireSize[0]
	}

	if p.Size < 1 {
		p.Size = size
	}

	if p.SortBy == "" {
		p.SortBy = DESC
	}
	if p.OrderBy == "" {
		p.OrderBy = "created_at"
	}
}
