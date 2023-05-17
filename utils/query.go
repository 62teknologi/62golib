package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetOrderByQuery(query *gorm.DB, ctx *gin.Context) {
	orders := ctx.QueryArray("order[]")

	if orders != nil {
		for _, order := range orders {
			query.Order(order)
		}
	} else {
		query.Order("id desc")
	}
}

func SetFilterByQuery(query *gorm.DB, transformer map[string]any, ctx *gin.Context) map[string]any {
	filter := map[string]any{}
	queries := ctx.Request.URL.Query()

	if transformer["filterable"] != nil {
		filterable := transformer["filterable"].(map[string]any)

		for name, values := range queries {
			name = strings.Replace(name, "[]", "", -1)

			if val, ok := filterable[name]; ok {
				filter[name] = values

				if values[0] != "" {
					if val == "string" {
						query.Where(query.Statement.Table+"."+name+" ILIKE ?", "%"+values[0]+"%")
						continue
					}

					if val == "timestamp" {
						query.Where("DATE("+query.Statement.Table+"."+name+") = ?", values[0])
						continue
					}

					query.Where(query.Statement.Table+"."+name+" IN ?", values)
				} else {
					query.Where(query.Statement.Table + "." + name + " IS NULL")
				}
			}
		}
	}

	delete(transformer, "filterable")

	return filter
}

func SetGlobalSearch(query *gorm.DB, transformer map[string]any, ctx *gin.Context) map[string]any {
	filter := map[string]any{}

	if transformer["searchable"] != nil {
		searchable := transformer["searchable"].([]interface{})
		search := ctx.Query("search")

		if search != "" {
			filter["value"] = search
			filter["column"] = searchable
			orConditions := []string{}

			for _, v := range searchable {
				orConditions = append(orConditions, query.Statement.Table+"."+v.(string)+" ILIKE '%"+search+"%'")
			}

			query.Where(strings.Join(orConditions, " OR "))

		}
	}

	delete(transformer, "searchable")

	return filter
}

func SetPagination(query *gorm.DB, ctx *gin.Context) map[string]any {
	if page, _ := strconv.Atoi(ctx.Query("page")); page != 0 {
		var total int64

		if err := query.Count(&total).Error; err != nil {
			fmt.Println(err)
		}

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

func SetBelongsTo(query *gorm.DB, transformer map[string]any, columns *[]string) {
	if transformer["belongs_to"] != nil {
		for _, v := range transformer["belongs_to"].(map[string]any) {
			v := v.(map[string]any)
			table := v["table"].(string)
			query.Joins("left join " + table + " on " + query.Statement.Table + "." + v["fk"].(string) + " = " + table + ".id")

			*columns = append(*columns, query.Statement.Table+"."+v["fk"].(string))

			for _, val := range v["columns"].([]any) {
				*columns = append(*columns, table+"."+val.(string)+" as "+table+"_"+val.(string))
			}

		}
	}
}

func AttachHasMany(transformer map[string]any) {
	if transformer["has_many"] != nil {
		for i, v := range transformer["has_many"].(map[string]any) {
			v := v.(map[string]any)
			values := []map[string]any{}
			colums := convertAnyToString(v["columns"].([]any))
			fk := v["fk"].(string)

			if err := DB.Table(v["table"].(string)).Select(colums).Where(fk+" = ?", transformer["id"]).Find(&values).Error; err != nil {
				fmt.Println(err)
			}

			transformer[i] = values
		}
	}

	delete(transformer, "has_many")
}

func MultiAttachHasMany(results []map[string]any) {
	ids := []string{}

	for _, result := range results {
		ids = append(ids, strconv.Itoa(int(result["id"].(int32))))
	}

	if len(results) > 0 {
		transformer := results[0]

		if transformer["has_many"] != nil {
			for i, v := range transformer["has_many"].(map[string]any) {
				v := v.(map[string]any)
				values := []map[string]any{}
				fk := v["fk"].(string)
				colums := convertAnyToString(v["columns"].([]any))
				colums = append(colums, fk)

				if err := DB.Table(v["table"].(string)).Select(colums).Where(fk+" in ?", ids).Find(&values).Error; err != nil {
					fmt.Println(err)
				}

				for _, result := range results {
					result[i] = filterSliceByMapIndex(values, fk, result["id"])
					delete(result, "has_many")
				}
			}
		}
	}
}

func AttachBelongsTo(transformer, value map[string]any) {
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
