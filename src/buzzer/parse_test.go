package buzzer

import (
	"sort"
	"testing"
)

func TestParseMentions(t *testing.T) {
	examples := map[string][]string{
		"Hello @taeber how are you?": {"taeber"},
		"@user42, are you there?":    {"user42"},
		"Email me taeber@email.com":  {},
		"":                           {},
		"@bob,@ross! Wow.":           {"bob", "ross"},
	}

	for msg, expected := range examples {
		actual := parseMentions(msg)
		if len(expected) != len(actual) {
			t.Errorf("parseMentions failed\nexpected = %v\ngot = %v\n", expected, actual)
		}
		sort.Strings(actual)
		sort.Strings(expected)
		for i := 0; i < len(expected); i++ {
			if expected[i] != actual[i] {
				t.Errorf("parseMentions failed\nexpected = %v\ngot = %v\n", expected[i], actual[i])
			}
		}
	}
}
