package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) ListDNSRecords(ctx context.Context, domain, teamID string) (r []DNSRecord, err error) {
	url := fmt.Sprintf("%s/v4/domains/%s/records?limit=100", c.baseURL, domain)
	if teamID != "" {
		url = fmt.Sprintf("%s&teamId=%s", url, teamID)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		url,
		strings.NewReader(""),
	)
	if err != nil {
		return r, err
	}

	dr := struct {
		Records []DNSRecord `json:"records"`
	}{}
	err = c.doRequest(req, &dr)
	return dr.Records, err
}
