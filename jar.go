package kuhnuri

import (
	"fmt"
	"net/url"
	"strings"
)

type JarUri struct {
	Url   url.URL
	Entry string
}

func Parse(u string) (*JarUri, error) {
	i := strings.Index(u, "!/")
	if len(u) < 6 || u[0:4] != "jar:" || i == -1 {
		return nil, fmt.Errorf("Failed to parse %s", u)
	}
	uri, err := url.Parse(u[4:i])
	if err != nil {
		return nil, fmt.Errorf("Failed to parse %s: %v", u, err)
	}
	return &JarUri{Url: *uri, Entry: u[i+2:]}, nil
}

func (j *JarUri) String() string {
	return fmt.Sprintf("jar:%s!/%s", j.Url.String(), j.Entry)
}
