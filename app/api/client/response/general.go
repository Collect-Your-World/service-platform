package response

type ErrorDetail struct {
	Key     string `json:"key,omitempty"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Page struct {
	Total         int64 `json:"total"`
	Page          int   `json:"page"`
	Size          int   `json:"size"`
	NumberOfPages int   `json:"num_page"`
}

type GeneralResponse[T any] struct {
	Code         int           `json:"code"`
	Message      string        `json:"message,omitempty"`
	Data         T             `json:"data,omitempty"`
	Paging       *Page         `json:"page,omitempty"`
	ErrorDetails []ErrorDetail `json:"error_details,omitempty"`
}

func ToSuccessResponse[T any](data T) GeneralResponse[T] {
	return GeneralResponse[T]{
		Message: "success",
		Data:    data,
	}
}

func ToErrorResponse(code int, message string) GeneralResponse[any] {
	return GeneralResponse[any]{
		Code:    code,
		Message: message,
	}
}

type PaginationResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    []T    `json:"data"`
	Paging  Page   `json:"page"`
}

func ToPaginationResponse[T any](data []T, total int64, page int, size int) PaginationResponse[T] {
	numberOfPages := 0
	if size > 0 {
		numberOfPages = (int(total) + size - 1) / size
	}

	return PaginationResponse[T]{
		Message: "success",
		Data:    data,
		Paging: Page{
			Total:         total,
			Page:          page,
			Size:          size,
			NumberOfPages: numberOfPages,
		},
	}
}
