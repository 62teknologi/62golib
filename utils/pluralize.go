package utils

import pluralize "github.com/gertd/go-pluralize"

var Pluralize *pluralize.Client

func InitPluralize() {
	Pluralize = pluralize.NewClient()
}
