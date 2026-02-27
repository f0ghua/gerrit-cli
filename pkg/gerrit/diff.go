package gerrit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) GetFileDiff(ctx context.Context, changeID, revision, file string, base int) (*DiffInfo, error) {
	revision = resolveRevision(revision)
	path := fmt.Sprintf("changes/%s/revisions/%s/files/%s/diff",
		url.PathEscape(changeID), revision, url.PathEscape(file))
	if base > 0 {
		path += fmt.Sprintf("?base=%d", base)
	}
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var diff DiffInfo
	if err := json.Unmarshal(data, &diff); err != nil {
		return nil, fmt.Errorf("parse diff: %w", err)
	}
	return &diff, nil
}
