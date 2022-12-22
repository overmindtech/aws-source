package ec2

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/sdp-go"
)

func TestType(t *testing.T) {
	s := EC2Source[string, string]{
		ItemType: "foo",
	}

	if s.Type() != "foo" {
		t.Errorf("expected type to be foo, got %v", s.Type())
	}
}

func TestName(t *testing.T) {
	// Basically just test that it's not empty. It doesn't matter what it is
	s := EC2Source[string, string]{
		ItemType: "foo",
	}

	if s.Name() == "" {
		t.Error("blank name")
	}
}

func TestScopes(t *testing.T) {
	s := EC2Source[string, string]{
		Config: aws.Config{
			Region: "outer-space",
		},
		AccountID: "mars",
	}

	scopes := s.Scopes()

	if len(scopes) != 1 {
		t.Errorf("expected 1 scope, got %v", len(scopes))
	}

	if scopes[0] != "mars.outer-space" {
		t.Errorf("expected scope to be mars.outer-space, got %v", scopes[0])
	}
}

func TestGet(t *testing.T) {
	t.Run("when everything goes well", func(t *testing.T) {
		var inputMapperCalled bool
		var outputMapperCalled bool
		var describeFuncCalled bool

		s := EC2Source[string, string]{
			Config: aws.Config{
				Region: "eu-west-2",
			},
			AccountID: "foo",
			InputMapperGet: func(scope, query string) (string, error) {
				inputMapperCalled = true
				return "input", nil
			},
			InputMapperList: func(scope string) (string, error) {
				return "input", nil
			},
			OutputMapper: func(scope, output string) ([]*sdp.Item, error) {
				outputMapperCalled = true
				return []*sdp.Item{
					{},
				}, nil
			},
			DescribeFunc: func(ctx context.Context, client *ec2.Client, input string, optFns ...func(*ec2.Options)) (string, error) {
				describeFuncCalled = true
				return "", nil
			},
		}

		item, err := s.Get(context.Background(), "foo.eu-west-2", "bar")

		if err != nil {
			t.Error(err)
		}

		if !inputMapperCalled {
			t.Error("input mapper not called")
		}

		if !outputMapperCalled {
			t.Error("output mapper not called")
		}

		if !describeFuncCalled {
			t.Error("describe func not called")
		}

		if item == nil {
			t.Error("nil item")
		}
	})

	t.Run("with too many results", func(t *testing.T) {
		s := EC2Source[string, string]{
			Config: aws.Config{
				Region: "eu-west-2",
			},
			AccountID: "foo",
			InputMapperGet: func(scope, query string) (string, error) {
				return "input", nil
			},
			InputMapperList: func(scope string) (string, error) {
				return "input", nil
			},
			OutputMapper: func(scope, output string) ([]*sdp.Item, error) {
				return []*sdp.Item{
					{},
					{},
					{},
				}, nil
			},
			DescribeFunc: func(ctx context.Context, client *ec2.Client, input string, optFns ...func(*ec2.Options)) (string, error) {
				return "", nil
			},
		}

		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar")

		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("with no results", func(t *testing.T) {
		s := EC2Source[string, string]{
			Config: aws.Config{
				Region: "eu-west-2",
			},
			AccountID: "foo",
			InputMapperGet: func(scope, query string) (string, error) {
				return "input", nil
			},
			InputMapperList: func(scope string) (string, error) {
				return "input", nil
			},
			OutputMapper: func(scope, output string) ([]*sdp.Item, error) {
				return []*sdp.Item{}, nil
			},
			DescribeFunc: func(ctx context.Context, client *ec2.Client, input string, optFns ...func(*ec2.Options)) (string, error) {
				return "", nil
			},
		}

		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar")

		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestSearch(t *testing.T) {
	s := EC2Source[string, string]{
		Config: aws.Config{
			Region: "region",
		},
		AccountID: "account-id",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(scope, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input string, optFns ...func(*ec2.Options)) (string, error) {
			return "fancy", nil
		},
	}

	items, err := s.Search(context.Background(), "account-id.region", "arn:partition:service:region:account-id:resource-type:resource-id")

	if err != nil {
		t.Error(err)
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %v", len(items))
	}
}

func TestNoInputMapper(t *testing.T) {
	s := EC2Source[string, string]{
		Config: aws.Config{
			Region: "eu-west-2",
		},
		AccountID: "foo",
		OutputMapper: func(scope, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input string, optFns ...func(*ec2.Options)) (string, error) {
			return "", nil
		},
	}

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar")

		if err == nil {
			t.Error("expected error but got nil")
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2")

		if err == nil {
			t.Error("expected error but got nil")
		}
	})
}

func TestNoOutputMapper(t *testing.T) {
	s := EC2Source[string, string]{
		Config: aws.Config{
			Region: "eu-west-2",
		},
		AccountID: "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input string, optFns ...func(*ec2.Options)) (string, error) {
			return "", nil
		},
	}

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar")

		if err == nil {
			t.Error("expected error but got nil")
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2")

		if err == nil {
			t.Error("expected error but got nil")
		}
	})
}

func TestNoDescribeFunc(t *testing.T) {
	s := EC2Source[string, string]{
		Config: aws.Config{
			Region: "eu-west-2",
		},
		AccountID: "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(scope, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
	}

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar")

		if err == nil {
			t.Error("expected error but got nil")
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2")

		if err == nil {
			t.Error("expected error but got nil")
		}
	})
}

func TestFailingInputMapper(t *testing.T) {
	s := EC2Source[string, string]{
		Config: aws.Config{
			Region: "eu-west-2",
		},
		AccountID: "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", errors.New("foobar")
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", errors.New("foobar")
		},
		OutputMapper: func(scope, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input string, optFns ...func(*ec2.Options)) (string, error) {
			return "", nil
		},
	}

	fooBar := regexp.MustCompile("foobar")

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar")

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2")

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})
}

