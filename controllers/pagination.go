package controllers

import (
	mgo "gopkg.in/mgo.v2"
	"math"
)

func PaginationMeta(query *mgo.Query, page int, perPage int) Meta {
	totalCount, _ := query.Count()
	fTotalPages := 0.0
	if totalCount > 0 {
		fTotalPages = math.Ceil(float64(totalCount) / float64(perPage))
	}
	totalPages := int(fTotalPages)
	offset := (page - 1) * perPage

	var nextPage *int
	var prevPage *int

	if page > 1 {
		iPrevPage := page - 1
		prevPage = &iPrevPage
	}

	if page < totalPages {
		iNextPage := page + 1
		nextPage = &iNextPage
	}

	return Meta{CurrentPage: &page, NextPage: nextPage, PrevPage: prevPage, TotalPages: &totalPages, TotalCount: &totalCount, Offset: &offset}
}
