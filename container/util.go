package container

import (
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	re2Prefix = "re2:" // Re2Prefix re2 regexp string prefix
)

func splitAndTrimSpaces(val string, sep string) []string {
	vals := []string{}
	list := strings.Split(val, sep)
	for _, s := range list {
		vals = append(vals, strings.TrimSpace(s))
	}
	return vals
}

func sliceContains(val string, s []string) bool {
	ret := false
	for _, curr := range s {
		if curr == val {
			ret = true
			break
		}
	}
	return ret
}

func inFilterOrList(val string, filter string) bool {
	// check if filter is a RE2 regexp
	if strings.HasPrefix(filter, re2Prefix) {
		pattern := strings.Trim(filter, re2Prefix)
		log.Debugf("Using RE2 pattern: '%s'", pattern)
		var matched bool
		var err error
		if matched, err = regexp.MatchString(pattern, val); err != nil {
			log.Error(err)
		}
		return matched
	}
	// check if value in list
	return sliceContains(val, splitAndTrimSpaces(filter, ","))
}

func mapContains(m map[string]string, kv []string) bool {
	if len(kv) == 2 {
		if val, ok := m[kv[0]]; ok {
			return val == kv[1]
		}
	} else if len(kv) == 1 {
		if _, ok := m[kv[0]]; ok {
			return true
		}
	}
	return false
}
