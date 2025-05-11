package models

// Pagination contains information about the current page and total items
type Pagination struct {
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	PageSize     int `json:"page_size"`
	TotalItems   int `json:"total_items"`
	HasNextPage  bool `json:"has_next_page"`
	HasPrevPage  bool `json:"has_prev_page"`
}

// NewPagination creates a new pagination object
func NewPagination(page, pageSize, totalItems int) *Pagination {
	totalPages := calculateTotalPages(pageSize, totalItems)
	
	return &Pagination{
		CurrentPage:  page,
		TotalPages:   totalPages,
		PageSize:     pageSize,
		TotalItems:   totalItems,
		HasNextPage:  page < totalPages,
		HasPrevPage:  page > 1,
	}
}

// calculateTotalPages determines how many pages are needed for the items
func calculateTotalPages(pageSize, totalItems int) int {
	if pageSize <= 0 {
		return 0
	}
	
	// Calculate total pages, ensuring we round up
	totalPages := totalItems / pageSize
	if totalItems%pageSize > 0 {
		totalPages++
	}
	
	return totalPages
}