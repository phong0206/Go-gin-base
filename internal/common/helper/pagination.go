package helper

type PaginationQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

func (q PaginationQuery) Normalize() PaginationQuery {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	return q
}

type PaginationResponse[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

func NewPaginationResponse[T any](items []T, total int, query PaginationQuery) PaginationResponse[T] {
	query = query.Normalize()
	totalPages := 0
	if total > 0 {
		totalPages = (total + query.Limit - 1) / query.Limit
	}
	return PaginationResponse[T]{
		Items:      items,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}
}
