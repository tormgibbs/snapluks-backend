package data

import (
	"math"
	"strings"

	"slices"

	"github.com/tormgibbs/snapluks-backend/internal/validator"
)

type Metadata struct {
	CurrentPage 	int `json:"current_page,omitempty"`
	PageSize 			int `json:"page_size,omitempty"`
	FirstPage 		int `json:"first_page,omitempty"`
	LastPage 			int `json:"last_page,omitempty"`
	TotalRecords 	int `json:"total_records,omitempty"`
}

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func CalculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}
	
	return Metadata{
		CurrentPage: page,
		PageSize: pageSize,
		FirstPage: 1,
		LastPage: int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func (f Filters) sortColumn() string {
	if slices.Contains(f.SortSafeList, f.Sort) {
		return strings.TrimPrefix(f.Sort, "-")
	}
	panic("unsafe sort paramter: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than 0")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
}
