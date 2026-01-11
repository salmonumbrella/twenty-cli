package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestClient_ListFavorites(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	companyID := "company-1"
	personID := "person-1"
	expectedFavorites := []types.Favorite{
		{
			ID:                "favorite-1",
			Position:          1.0,
			WorkspaceMemberID: "member-1",
			CompanyID:         &companyID,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "favorite-2",
			Position:          2.0,
			WorkspaceMemberID: "member-1",
			PersonID:          &personID,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/favorites" {
			t.Errorf("expected path /rest/favorites, got %s", r.URL.Path)
		}

		resp := types.FavoritesListResponse{
			TotalCount: 2,
			PageInfo:   &types.PageInfo{HasNextPage: false, EndCursor: "cursor-2"},
		}
		resp.Data.Favorites = expectedFavorites

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	result, err := client.ListFavorites(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected 2 favorites, got %d", len(result.Data))
	}
	if result.Data[0].ID != "favorite-1" {
		t.Errorf("expected first favorite ID 'favorite-1', got %s", result.Data[0].ID)
	}
	if result.Data[0].Position != 1.0 {
		t.Errorf("expected position 1.0, got %f", result.Data[0].Position)
	}
	if result.Data[0].WorkspaceMemberID != "member-1" {
		t.Errorf("expected workspaceMemberId 'member-1', got %s", result.Data[0].WorkspaceMemberID)
	}
	if result.Data[0].CompanyID == nil || *result.Data[0].CompanyID != "company-1" {
		t.Errorf("expected companyId 'company-1', got %v", result.Data[0].CompanyID)
	}
	if result.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if result.PageInfo.EndCursor != "cursor-2" {
		t.Errorf("expected EndCursor 'cursor-2', got %s", result.PageInfo.EndCursor)
	}
}

func TestClient_ListFavorites_WithOptions(t *testing.T) {
	tests := []struct {
		name           string
		opts           *ListOptions
		expectedParams map[string]string
	}{
		{
			name: "with limit",
			opts: &ListOptions{Limit: 10},
			expectedParams: map[string]string{
				"limit": "10",
			},
		},
		{
			name: "with cursor",
			opts: &ListOptions{Cursor: "abc123"},
			expectedParams: map[string]string{
				"starting_after": "abc123",
			},
		},
		{
			name: "combined options",
			opts: &ListOptions{
				Limit:  25,
				Cursor: "xyz789",
				Sort:   "position",
				Order:  "asc",
			},
			expectedParams: map[string]string{
				"limit":              "25",
				"starting_after":     "xyz789",
				"order_by":           "position",
				"order_by_direction": "asc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedQuery = r.URL.RawQuery
				resp := types.FavoritesListResponse{TotalCount: 0}
				resp.Data.Favorites = []types.Favorite{}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", false, WithNoRetry())
			_, err := client.ListFavorites(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for param, expected := range tt.expectedParams {
				if !strings.Contains(receivedQuery, param+"="+expected) {
					t.Errorf("expected query to contain %s=%s, got query: %s", param, expected, receivedQuery)
				}
			}
		})
	}
}

func TestClient_ListFavorites_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Unauthorized"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.ListFavorites(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error for unauthorized request, got nil")
	}
}

func TestClient_GetFavorite(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	opportunityID := "opportunity-123"
	expectedFavorite := types.Favorite{
		ID:                "favorite-456",
		Position:          3.5,
		WorkspaceMemberID: "member-2",
		OpportunityID:     &opportunityID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	t.Run("basic get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/rest/favorites/favorite-456" {
				t.Errorf("expected path /rest/favorites/favorite-456, got %s", r.URL.Path)
			}

			resp := types.FavoriteResponse{}
			resp.Data.Favorite = expectedFavorite
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		favorite, err := client.GetFavorite(context.Background(), "favorite-456")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if favorite.ID != "favorite-456" {
			t.Errorf("expected ID 'favorite-456', got %s", favorite.ID)
		}
		if favorite.Position != 3.5 {
			t.Errorf("expected position 3.5, got %f", favorite.Position)
		}
		if favorite.WorkspaceMemberID != "member-2" {
			t.Errorf("expected workspaceMemberId 'member-2', got %s", favorite.WorkspaceMemberID)
		}
		if favorite.OpportunityID == nil || *favorite.OpportunityID != "opportunity-123" {
			t.Errorf("expected opportunityId 'opportunity-123', got %v", favorite.OpportunityID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"Favorite not found"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		_, err := client.GetFavorite(context.Background(), "non-existent")

		if err == nil {
			t.Fatal("expected error for non-existent favorite, got nil")
		}
	})
}
