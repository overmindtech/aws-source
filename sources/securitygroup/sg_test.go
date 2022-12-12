package securitygroup

// import (
// 	"context"
// 	"testing"

// 	"github.com/aws/aws-sdk-go-v2/service/ec2"
// 	"github.com/overmindtech/discovery"
// )

// func TestSG(t *testing.T) {
// 	t.Parallel()
// 	tr := createSG(t)

// 	src := SecurityGroupSource{
// 		Config:    TestAWSConfig,
// 		AccountID: TestAccountID,
// 	}

// 	t.Run("Get with correct security group ID", func(t *testing.T) {
// 		item, err := src.Get(context.Background(), TestContext, tr.GroupID)

// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		discovery.TestValidateItem(t, item)
// 	})

// 	t.Run("Get with incorrect security group ID", func(t *testing.T) {
// 		_, err := src.Get(context.Background(), TestContext, "i-0ecfa0a234cbc132")

// 		if err == nil {
// 			t.Error("expected error but got nil")
// 		}
// 	})

// 	t.Run("Find", func(t *testing.T) {
// 		items, err := src.Find(context.Background(), TestContext)

// 		if err != nil {
// 			t.Error(err)
// 		}

// 		if len(items) == 0 {
// 			t.Error("Expected items to be found but got nothing")
// 		}

// 		discovery.TestValidateItems(t, items)
// 	})
// }
