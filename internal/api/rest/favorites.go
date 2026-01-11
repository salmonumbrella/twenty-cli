package rest

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func (c *Client) ListFavorites(ctx context.Context, opts *ListOptions) (*types.ListResponse[types.Favorite], error) {
	path := "/rest/favorites"
	path, err := applyListParams(path, opts)
	if err != nil {
		return nil, err
	}

	var apiResp types.FavoritesListResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}

	result := &types.ListResponse[types.Favorite]{
		Data:       apiResp.Data.Favorites,
		TotalCount: apiResp.TotalCount,
		PageInfo:   apiResp.PageInfo,
	}
	return result, nil
}

func (c *Client) GetFavorite(ctx context.Context, id string) (*types.Favorite, error) {
	path := fmt.Sprintf("/rest/favorites/%s", id)
	var apiResp types.FavoriteResponse
	if err := c.Get(ctx, path, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp.Data.Favorite, nil
}
