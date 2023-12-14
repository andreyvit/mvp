package mvputil

import (
	"regexp"
)

const (
	maxDomainLength            = 253
	maxEmailLength             = 254
	domainPartRegexStr         = `[a-zA-Z0-9-]{1,63}`
	topLevelDomainPartRegexStr = `(?:[xX][nN]--[a-zA-Z0-9]{1,59}|[a-zA-Z]{2,63})`
	domainRegexStr             = `(?:` + domainPartRegexStr + `\.)+` + topLevelDomainPartRegexStr
	emailLocalPartRegexStr     = `[a-zA-Z0-9._%!+-]{1,64}`
)

var (
	emailRegex  = regexp.MustCompile(`^` + emailLocalPartRegexStr + `@` + domainRegexStr + `$`)
	domainRegex = regexp.MustCompile(`^` + domainRegexStr + `$`)
)

func IsValidEmail(email string) bool {
	return len(email) <= maxEmailLength && emailRegex.MatchString(email)
}

func IsValidDomain(domain string) bool {
	return len(domain) <= maxDomainLength && domainRegex.MatchString(domain)
}
