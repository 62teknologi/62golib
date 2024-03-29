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

					if val == "belongs_to" {
						query.Where(name+" IN ?", values)
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

func SetBelongsTo(query *gorm.DB, transformer map[string]any, columns *[]string, ctx *gin.Context) {
	if transformer["belongs_to"] != nil {
		for name, v := range transformer["belongs_to"].(map[string]any) {
			v := v.(map[string]any)
			table := v["table"].(string)
			fk := v["fk"].(string)

			for _, val := range v["columns"].([]any) {
				*columns = append(*columns, name+"."+val.(string)+" as "+name+"_"+val.(string))
			}

			if v["composite"] != nil {
				composite := v["composite"].(string)
				compositeValue := ctx.DefaultQuery("composite_"+composite, "0")
				query.Joins("left join " + table + " as " + name + " on " + query.Statement.Table + ".id = " + name + "." + fk + " and " + name + "." + composite + "=" + compositeValue)
				*columns = append(*columns, query.Statement.Table+"."+composite)
				*columns = append(*columns, "CASE WHEN "+name+"."+composite+" > 0 THEN 1 ELSE 0 END AS "+name+"_is_true")
				v["columns"] = append(v["columns"].([]any), "is_true")
			} else {
				query.Joins("left join " + table + " as " + name + " on " + query.Statement.Table + "." + fk + " = " + name + ".id")
				*columns = append(*columns, query.Statement.Table+"."+fk)
			}

			//need better variable name
			if v["belongs_to"] != nil {
				for name2, v2 := range v["belongs_to"].(map[string]any) {
					v2 := v2.(map[string]any)
					table2 := v2["table"].(string)
					fk2 := v2["fk"].(string)

					for _, val2 := range v2["columns"].([]any) {
						*columns = append(*columns, name+"_"+name2+"."+val2.(string)+" as  "+name+"_"+name2+"_"+val2.(string))
					}

					if v2["composite"] != nil {
						composite := v2["composite"].(string)
						compositeValue := ctx.DefaultQuery("composite_"+composite, "0")
						query.Joins("left join " + table2 + " as " + name + "_" + name2 + " on " + name + ".id = " + name + "_" + name2 + "." + fk2 + " and " + name + "_" + name2 + "." + composite + "=" + compositeValue)
						*columns = append(*columns, query.Statement.Table+"."+composite)
						*columns = append(*columns, "CASE WHEN "+name+"_"+name2+"."+composite+" > 0 THEN 1 ELSE 0 END AS "+name+"_"+name2+"_is_true")
						v2["columns"] = append(v2["columns"].([]any), "is_true")
					} else {
						query.Joins("left join " + table2 + " as " + name + "_" + name2 + " on " + name + "." + fk2 + " = " + name + "_" + name2 + ".id")
						*columns = append(*columns, query.Statement.Table+"."+fk)
					}

				}
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

			// need implement limit
			if err := DB.Table(v["table"].(string)).Select(colums).Where(fk+" = ?", transformer["id"]).Find(&values).Error; err != nil {
				fmt.Println(err)
			}

			transformer[i] = values

			if v["count"] != nil {
				count := map[string]any{}

				if err := DB.Table(v["table"].(string)).Select("COUNT(id) as count").Where(fk+" = ?", transformer["id"]).Take(&count).Error; err != nil {
					fmt.Println(err)
				}

				transformer[i+"_count"] = count["count"]
			}
		}
	}

	delete(transformer, "has_many")
}

func AttachManyToMany(transformer map[string]any) {
	if transformer["many_to_many"] != nil {
		for i, v := range transformer["many_to_many"].(map[string]any) {
			v := v.(map[string]any)
			values := []map[string]any{}
			columns := convertAnyToString(v["columns"].([]any))
			fk1 := v["fk1"].(string)
			fk2 := v["fk2"].(string)

			if err := DB.Table(v["table"].(string)).Select("*").Where(fk1+" = ?", transformer["id"]).Find(&values).Error; err != nil {
				fmt.Println(err)
			}

			var m2mValues []map[string]any
			for _, val := range values {
				var m2mValue map[string]any
				if err := DB.Table(i).Select(columns).Where("id = ?", val[fk2]).Find(&m2mValue).Error; err != nil {
					fmt.Println(err)
				}
				m2mValues = append(m2mValues, m2mValue)
			}

			transformer[i] = m2mValues

			if v["count"] != nil {
				count := map[string]any{}

				if err := DB.Table(v["table"].(string)).Select("COUNT(id) as count").Where(fk1+" = ?", transformer["id"]).Take(&count).Error; err != nil {
					fmt.Println(err)
				}

				transformer[i+"_count"] = count["count"]
			}
		}
	}
}

func MultiAttachHasMany(results []map[string]any, ctx *gin.Context) {
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
				fk := v["fk"].(string)
				colums := convertAnyToString(v["columns"].([]any))
				colums = append(colums, fk)
				values := []map[string]any{}

				if limit := ConvertToInt(v["limit"]); limit > 0 {
					subSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
						return tx.
							Select("*", "ROW_NUMBER() OVER (PARTITION BY "+fk+" ORDER BY id) AS rn").
							Table(v["table"].(string)).
							Where(fk, ids).
							Find(&[]any{})
					})

					query := DB.Table("(" + subSql + ") AS subQ")

					for i := range colums {
						colums[i] = "subQ." + colums[i]
					}

					SetBelongsTo(query, v, &colums, ctx)

					if err := query.Select(colums).Where("rn <= " + strconv.Itoa(limit)).Find(&values).Error; err != nil {
						fmt.Println(err)
					}

					// todo : need to fix, should return null if belong to not exist
					values = MultiMapValuesShifter2(v, values)
				} else {
					if err := DB.Table(v["table"].(string)).Select(colums).Where(fk+" in ?", ids).Find(&values).Error; err != nil {
						fmt.Println(err)
					}
				}

				for _, result := range results {
					result[i] = filterSliceByMapIndex(values, fk, result["id"])
					delete(result, "has_many")
				}

				if v["count"] != nil {
					counts := []map[string]any{}

					if err := DB.Table(v["table"].(string)).Select(fk, "COUNT("+fk+") as count").Where(fk+" in ?", ids).Group(fk).Find(&counts).Error; err != nil {
						fmt.Println(err)
					}

					for _, result := range results {
						count := filterSliceByMapIndex(counts, fk, result["id"])
						result[i+"_count"] = 0

						if len(count) != 0 {
							countz := count[0].(map[string]any)
							result[i+"_count"] = countz["count"]
							delete(result, "has_many")
						}
					}
				}
			}
		}
	}
}

