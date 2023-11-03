package sources

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/overmindtech/sdp-go"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMaxParallel(t *testing.T) {
	var p MaxParallel

	if p.Value() != 10 {
		t.Errorf("expected max parallel to be 10, got %v", p)
	}
}

func TestAlwaysGetSourceType(t *testing.T) {
	lgs := AlwaysGetSource[any, any, any, any, any, any]{
		ItemType: "foo",
	}

	if lgs.Type() != "foo" {
		t.Errorf("expected type to be foo, got %v", lgs.Type())
	}
}

func TestAlwaysGetSourceName(t *testing.T) {
	lgs := AlwaysGetSource[any, any, any, any, any, any]{
		ItemType: "foo",
	}

	if lgs.Name() != "foo-source" {
		t.Errorf("expected name to be foo-source, got %v", lgs.Name())
	}
}

func TestAlwaysGetSourceScopes(t *testing.T) {
	lgs := AlwaysGetSource[any, any, any, any, any, any]{
		AccountID: "foo",
		Region:    "bar",
	}

	if lgs.Scopes()[0] != "foo.bar" {
		t.Errorf("expected scope to be foo.bar, got %v", lgs.Scopes()[0])
	}
}

func TestAlwaysGetSourceGet(t *testing.T) {
	t.Run("with no errors", func(t *testing.T) {
		lgs := AlwaysGetSource[string, string, string, string, struct{}, struct{}]{
			ItemType:  "test",
			AccountID: "foo",
			Region:    "bar",
			Client:    struct{}{},
			ListInput: "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{DataFunc: func() string {
					return "foo"
				}}
			},
			ListFuncOutputMapper: func(output, input string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, input string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
			GetInputMapper: func(scope, query string) string {
				return ""
			},
		}

		_, err := lgs.Get(context.Background(), "foo.bar", "", false)

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("with an error", func(t *testing.T) {
		lgs := AlwaysGetSource[string, string, string, string, struct{}, struct{}]{
			ItemType:  "test",
			AccountID: "foo",
			Region:    "bar",
			Client:    struct{}{},
			ListInput: "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{DataFunc: func() string {
					return "foo"
				}}
			},
			ListFuncOutputMapper: func(output, input string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, input string) (*sdp.Item, error) {
				return &sdp.Item{}, errors.New("foo")
			},
			GetInputMapper: func(scope, query string) string {
				return ""
			},
		}

		_, err := lgs.Get(context.Background(), "foo.bar", "", false)

		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestAlwaysGetSourceList(t *testing.T) {
	t.Run("with no errors", func(t *testing.T) {
		lgs := AlwaysGetSource[string, string, string, string, struct{}, struct{}]{
			ItemType:    "test",
			AccountID:   "foo",
			Region:      "bar",
			Client:      struct{}{},
			MaxParallel: MaxParallel(1),
			ListInput:   "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{DataFunc: func() string {
					return "foo"
				}}
			},
			ListFuncOutputMapper: func(output, input string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, input string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
			GetInputMapper: func(scope, query string) string {
				return ""
			},
		}

		items, err := lgs.List(context.Background(), "foo.bar", false)

		if err != nil {
			t.Error(err)
		}

		if len(items) != 6 {
			t.Errorf("expected 6 results, got %v", len(items))
		}
	})

	t.Run("with a failing output mapper", func(t *testing.T) {
		lgs := AlwaysGetSource[string, string, string, string, struct{}, struct{}]{
			ItemType:    "test",
			AccountID:   "foo",
			Region:      "bar",
			Client:      struct{}{},
			MaxParallel: MaxParallel(1),
			ListInput:   "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{DataFunc: func() string {
					return "foo"
				}}
			},
			ListFuncOutputMapper: func(output, input string) ([]string, error) {
				// Returns 2 gets per page
				return nil, errors.New("output mapper error")
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, input string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
			GetInputMapper: func(scope, query string) string {
				return ""
			},
		}

		_, err := lgs.List(context.Background(), "foo.bar", false)

		if err == nil {
			t.Fatal("expected error but got nil")
		}

		qErr := &sdp.QueryError{}
		if !errors.As(err, &qErr) {
			t.Errorf("expected error to be a QueryError, got %v", err)
		} else {
			if qErr.ErrorString != "output mapper error" {
				t.Errorf("expected 'output mapper error', got '%v'", qErr.ErrorString)
			}
		}
	})

	t.Run("with a failing GetFunc", func(t *testing.T) {
		lgs := AlwaysGetSource[string, string, string, string, struct{}, struct{}]{
			ItemType:    "test",
			AccountID:   "foo",
			Region:      "bar",
			Client:      struct{}{},
			MaxParallel: MaxParallel(1),
			ListInput:   "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{DataFunc: func() string {
					return "foo"
				}}
			},
			ListFuncOutputMapper: func(output, input string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, input string) (*sdp.Item, error) {
				return &sdp.Item{}, errors.New("get func error")
			},
			GetInputMapper: func(scope, query string) string {
				return ""
			},
		}

		items, err := lgs.List(context.Background(), "foo.bar", false)

		// If GetFunc fails it doesn't cause an error
		if err != nil {
			t.Error(err)
		}

		if len(items) != 0 {
			t.Errorf("expected no items, got %v", len(items))
		}
	})
}

