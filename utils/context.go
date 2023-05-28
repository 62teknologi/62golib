package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ParseForm(ctx *gin.Context) map[string]any {
	input := map[string]any{}
	contentType := ctx.GetHeader("Content-Type")

	fmt.Println(contentType)

	if contentType == "application/json" {
		if err := ctx.BindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, ResponseData("error", err.Error(), nil))
		}
	}

	if contentType == "application/x-www-form-urlencoded" {
		if err := ctx.Request.ParseForm(); err != nil {
			ctx.JSON(http.StatusBadRequest, ResponseData("error", err.Error(), nil))
		}

		input = make(map[string]interface{})

		for key, values := range ctx.Request.PostForm {
			input[key] = values[0]
		}
	}

	return input
}
