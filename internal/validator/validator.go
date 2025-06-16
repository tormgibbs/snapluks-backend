package validator

import (
	"net/url"
	"regexp"
	"slices"
)

var (
	EmailRX = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+\/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	PhoneRX = regexp.MustCompile(`^[0-9]{10}$`)
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, msg string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = msg
	}
}

func (v *Validator) Check(ok bool, key string, msg string) {
	if !ok {
		v.AddError(key, msg)
	}
}

func In(value string, list ...string) bool {
	return slices.Contains(list, value)
}

func IsURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}
