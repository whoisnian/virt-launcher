package image

import (
	"context"
	"net/http"

	"github.com/whoisnian/virt-launcher/global"
)

func requestGet(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	// net/http default user-agent "Go-http-client/1.1" maybe blocked by some websites.
	req.Header.Set("User-Agent", global.AppName+"/"+global.Version)
	return http.DefaultClient.Do(req)
}
