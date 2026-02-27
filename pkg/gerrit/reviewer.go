package gerrit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) AddReviewer(ctx context.Context, changeID string, input *ReviewerInput) (*AddReviewerResult, error) {
	path := fmt.Sprintf("changes/%s/reviewers", url.PathEscape(changeID))
	data, err := c.Post(ctx, path, input)
	if err != nil {
		return nil, err
	}
	var result AddReviewerResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse add reviewer: %w", err)
	}
	return &result, nil
}

func (c *Client) RemoveReviewer(ctx context.Context, changeID, accountID string) error {
	path := fmt.Sprintf("changes/%s/reviewers/%s", url.PathEscape(changeID), url.PathEscape(accountID))
	return c.Delete(ctx, path)
}