func TestAlwaysGetSourceSearch(t *testing.T) {
	t.Run("with ARN search", func(t *testing.T) {
		lgs := AlwaysGetSource[string, string, string, string, struct{}, struct{}]{
			ItemType:    "test",
			AccountID:   "foo",
			Region:      "bar",
			Client:      struct{}{},
			MaxParallel: MaxParallel(1),
			ListInput:   "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{DataFunc: func() string {
					return "foo"
				}}
			},
			ListFuncOutputMapper: func(output, input string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, input string) (*sdp.Item, error) {
				if input == "foo.bar.id" {
					return &sdp.Item{}, nil
				} else {
					return nil, sdp.NewQueryError(errors.New("bad query details"))
				}
			},
			GetInputMapper: func(scope, query string) string {
				return scope + "." + query
			},
		}

		t.Run("bad ARN", func(t *testing.T) {
			_, err := lgs.Search(context.Background(), "foo.bar", "query", false)

			if err == nil {
				t.Error("expected error because the ARN was bad")
			}
		})

		t.Run("good ARN but bad scope", func(t *testing.T) {
			_, err := lgs.Search(context.Background(), "foo.bar", "arn:aws:service:region:account:type/id", false)

			if err == nil {
				t.Error("expected error because the ARN had a bad scope")
			}
		})

		t.Run("good ARN", func(t *testing.T) {
			_, err := lgs.Search(context.Background(), "foo.bar", "arn:aws:service:bar:foo:type/id", false)

			if err != nil {
				t.Error(err)
			}
		})
	})

	t.Run("with Custom & ARN search", func(t *testing.T) {
		lgs := AlwaysGetSource[string, string, string, string, struct{}, struct{}]{
			ItemType:         "test",
			AccountID:        "foo",
			Region:           "bar",
			Client:           struct{}{},
			MaxParallel:      MaxParallel(1),
			ListInput:        "",
			AlwaysSearchARNs: true,
			SearchInputMapper: func(scope, query string) (string, error) {
				return query, nil
			},
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{DataFunc: func() string {
					return "foo"
				}}
			},
			ListFuncOutputMapper: func(output, input string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, input string) (*sdp.Item, error) {
				if input == "foo.bar.id" {
					return &sdp.Item{}, nil
				} else {
					return nil, sdp.NewQueryError(errors.New("bad query details"))
				}
			},
			GetInputMapper: func(scope, query string) string {
				return scope + "." + query
			},
		}

		t.Run("ARN", func(t *testing.T) {
			items, err := lgs.Search(context.Background(), "foo.bar", "arn:aws:service:bar:foo:type/id", false)

			if err != nil {
				t.Error(err)
			}

			if len(items) != 1 {
				t.Errorf("expected 1 item, got %v", len(items))
			}
		})

		t.Run("other search", func(t *testing.T) {
			items, err := lgs.Search(context.Background(), "foo.bar", "id", false)

			if err != nil {
				t.Error(err)
			}

			if len(items) != 0 {
				t.Errorf("expected 0 item, got %v", len(items))
			}
		})
	})

	t.Run("with custom search logic", func(t *testing.T) {
		var searchMapperCalled bool

		lgs := AlwaysGetSource[string, string, string, string, struct{}, struct{}]{
			ItemType:  "test",
			AccountID: "foo",
			Region:    "bar",
			Client:    struct{}{},
			ListInput: "",
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{DataFunc: func() string {
					return "foo"
				}}
			},
			ListFuncOutputMapper: func(output, input string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, input string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
			SearchInputMapper: func(scope, query string) (string, error) {
				searchMapperCalled = true
				return "", nil
			},
			GetInputMapper: func(scope, query string) string {
				return ""
			},
		}

		_, err := lgs.Search(context.Background(), "foo.bar", "bar", false)

		if err != nil {
			t.Error(err)
		}

		if !searchMapperCalled {
			t.Error("search mapper not called")
		}
	})

	t.Run("with SearchGetInputMapper", func(t *testing.T) {
		ags := AlwaysGetSource[string, string, string, string, struct{}, struct{}]{
			ItemType:         "test",
			AccountID:        "foo",
			Region:           "bar",
			Client:           struct{}{},
			MaxParallel:      MaxParallel(1),
			ListInput:        "",
			AlwaysSearchARNs: true,
			SearchGetInputMapper: func(scope, query string) (string, error) {
				return "foo.bar.id", nil
			},
			ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
				// Returns 3 pages
				return &TestPaginator{DataFunc: func() string {
					return "foo"
				}}
			},
			ListFuncOutputMapper: func(output, input string) ([]string, error) {
				// Returns 2 gets per page
				return []string{"", ""}, nil
			},
			GetFunc: func(ctx context.Context, client struct{}, scope, input string) (*sdp.Item, error) {
				if input == "foo.bar.id" {
					return &sdp.Item{}, nil
				} else {
					return nil, sdp.NewQueryError(errors.New("bad query details"))
				}
			},
			GetInputMapper: func(scope, query string) string {
				return scope + "." + query
			},
		}

		items, err := ags.Search(context.Background(), "foo.bar", "id", false)

		if err != nil {
			t.Error(err)
		}

		if len(items) != 1 {
			t.Errorf("expected 1 item, got %v", len(items))
		}
	})
}

