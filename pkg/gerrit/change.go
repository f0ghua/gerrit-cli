package gerrit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) ListChanges(ctx context.Context, query string, limit int, opts ...string) ([]ChangeInfo, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("n", fmt.Sprintf("%d", limit))
	for _, o := range opts {
		params.Add("o", o)
	}
	if len(opts) == 0 {
		for _, o := range []string{"LABELS", "CURRENT_REVISION", "DETAILED_ACCOUNTS"} {
			params.Add("o", o)
		}
	}
	data, err := c.Get(ctx, "changes/?"+params.Encode())
	if err != nil {
		return nil, err
	}
	var changes []ChangeInfo
	if err := json.Unmarshal(data, &changes); err != nil {
		return nil, fmt.Errorf("parse changes: %w", err)
	}
	return changes, nil
}

func (c *Client) GetChange(ctx context.Context, changeID string, opts ...string) (*ChangeInfo, error) {
	params := url.Values{}
	if len(opts) == 0 {
		for _, o := range []string{"LABELS", "CURRENT_REVISION", "CURRENT_COMMIT", "DETAILED_ACCOUNTS", "DETAILED_LABELS"} {
			params.Add("o", o)
		}
	} else {
		for _, o := range opts {
			params.Add("o", o)
		}
	}
	path := fmt.Sprintf("changes/%s?%s", url.PathEscape(changeID), params.Encode())
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var change ChangeInfo
	if err := json.Unmarshal(data, &change); err != nil {
		return nil, fmt.Errorf("parse change: %w", err)
	}
	return &change, nil
}

func (c *Client) GetChangeFiles(ctx context.Context, changeID, revision string) (map[string]FileInfo, error) {
	if revision == "" {
		revision = "current"
	}
	path := fmt.Sprintf("changes/%s/revisions/%s/files", url.PathEscape(changeID), revision)
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var files map[string]FileInfo
	if err := json.Unmarshal(data, &files); err != nil {
		return nil, fmt.Errorf("parse files: %w", err)
	}
	return files, nil
}

func (c *Client) GetChangeMessages(ctx context.Context, changeID string) ([]ChangeMessageInfo, error) {
	path := fmt.Sprintf("changes/%s/messages", url.PathEscape(changeID))
	data, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var msgs []ChangeMessageInfo
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, fmt.Errorf("parse messages: %w", err)
	}
	return msgs, nil
}

func (c *Client) Submit(ctx context.Context, changeID string) (*ChangeInfo, error) {
	data, err := c.Post(ctx, fmt.Sprintf("changes/%s/submit", url.PathEscape(changeID)), &SubmitInput{})
	if err != nil {
		return nil, err
	}
	var change ChangeInfo
	if err := json.Unmarshal(data, &change); err != nil {
		return nil, fmt.Errorf("parse submit response: %w", err)
	}
	return &change, nil
}

func (c *Client) Abandon(ctx context.Context, changeID string, input *AbandonInput) (*ChangeInfo, error) {
	data, err := c.Post(ctx, fmt.Sprintf("changes/%s/abandon", url.PathEscape(changeID)), input)
	if err != nil {
		return nil, err
	}
	var change ChangeInfo
	if err := json.Unmarshal(data, &change); err != nil {
		return nil, fmt.Errorf("parse abandon response: %w", err)
	}
	return &change, nil
}

func (c *Client) Restore(ctx context.Context, changeID string) (*ChangeInfo, error) {
	data, err := c.Post(ctx, fmt.Sprintf("changes/%s/restore", url.PathEscape(changeID)), struct{}{})
	if err != nil {
		return nil, err
	}
	var change ChangeInfo
	if err := json.Unmarshal(data, &change); err != nil {
		return nil, fmt.Errorf("parse restore response: %w", err)
	}
	return &change, nil
}

func (c *Client) Rebase(ctx context.Context, changeID string, input *RebaseInput) (*ChangeInfo, error) {
	data, err := c.Post(ctx, fmt.Sprintf("changes/%s/rebase", url.PathEscape(changeID)), input)
	if err != nil {
		return nil, err
	}
	var change ChangeInfo
	if err := json.Unmarshal(data, &change); err != nil {
		return nil, fmt.Errorf("parse rebase response: %w", err)
	}
	return &change, nil
}

// resolveRevision converts a revision string, defaulting empty to "current".
func resolveRevision(revision string) string {
	if revision == "" {
		return "current"
	}
	return strings.TrimSpace(revision)
}
