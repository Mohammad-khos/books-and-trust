package domain

import "regexp"

var (
	uppercase = regexp.MustCompile(`[A-Z]`)
	number    = regexp.MustCompile(`[0-9]`)
	special   = regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)
)

func (p *Password) IsMatched(text string) bool {
	return len(text) >= 8 &&
		uppercase.MatchString(text) &&
		number.MatchString(text) &&
		special.MatchString(text)
}
