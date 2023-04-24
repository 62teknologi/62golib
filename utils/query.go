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

func SetJoin(query *gorm.DB, transformer map[string]any, columns *[]string) {
	if transformer["belongs_to"] != nil {
		for _, v := range transformer["belongs_to"].(map[string]any) {
			v := v.(map[string]any)
			table := v["table"].(string)
			query.Joins("left join " + table + " on " + query.Statement.Table + "." + v["fk"].(string) + " = " + table + ".id")
			query.Select("products.height").Select("users.id")

			for _, val := range v["columns"].([]any) {
				*columns = append(*columns, table+"."+val.(string)+" as "+table+"_"+val.(string))
			}

		}
	}
}

func AttachJoin(transformer, value map[string]any) {
	if transformer["belongs_to"] != nil {
		for i, v := range transformer["belongs_to"].(map[string]any) {
			v := v.(map[string]any)
			values := map[string]any{}

			for _, val := range v["columns"].([]any) {
				values[val.(string)] = value[v["table"].(string)+"_"+val.(string)]
				//delete(transformer, v["fk"].(string))
			}

			transformer[i] = values
		}
	}

	delete(transformer, "belongs_to")
}
