package sources

import (
	"context"
	"errors"
	"testing"

	"github.com/overmindtech/sdp-go"
)

func TestMaxParallel(t *testing.T) {
	var p MaxParallel

	if p.Value() != 10 {
		t.Errorf("expected max parallel to be 10, got %v", p)
	}
}

func TestListGetSourceType(t *testing.T) {
	lgs := ListGetSource[any, any, any, any]{
		ItemType: "foo",
	}

	if lgs.Type() != "foo" {
		t.Errorf("expected type to be foo, got %v", lgs.Type())
	}
}

func TestListGetSourceName(t *testing.T) {
	lgs := ListGetSource[any, any, any, any]{
		ItemType: "foo",
	}

	if lgs.Name() != "foo-source" {
		t.Errorf("expected name to be foo-source, got %v", lgs.Name())
	}
}

func TestListGetSourceScopes(t *testing.T) {
	lgs := ListGetSource[any, any, any, any]{
		AccountID: "foo",
		Region:    "bar",
	}

	if lgs.Scopes()[0] != "foo.bar" {
		t.Errorf("expected scope to be foo.bar, got %v", lgs.Scopes()[0])
	}
}

func TestListGetSourceGet(t *testing.T) {
	t.Run("with no errors", func(t *testing.T) {
		lgs := ListGetSource[string, string, struct{}, struct{}]{
			ItemType:  "test",
			AccountID: "foo",
			Region:    "bar",
			Client:    struct{}{},
			ListInput: "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{}
			},
			ListFuncOutputMapper: func(output string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
		}

		_, err := lgs.Get(context.Background(), "foo.bar", "")

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("with an error", func(t *testing.T) {
		lgs := ListGetSource[string, string, struct{}, struct{}]{
			ItemType:  "test",
			AccountID: "foo",
			Region:    "bar",
			Client:    struct{}{},
			ListInput: "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{}
			},
			ListFuncOutputMapper: func(output string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (*sdp.Item, error) {
				return &sdp.Item{}, errors.New("foo")
			},
		}

		_, err := lgs.Get(context.Background(), "foo.bar", "")

		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestListGetSourceList(t *testing.T) {
	t.Run("with no errors", func(t *testing.T) {
		lgs := ListGetSource[string, string, struct{}, struct{}]{
			ItemType:    "test",
			AccountID:   "foo",
			Region:      "bar",
			Client:      struct{}{},
			MaxParallel: MaxParallel(1),
			ListInput:   "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{}
			},
			ListFuncOutputMapper: func(output string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
		}

		items, err := lgs.List(context.Background(), "foo.bar")

		if err != nil {
			t.Error(err)
		}

		if len(items) != 6 {
			t.Errorf("expected 6 results, got %v", len(items))
		}
	})

	t.Run("with a failing output mapper", func(t *testing.T) {
		lgs := ListGetSource[string, string, struct{}, struct{}]{
			ItemType:    "test",
			AccountID:   "foo",
			Region:      "bar",
			Client:      struct{}{},
			MaxParallel: MaxParallel(1),
			ListInput:   "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{}
			},
			ListFuncOutputMapper: func(output string) ([]string, error) {
				// Returns 2 gets per page
				return nil, errors.New("output mapper error")
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
		}

		_, err := lgs.List(context.Background(), "foo.bar")

		if err == nil {
			t.Fatal("expected error but got nil")
		}

		if err.Error() != "output mapper error" {
			t.Errorf("expected output mapper error, got %v", err.Error())
		}
	})

	t.Run("with a failing GetFunc", func(t *testing.T) {
		lgs := ListGetSource[string, string, struct{}, struct{}]{
			ItemType:    "test",
			AccountID:   "foo",
			Region:      "bar",
			Client:      struct{}{},
			MaxParallel: MaxParallel(1),
			ListInput:   "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{}
			},
			ListFuncOutputMapper: func(output string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (*sdp.Item, error) {
				return &sdp.Item{}, errors.New("get func error")
			},
		}

		items, err := lgs.List(context.Background(), "foo.bar")

		// If GetFunc fails it doesn't cause an error
		if err != nil {
			t.Error(err)
		}

		if len(items) != 0 {
			t.Errorf("expected no items, got %v", len(items))
		}
	})
}
