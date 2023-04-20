package utils

import (
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func FilterByQueries(query *gorm.DB, filterable map[string]any, ctx *gin.Context) map[string]any {
	filter := map[string]any{}
	queries := ctx.Request.URL.Query()

	for name, values := range queries {
		name = strings.Replace(name, "[]", "", -1)

		if _, ok := filterable[name]; ok {
			if values[0] != "" {
				query.Where(name+" IN ?", values)
			} else {
				query.Where(name + " IS NULL")
			}
			filter[name] = values
		}
	}

	return filter
}

func SetPagination(query *gorm.DB, ctx *gin.Context) map[string]any {
	if page, _ := strconv.Atoi(ctx.Query("page")); page != 0 {
		var total int64

		DB.Table(query.Statement.Table).Count(&total)
		per_page, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "30"))
		offset := (page - 1) * per_page
		query.Limit(per_page).Offset(offset)

		return map[string]any{
			"total":        total,
			"per_page":     per_page,
			"current_page": page,
			"last_page":    int(math.Ceil(float64(total) / float64(per_page))),
		}
	}

	return map[string]any{}
}
