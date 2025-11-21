package utils

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func ToTitleCase(s string) string {
	c := cases.Title(language.Indonesian)
	return c.String(s)
}
