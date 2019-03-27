package buzzer

import "regexp"

var mentions = regexp.MustCompile(`(^|\W)@(\w+)`)

func parseMentions(msg string) (found []string) {
	for _, sub := range mentions.FindAllStringSubmatch(msg, -1) {
		found = append(found, sub[2])
	}
	return found
}
