package sources

import (
	"context"
	"errors"
	"testing"

	"github.com/overmindtech/sdp-go"
)

func TestGetListSourceType(t *testing.T) {
	s := GetListSource[string, struct{}, struct{}]{
		ItemType: "foo",
	}

	if s.Type() != "foo" {
		t.Errorf("expected type to be foo got %v", s.Type())
	}
}

func TestGetListSourceName(t *testing.T) {
	s := GetListSource[string, struct{}, struct{}]{
		ItemType: "foo",
	}

	if s.Name() != "foo-source" {
		t.Errorf("expected type to be foo-source got %v", s.Name())
	}
}

func TestGetListSourceScopes(t *testing.T) {
	s := GetListSource[string, struct{}, struct{}]{
		AccountID: "foo",
		Region:    "bar",
	}

	if s.Scopes()[0] != "foo.bar" {
		t.Errorf("expected scope to be foo.bar, got %v", s.Scopes()[0])
	}
}

func TestGetListSourceGet(t *testing.T) {
	t.Run("with no errors", func(t *testing.T) {
		s := GetListSource[string, struct{}, struct{}]{
			ItemType:  "person",
			Region:    "eu-west-2",
			AccountID: "12345",
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (string, error) {
				return "", nil
			},
			ListFunc: func(ctx context.Context, client struct{}, scope string) ([]string, error) {
				return []string{"", ""}, nil
			},
			ItemMapper: func(scope string, awsItem string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
		}

		if _, err := s.Get(context.Background(), "12345.eu-west-2", ""); err != nil {
			t.Error(err)
		}
	})

	t.Run("with an error in the GetFunc", func(t *testing.T) {
		s := GetListSource[string, struct{}, struct{}]{
			ItemType:  "person",
			Region:    "eu-west-2",
			AccountID: "12345",
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (string, error) {
				return "", errors.New("get func error")
			},
			ListFunc: func(ctx context.Context, client struct{}, scope string) ([]string, error) {
				return []string{"", ""}, nil
			},
			ItemMapper: func(scope string, awsItem string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
		}

		if _, err := s.Get(context.Background(), "12345.eu-west-2", ""); err == nil {
			t.Error("expected error got nil")
		}
	})

	t.Run("with an error in the mapper", func(t *testing.T) {
		s := GetListSource[string, struct{}, struct{}]{
			ItemType:  "person",
			Region:    "eu-west-2",
			AccountID: "12345",
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (string, error) {
				return "", nil
			},
			ListFunc: func(ctx context.Context, client struct{}, scope string) ([]string, error) {
				return []string{"", ""}, nil
			},
			ItemMapper: func(scope string, awsItem string) (*sdp.Item, error) {
				return &sdp.Item{}, errors.New("mapper error")
			},
		}

		if _, err := s.Get(context.Background(), "12345.eu-west-2", ""); err == nil {
			t.Error("expected error got nil")
		}
	})
}

func TestGetListSourceList(t *testing.T) {
	t.Run("with no errors", func(t *testing.T) {
		s := GetListSource[string, struct{}, struct{}]{
			ItemType:  "person",
			Region:    "eu-west-2",
			AccountID: "12345",
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (string, error) {
				return "", nil
			},
			ListFunc: func(ctx context.Context, client struct{}, scope string) ([]string, error) {
				return []string{"", ""}, nil
			},
			ItemMapper: func(scope string, awsItem string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
		}

		if items, err := s.List(context.Background(), "12345.eu-west-2"); err != nil {
			t.Error(err)
		} else {
			if len(items) != 2 {
				t.Errorf("expected 2 items, got %v", len(items))
			}
		}
	})

	t.Run("with an error in the ListFunc", func(t *testing.T) {
		s := GetListSource[string, struct{}, struct{}]{
			ItemType:  "person",
			Region:    "eu-west-2",
			AccountID: "12345",
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (string, error) {
				return "", nil
			},
			ListFunc: func(ctx context.Context, client struct{}, scope string) ([]string, error) {
				return []string{"", ""}, errors.New("list func error")
			},
			ItemMapper: func(scope string, awsItem string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
		}

		if _, err := s.List(context.Background(), "12345.eu-west-2"); err == nil {
			t.Error("expected error got nil")
		}
	})

	t.Run("with an error in the mapper", func(t *testing.T) {
		s := GetListSource[string, struct{}, struct{}]{
			ItemType:  "person",
			Region:    "eu-west-2",
			AccountID: "12345",
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (string, error) {
				return "", nil
			},
			ListFunc: func(ctx context.Context, client struct{}, scope string) ([]string, error) {
				return []string{"", ""}, nil
			},
			ItemMapper: func(scope string, awsItem string) (*sdp.Item, error) {
				return &sdp.Item{}, errors.New("mapper error")
			},
		}

		if items, err := s.List(context.Background(), "12345.eu-west-2"); err != nil {
			t.Error(err)
		} else {
			if len(items) != 0 {
				t.Errorf("expected no items, got %v", len(items))
			}
		}
	})
}

func TestGetListSourceSearch(t *testing.T) {
	t.Run("with ARN search", func(t *testing.T) {
		s := GetListSource[string, struct{}, struct{}]{
			ItemType:  "person",
			Region:    "eu-west-2",
			AccountID: "12345",
			GetFunc: func(ctx context.Context, client struct{}, scope, query string) (string, error) {
				return "", nil
			},
			ListFunc: func(ctx context.Context, client struct{}, scope string) ([]string, error) {
				return []string{"", ""}, nil
			},
			ItemMapper: func(scope string, awsItem string) (*sdp.Item, error) {
				return &sdp.Item{}, nil
			},
		}

		t.Run("bad ARN", func(t *testing.T) {
			_, err := s.Search(context.Background(), "12345.eu-west-2", "query")

			if err == nil {
				t.Error("expected error because the ARN was bad")
			}
		})

		t.Run("good ARN but bad scope", func(t *testing.T) {
			_, err := s.Search(context.Background(), "12345.eu-west-2", "arn:aws:service:region:account:type/id")

			if err == nil {
				t.Error("expected error because the ARN had a bad scope")
			}
		})

		t.Run("good ARN", func(t *testing.T) {
			_, err := s.Search(context.Background(), "12345.eu-west-2", "arn:aws:service:eu-west-2:12345:type/id")

			if err != nil {
				t.Error(err)
			}
		})
	})
}
