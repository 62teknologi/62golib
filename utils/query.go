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
	orders := append(ctx.QueryArray("order"), ctx.QueryArray("order[]")...)

	//todo : should may filter by join table
	table := query.Statement.Table

	if orders != nil {
		for _, order := range orders {
			query.Order(table + "." + order)
		}
	} else {
		query.Order(table + ".id desc")
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

				//todo : should may filter by join table
				table := query.Statement.Table

				if values[0] != "" {
					if val == "string" {
						query.Where("LOWER("+table+"."+name+") LIKE LOWER(?)", "%"+values[0]+"%")
						continue
					}

					if val == "timestamp" {
						query.Where("DATE("+table+"."+name+") = ?", values[0])
						continue
					}

					if val == "beetwen" {
						query.Where(table+"."+name+" >= ?", values[0])

						if len(values) >= 2 {
							query.Where(table+"."+name+" <= ?", values[1])
						}

						continue
					}

					if val == "boolean" {
						if len(values) >= 2 && values[0] != values[1] {
							continue
						}

						if num, _ := strconv.Atoi(values[0]); num >= 1 {
							query.Where(table+"."+name+" >= ?", 1)
						} else {
							query.Where(table+"."+name+" = ?", 0)
						}

						continue
					}

					query.Where(table+"."+name+" IN ?", values)
				} else {
					query.Where(table + "." + name + " IS NULL")
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
				orConditions = append(orConditions, "LOWER("+query.Statement.Table+"."+v.(string)+") LIKE LOWER('%"+search+"%')")
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
		for name, v := range transformer["belongs_to"].(map[string]any) {
			v := v.(map[string]any)
			table := v["table"].(string)
			query.Joins("left join " + table + " as " + name + " on " + query.Statement.Table + "." + v["fk"].(string) + " = " + name + ".id")

			*columns = append(*columns, query.Statement.Table+"."+v["fk"].(string))

			for _, val := range v["columns"].([]any) {
				*columns = append(*columns, name+"."+val.(string)+" as "+name+"_"+val.(string))
			}

		}
	}
}

func SetOperation(query *gorm.DB, transformer map[string]any, columns *[]string) {
	if transformer["operation"] != nil {
		for i, v := range transformer["operation"].(map[string]any) {
			*columns = append(*columns, "("+v.(string)+") as operation_"+i)
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
		if result["id"] != nil {
			ids = append(ids, strconv.Itoa(ConvertToInt(result["id"])))
		}
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
		for name, v := range transformer["belongs_to"].(map[string]any) {
			v := v.(map[string]any)
			values := map[string]any{}

			for _, val := range v["columns"].([]any) {
				values[val.(string)] = value[name+"_"+val.(string)]
				//delete(transformer, v["fk"].(string))
			}

			transformer[name] = values
		}
	}

	delete(transformer, "belongs_to")
}

func AttachOperation(transformer, value map[string]any) {
	if transformer["operation"] != nil {
		operation := map[string]any{}

		for i, _ := range transformer["operation"].(map[string]any) {
			operation[i] = value["operation_"+i]
		}

		transformer["operation"] = operation
	}
}

func GetSummary(transformer map[string]any, values []map[string]any) map[string]any {
	summary := map[string]any{}

	if transformer["summary"] != nil {
		if s := transformer["summary"].(map[string]any); s["total"] != "" {
			var total int32 = 0
			for _, v := range values {
				switch val := v[s["total"].(string)].(type) {
				case int32:
					total += val
				case float64:
					total += int32(val)
				}
				delete(v, "summary")
			}
			summary["total"] = total
		}
	}

	return summary
}
