package sources

import (
	"testing"
)

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
			t.Errorf("expected region to be region, got %v", a.Region)
		}

		if a.ResourceID() != "resource-id" {
			t.Errorf("expected resource ID to be resource-id, got %v", a.ResourceID())
		}

		if a.Service != "service" {
			t.Errorf("expected service to be service, got %v", a.Service)
		}
	})

	t.Run("arn:aws:ecs:eu-west-1:052392120703:task-definition/ecs-template-ecs-demo-app:1", func(t *testing.T) {
		arn := "arn:aws:ecs:eu-west-1:052392120703:task-definition/ecs-template-ecs-demo-app:1"

		a, err := ParseARN(arn)

		if err != nil {
			t.Error(err)
		}

		if a.AccountID != "052392120703" {
			t.Errorf("expected account ID to be 052392120703, got %v", a.AccountID)
		}

		if a.Region != "eu-west-1" {
			t.Errorf("expected region to be eu-west-1, got %v", a.Region)
		}

		if a.Service != "ecs" {
			t.Errorf("expected service to be ecs, got %v", a.Service)
		}

		if a.Resource != "task-definition/ecs-template-ecs-demo-app:1" {
			t.Errorf("expected resource ID to be task-definition/ecs-template-ecs-demo-app:1, got %v", a.ResourceID())
		}

		if a.ResourceID() != "ecs-template-ecs-demo-app:1" {
			t.Errorf("expected ResourceID to be ecs-template-ecs-demo-app:1, got %v", a.ResourceID())
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

		if a.ResourceID() != "i-054dsfg34gdsfg38" {
			t.Errorf("expected account ID to be i-054dsfg34gdsfg38, got %v", a.ResourceID())
		}
	})

	t.Run("arn:aws:eks:eu-west-2:944651592624:nodegroup/dogfood/intel-20230616142016591700000005/6ec4624a-05ef-bdad-e69a-fe9832885421", func(t *testing.T) {
		arn := "arn:aws:eks:eu-west-2:944651592624:nodegroup/dogfood/intel-20230616142016591700000005/6ec4624a-05ef-bdad-e69a-fe9832885421"

		a, err := ParseARN(arn)

		if err != nil {
			t.Error(err)
		}

		if a.AccountID != "944651592624" {
			t.Errorf("expected account ID to be 944651592624, got %v", a.AccountID)
		}

		if a.Region != "eu-west-2" {
			t.Errorf("expected account ID to be eu-west-2, got %v", a.Region)
		}

		if a.ResourceID() != "dogfood/intel-20230616142016591700000005/6ec4624a-05ef-bdad-e69a-fe9832885421" {
			t.Errorf("expected account ID to be dogfood/intel-20230616142016591700000005/6ec4624a-05ef-bdad-e69a-fe9832885421, got %v", a.ResourceID())
		}
	})

	t.Run("arn:aws:iam::942836531449:policy/OvermindReadonly", func(t *testing.T) {
		arn := "arn:aws:iam::942836531449:policy/OvermindReadonly"

		a, err := ParseARN(arn)

		if err != nil {
			t.Error(err)
		}

		if a.ResourceID() != "OvermindReadonly" {
			t.Errorf("expected account ID to be OvermindReadonly, got %v", a.ResourceID())
		}
	})

	t.Run("arn:aws:elasticloadbalancing:eu-west-2:540044833068:targetgroup/lambda-rvaaio9n3auuhnvvvjmp/6f23de9c63bd4653", func(t *testing.T) {
		arn := "arn:aws:elasticloadbalancing:eu-west-2:540044833068:targetgroup/lambda-rvaaio9n3auuhnvvvjmp/6f23de9c63bd4653"

		a, err := ParseARN(arn)

		if err != nil {
			t.Error(err)
		}

		if a.Type() != "targetgroup" {
			t.Errorf("expected type to be targetgroup, got %v", a.Type())
		}
	})
}
