package gerrit

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) SetReview(ctx context.Context, changeID, revision string, input *ReviewInput) error {
	revision = resolveRevision(revision)
	path := fmt.Sprintf("changes/%s/revisions/%s/review", url.PathEscape(changeID), revision)
	_, err := c.Post(ctx, path, input)
	return err
}
