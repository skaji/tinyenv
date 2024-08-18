package installer

import (
	"context"
	"regexp"
	"strings"
)

type GitHub struct{}

func (g *GitHub) Tags(ctx context.Context, url string) ([]string, error) {
	b, err := HTTPGet(ctx, url+"/releases")
	if err != nil {
		return nil, err
	}
	matches := regexp.MustCompile(`href="(.+?)"`).FindAllStringSubmatch(string(b), -1)
	var out []string
	for _, match := range matches {
		href := match[1]
		m := regexp.MustCompile(`/releases/tag/([^/]+)`).FindStringSubmatch(href)
		if m != nil {
			out = append(out, m[1])
		}
	}
	return out, nil
}

func (g *GitHub) Assets(ctx context.Context, url string, tag string) ([]string, error) {
	b, err := HTTPGet(ctx, url+"/releases/expanded_assets/"+tag)
	if err != nil {
		return nil, err
	}
	matches := regexp.MustCompile(`href="(.+?)"`).FindAllStringSubmatch(string(b), -1)
	var out []string
	for _, match := range matches {
		href := match[1]
		if strings.Contains(href, "/releases/download/") {
			if strings.HasPrefix(href, "https") {
				out = append(out, href)
			} else {
				out = append(out, "https://github.com"+href)
			}
		}
	}
	return out, nil
}
