package gerrit

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
)

func (c *Client) GetPatch(ctx context.Context, changeID, revision string) ([]byte, error) {
	revision = resolveRevision(revision)
	path := fmt.Sprintf("changes/%s/revisions/%s/patch", url.PathEscape(changeID), revision)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return data, nil
	}
	return decoded, nil
}
