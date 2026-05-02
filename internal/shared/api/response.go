package api

type Response[T any] struct {
	Data  T               `json:"data,omitempty"`
	Meta  *PaginationMeta `json:"meta,omitempty"`
	Error *APIError       `json:"error,omitempty"`
}

type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

type APIError struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details map[string][]string `json:"details,omitempty"`
}
