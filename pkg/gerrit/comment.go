package gerrit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) GetComments(ctx context.Context, changeID string) (map[string][]CommentInfo, error) {
	path := fmt.Sprintf("changes/%s/comments", url.PathEscape(changeID))
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var comments map[string][]CommentInfo
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, fmt.Errorf("parse comments: %w", err)
	}
	return comments, nil
}

func (c *Client) GetDrafts(ctx context.Context, changeID, revision string) (map[string][]CommentInfo, error) {
	revision = resolveRevision(revision)
	path := fmt.Sprintf("changes/%s/revisions/%s/drafts", url.PathEscape(changeID), revision)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var drafts map[string][]CommentInfo
	if err := json.Unmarshal(data, &drafts); err != nil {
		return nil, fmt.Errorf("parse drafts: %w", err)
	}
	return drafts, nil
}

func (c *Client) CreateDraft(ctx context.Context, changeID, revision string, input *DraftInput) (*CommentInfo, error) {
	revision = resolveRevision(revision)
	path := fmt.Sprintf("changes/%s/revisions/%s/drafts", url.PathEscape(changeID), revision)
	data, err := c.Put(ctx, path, input)
	if err != nil {
		return nil, err
	}
	var comment CommentInfo
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("parse draft: %w", err)
	}
	return &comment, nil
}

func (c *Client) DeleteDraft(ctx context.Context, changeID, revision, draftID string) error {
	revision = resolveRevision(revision)
	path := fmt.Sprintf("changes/%s/revisions/%s/drafts/%s", url.PathEscape(changeID), revision, draftID)
	return c.Delete(ctx, path)
}
