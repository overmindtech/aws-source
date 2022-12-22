package sources

import "testing"

func TestParseARN(t *testing.T) {
	t.Run("arn:partition:service:region:account-id:resource-type:resource-id", func(t *testing.T) {
		arn := "arn:partition:service:region:account-id:resource-type:resource-id"

		a, err := ParseARN(arn)

		if err != nil {
			t.Error(err)
		}

		if a.AccountID != "account-id" {
			t.Errorf("expected account ID to be account-id, got %v", a.AccountID)
		}

		if a.Region != "region" {
			t.Errorf("expected account ID to be region, got %v", a.Region)
		}

		if a.ResourceID != "resource-id" {
			t.Errorf("expected account ID to be resource-id, got %v", a.ResourceID)
		}
	})

	t.Run("arn:aws:ec2:us-east-1:4575734578134:instance/i-054dsfg34gdsfg38", func(t *testing.T) {
		arn := "arn:aws:ec2:us-east-1:4575734578134:instance/i-054dsfg34gdsfg38"

		a, err := ParseARN(arn)

		if err != nil {
			t.Error(err)
		}

		if a.AccountID != "4575734578134" {
			t.Errorf("expected account ID to be 4575734578134, got %v", a.AccountID)
		}

		if a.Region != "us-east-1" {
			t.Errorf("expected account ID to be us-east-1, got %v", a.Region)
		}

		if a.ResourceID != "i-054dsfg34gdsfg38" {
			t.Errorf("expected account ID to be i-054dsfg34gdsfg38, got %v", a.ResourceID)
		}
	})
}