func TestFailingOutputMapper(t *testing.T) {
	s := EC2Source[string, string]{
		Config: aws.Config{
			Region: "eu-west-2",
		},
		AccountID: "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(scope, output string) ([]*sdp.Item, error) {
			return nil, errors.New("foobar")
		},
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input string, optFns ...func(*ec2.Options)) (string, error) {
			return "", nil
		},
	}

	fooBar := regexp.MustCompile("foobar")

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar")

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2")

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})
}

func TestFailingDescribeFunc(t *testing.T) {
	s := EC2Source[string, string]{
		Config: aws.Config{
			Region: "eu-west-2",
		},
		AccountID: "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(scope, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input string, optFns ...func(*ec2.Options)) (string, error) {
			return "", errors.New("foobar")
		},
	}

	fooBar := regexp.MustCompile("foobar")

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar")

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2")

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})
}

type TestPaginator struct {
	page int
}

func (t *TestPaginator) HasMorePages() bool {
	return t.page < 3
}

func (t *TestPaginator) NextPage(context.Context, ...func(*ec2.Options)) (string, error) {
	t.page++

	return "", nil
}

func TestPaginated(t *testing.T) {
	s := EC2Source[string, string]{
		MaxResultsPerPage: 1,
		Config: aws.Config{
			Region: "eu-west-2",
		},
		AccountID: "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(scope, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
		PaginatorBuilder: func(client *ec2.Client, params string) Paginator[string] {
			return &TestPaginator{}
		},
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input string, optFns ...func(*ec2.Options)) (string, error) {
			return "", nil
		},
	}

	t.Run("detecting pagination", func(t *testing.T) {
		if !s.Paginated() {
			t.Error("pagination not detected")
		}

		if err := s.Validate(); err != nil {
			t.Error(err)
		}
	})

	t.Run("paginsting a List query", func(t *testing.T) {
		items, err := s.List(context.Background(), "foo.eu-west-2")

		if err != nil {
			t.Error(err)
		}

		if len(items) != 3 {
			t.Errorf("expected 3 items, got %v", len(items))
		}
	})
}
