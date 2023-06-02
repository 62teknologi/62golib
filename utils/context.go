package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// todo: need to handle form data
func ParseForm(ctx *gin.Context) map[string]any {
	input := map[string]any{}
	contentType := ctx.GetHeader("Content-Type")

	if contentType == "application/json" {
		if err := ctx.BindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, ResponseData("error", err.Error(), nil))
		}
	}

	if contentType == "application/x-www-form-urlencoded" {
		if err := ctx.Request.ParseForm(); err != nil {
			ctx.JSON(http.StatusBadRequest, ResponseData("error", err.Error(), nil))
		}

		input = make(map[string]any)

		for name, values := range ctx.Request.PostForm {
			if !strings.Contains(name, "[]") {
				input[name] = values[0]
			} else {
				// looks complex, but it is needed for make sure the type as any
				newVals := make([]any, len(values))

				for i, s := range values {
					newVals[i] = s
				}

				name = strings.Replace(name, "[]", "", -1)
				input[name] = newVals
			}
		}
	}

	return input
}
