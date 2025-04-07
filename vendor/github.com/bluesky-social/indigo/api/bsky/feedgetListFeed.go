// Code generated by cmd/lexgen (see Makefile's lexgen); DO NOT EDIT.

package bsky

// schema: app.bsky.feed.getListFeed

import (
	"context"

	"github.com/bluesky-social/indigo/xrpc"
)

// FeedGetListFeed_Output is the output of a app.bsky.feed.getListFeed call.
type FeedGetListFeed_Output struct {
	Cursor *string                  `json:"cursor,omitempty" cborgen:"cursor,omitempty"`
	Feed   []*FeedDefs_FeedViewPost `json:"feed" cborgen:"feed"`
}

// FeedGetListFeed calls the XRPC method "app.bsky.feed.getListFeed".
//
// list: Reference (AT-URI) to the list record.
func FeedGetListFeed(ctx context.Context, c *xrpc.Client, cursor string, limit int64, list string) (*FeedGetListFeed_Output, error) {
	var out FeedGetListFeed_Output

	params := map[string]interface{}{
		"cursor": cursor,
		"limit":  limit,
		"list":   list,
	}
	if err := c.Do(ctx, xrpc.Query, "", "app.bsky.feed.getListFeed", params, nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}
