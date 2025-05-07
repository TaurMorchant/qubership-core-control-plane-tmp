package format

import (
	"regexp"
)

type UrlValidator struct {
	pathPattern *regexp.Regexp
}

var DefaultUrlValidator = NewUrlValidator("\\s+")

func NewUrlValidator(pathPattern string) *UrlValidator {
	pathRegexp, err := regexp.Compile(pathPattern)
	if err != nil {
		logger.Panicf("Failed to build new UrlValidator due to pathRegexp pattern compilation error: %v", err)
	}
	return &UrlValidator{pathPattern: pathRegexp}
}

func (validator *UrlValidator) IsValidPath(path string) bool {
	return path != "" && !validator.pathPattern.MatchString(path)
}
