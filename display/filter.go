package display

import "regexp"

type Filter func(message string, pattern string) bool

func SimpleFilter(message, pattern string) bool {
	rx := regexp.MustCompile(pattern)
	return rx.MatchString(message)
}