func TestAlwaysGetSourceCaching(t *testing.T) {
	ctx := context.Background()
	generation := 0
	s := AlwaysGetSource[string, string, string, string, struct{}, struct{}]{
		ItemType:  "test",
		AccountID: "foo",
		Region:    "eu-west-2",
		Client:    struct{}{},
		ListInput: "",
		ListFuncPaginatorBuilder: func(client struct{}, input string) Paginator[string, struct{}] {
			return &TestPaginator{
				DataFunc: func() string {
					generation += 1
					return fmt.Sprintf("%v", generation)
				},
				MaxPages: 1,
			}
		},
		ListFuncOutputMapper: func(output, input string) ([]string, error) {
			// Returns only 1 get per page to avoid confusing the cache with duplicate items
			return []string{""}, nil
		},
		GetFunc: func(ctx context.Context, client struct{}, scope, input string) (*sdp.Item, error) {
			generation += 1
			return &sdp.Item{Scope: "foo.eu-west-2",
				Type:            "test-type",
				UniqueAttribute: "name",
				Attributes: &sdp.ItemAttributes{
					AttrStruct: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"name":       structpb.NewStringValue("test-item"),
							"generation": structpb.NewStringValue(fmt.Sprintf("%v%v", input, generation)),
						},
					},
				}}, nil
		},
		GetInputMapper: func(scope, query string) string {
			return ""
		},
	}

	t.Run("get", func(t *testing.T) {
		// get
		first, err := s.Get(ctx, "foo.eu-west-2", "test-item", false)
		if err != nil {
			t.Fatal(err)
		}
		firstGen, err := first.Attributes.Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		// get again
		withCache, err := s.Get(ctx, "foo.eu-west-2", "test-item", false)
		if err != nil {
			t.Fatal(err)
		}
		withCacheGen, err := withCache.Attributes.Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		if firstGen != withCacheGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withCacheGen)
		}

		// get ignore cache
		withoutCache, err := s.Get(ctx, "foo.eu-west-2", "test-item", true)
		if err != nil {
			t.Fatal(err)
		}
		withoutCacheGen, err := withoutCache.Attributes.Get("generation")
		if err != nil {
			t.Fatal(err)
		}
		if withoutCacheGen == firstGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withoutCacheGen)
		}
	})

	t.Run("list", func(t *testing.T) {
		// list
		first, err := s.List(ctx, "foo.eu-west-2", false)
		if err != nil {
			t.Fatal(err)
		}
		firstGen, err := first[0].Attributes.Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		// list again
		withCache, err := s.List(ctx, "foo.eu-west-2", false)
		if err != nil {
			t.Fatal(err)
		}
		withCacheGen, err := withCache[0].Attributes.Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		if firstGen != withCacheGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withCacheGen)
		}

		// list ignore cache
		withoutCache, err := s.List(ctx, "foo.eu-west-2", true)
		if err != nil {
			t.Fatal(err)
		}
		withoutCacheGen, err := withoutCache[0].Attributes.Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		if withoutCacheGen == firstGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withoutCacheGen)
		}
	})

	t.Run("search", func(t *testing.T) {
		// search
		first, err := s.Search(ctx, "foo.eu-west-2", "arn:aws:test-type:eu-west-2:foo:test-item", false)
		if err != nil {
			t.Fatal(err)
		}
		firstGen, err := first[0].Attributes.Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		// search again
		withCache, err := s.Search(ctx, "foo.eu-west-2", "arn:aws:test-type:eu-west-2:foo:test-item", false)
		if err != nil {
			t.Fatal(err)
		}
		withCacheGen, err := withCache[0].Attributes.Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		if firstGen != withCacheGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withCacheGen)
		}

		// search ignore cache
		withoutCache, err := s.Search(ctx, "foo.eu-west-2", "arn:aws:test-type:eu-west-2:foo:test-item", true)
		if err != nil {
			t.Fatal(err)
		}
		withoutCacheGen, err := withoutCache[0].Attributes.Get("generation")
		if err != nil {
			t.Fatal(err)
		}
		if withoutCacheGen == firstGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withoutCacheGen)
		}
	})
}