func MultiAttachManyToMany(results []map[string]any, ctx *gin.Context) {
	ids := []string{}

	for _, result := range results {
		if result["id"] != nil {
			ids = append(ids, strconv.Itoa(ConvertToInt(result["id"])))
		}
	}

	if len(results) > 0 {
		transformer := results[0]

		if transformer["many_to_many"] != nil {
			for i, v := range transformer["many_to_many"].(map[string]any) {
				v := v.(map[string]any)
				fk1 := v["fk1"].(string)
				fk2 := v["fk2"].(string)
				colums := convertAnyToString(v["columns"].([]any))
				colums = append(colums, fk1)
				colums = append(colums, fk2)
				values := []map[string]any{}

				if limit := ConvertToInt(v["limit"]); limit > 0 {
					subSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
						return tx.
							Select("*", "ROW_NUMBER() OVER (PARTITION BY "+fk1+" ORDER BY id) AS rn").
							Table(v["table"].(string)).
							Where(fk1, ids).
							Find(&[]any{})
					})

					query := DB.Table("(" + subSql + ") AS subQ")

					for i := range colums {
						colums[i] = "subQ." + colums[i]
					}

					SetBelongsTo(query, v, &colums, ctx)

					if err := query.Select(colums).Where("rn <= " + strconv.Itoa(limit)).Find(&values).Error; err != nil {
						fmt.Println(err)
					}

					// todo : need to fix, return null if belong to not exist
					values = MultiMapValuesShifter2(v, values)
				} else {
					if err := DB.Table(v["table"].(string)).Select(colums).Where(fk1+" in ?", ids).Find(&values).Error; err != nil {
						fmt.Println(err)
					}
				}

				for _, result := range results {
					result[i] = filterSliceByMapIndex(values, fk1, result["id"])
					delete(result, "many_to_many")
				}

				if v["count"] != nil {
					counts := []map[string]any{}

					if err := DB.Table(v["table"].(string)).Select(fk1, "COUNT("+fk1+") as count").Where(fk1+" in ?", ids).Group(fk1).Find(&counts).Error; err != nil {
						fmt.Println(err)
					}

					for _, result := range results {
						count := filterSliceByMapIndex(counts, fk1, result["id"])
						result[i+"_count"] = 0

						if len(count) != 0 {
							countz := count[0].(map[string]any)
							result[i+"_count"] = countz["count"]
							delete(result, "many_to_many")
						}
					}
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

			//need better variable name
			if v["belongs_to"] != nil {
				for name2, v2 := range v["belongs_to"].(map[string]any) {
					v2 := v2.(map[string]any)
					values2 := map[string]any{}

					for _, val2 := range v2["columns"].([]any) {
						values2[val2.(string)] = value[name+"_"+name2+"_"+val2.(string)]
						//delete(transformer, v["fk"].(string))
					}

					t := transformer[name].(map[string]any)
					t[name2] = values2
				}
			}
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
			var total int = 0

			for _, v := range values {
				val := v[s["total"].(string)]
				total += ConvertToInt(val)
				delete(v, "summary")
			}

			summary["total"] = total
		}
	}

	return summary
}
