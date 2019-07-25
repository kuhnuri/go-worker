package jar

import (
	"testing"
)

func TestParse(t *testing.T) {
	parse(t, "jar:foo!/", "foo", "")
}
func TestParseWithEntry(t *testing.T) {
	parse(t, "jar:foo!/bar", "foo", "bar")
}
func TestParseMissingSeparator(t *testing.T) {
	_, err := Parse("jar:foo")
	if err == nil {
		t.Fatalf("Failed to parse. Expected error")
	}
}
func TestParseWrongScheme(t *testing.T) {
	_, err := Parse("file:foo")
	if err == nil {
		t.Fatalf("Failed to parse. Expected error")
	}
}

func parse(t *testing.T, input string, url string, entry string) {
	act, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %s", err.Error())
	}
	if act.Url.String() != url {
		t.Fatalf("Failed to parse. Expected %s, actual %s", url, act.Url.String())
	}
	if act.Entry != entry {
		t.Fatalf("Failed to parse. Expected %s, actual %s", entry, act.Entry)
	}
}