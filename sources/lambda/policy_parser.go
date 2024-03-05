package lambda

// This is derived from the AWS example:
// https://github.com/awsdocs/aws-doc-sdk-examples/blob/main/gov2/iam/actions/policies.go#L21C1-L32C2
// and represents the structure of an IAM policy document
type PolicyDocument struct {
	Version   string
	Statement []PolicyStatement
}

// PolicyStatement defines a statement in a policy document.
type PolicyStatement struct {
	Action    string
	Principal Principal `json:",omitempty"`
	Condition Condition `json:",omitempty"`
}

type Principal struct {
	Service string `json:",omitempty"`
}

type Condition struct {
	ArnLike      ArnLikeCondition      `json:",omitempty"`
	StringEquals StringEqualsCondition `json:",omitempty"`
}

type StringEqualsCondition struct {
	AWSSourceAccount string `json:"AWS:SourceAccount,omitempty"`
}

type ArnLikeCondition struct {
	AWSSourceArn string `json:"AWS:SourceArn,omitempty"`
}
