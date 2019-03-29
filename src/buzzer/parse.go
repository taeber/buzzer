package buzzer

import (
	"regexp"
	"strings"
)

var (
	usernames = regexp.MustCompile(`(^|\W)@(\w+)`)
	topics    = regexp.MustCompile(`(^|\W)#(\w+)`)
)

func parseMentions(msg string) (found []string) {
	for _, sub := range usernames.FindAllStringSubmatch(msg, -1) {
		found = append(found, sub[2])
	}
	return found
}

func parseTags(msg string) (found []string) {
	for _, sub := range topics.FindAllStringSubmatch(msg, -1) {
		found = append(found, strings.ToLower(sub[2]))
	}
	return found
}
