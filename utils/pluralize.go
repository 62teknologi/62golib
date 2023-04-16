package utils

import pluralize "github.com/gertd/go-pluralize"

var Pluralize *pluralize.Client
var SingularName string
var PluralName string

func InitPluralize() {
	Pluralize = pluralize.NewClient()
}

func SetPluralizeNames(name string) {
	SingularName = Pluralize.Singular(name)
	PluralName = Pluralize.Plural(name)
}
