package apigateway

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/overmindtech/aws-source/adapters"
)

/*
{
   "id": "string",
   "parentId": "string",
   "path": "string",
   "pathPart": "string",
   "resourceMethods": {
      "string" : {
         "apiKeyRequired": boolean,
         "authorizationScopes": [ "string" ],
         "authorizationType": "string",
         "authorizerId": "string",
         "httpMethod": "string",
         "methodIntegration": {
            "cacheKeyParameters": [ "string" ],
            "cacheNamespace": "string",
            "connectionId": "string",
            "connectionType": "string",
            "contentHandling": "string",
            "credentials": "string",
            "httpMethod": "string",
            "integrationResponses": {
               "string" : {
                  "contentHandling": "string",
                  "responseParameters": {
                     "string" : "string"
                  },
                  "responseTemplates": {
                     "string" : "string"
                  },
                  "selectionPattern": "string",
                  "statusCode": "string"
               }
            },
            "passthroughBehavior": "string",
            "requestParameters": {
               "string" : "string"
            },
            "requestTemplates": {
               "string" : "string"
            },
            "timeoutInMillis": number,
            "tlsConfig": {
               "insecureSkipVerification": boolean
            },
            "type": "string",
            "uri": "string"
         },
         "methodResponses": {
            "string" : {
               "responseModels": {
                  "string" : "string"
               },
               "responseParameters": {
                  "string" : boolean
               },
               "statusCode": "string"
            }
         },
         "operationName": "string",
         "requestModels": {
            "string" : "string"
         },
         "requestParameters": {
            "string" : boolean
         },
         "requestValidatorId": "string"
      }
   }
}
*/

func TestResourceOutputMapper(t *testing.T) {
	resource := &types.Resource{
		Id:       adapters.PtrString("test-id"),
		ParentId: adapters.PtrString("parent-id"),
		Path:     adapters.PtrString("/test-path"),
		PathPart: adapters.PtrString("test-path-part"),
		ResourceMethods: map[string]types.Method{
			"GET": {
				ApiKeyRequired:      adapters.PtrBool(true),
				AuthorizationScopes: []string{"scope1", "scope2"},
				AuthorizationType:   adapters.PtrString("NONE"),
				AuthorizerId:        adapters.PtrString("authorizer-id"),
				HttpMethod:          adapters.PtrString("GET"),
				MethodIntegration: &types.Integration{
					CacheKeyParameters: []string{"param1", "param2"},
					CacheNamespace:     adapters.PtrString("namespace"),
					ConnectionId:       adapters.PtrString("connection-id"),
					ConnectionType:     types.ConnectionTypeInternet,
					ContentHandling:    types.ContentHandlingStrategyConvertToBinary,
					Credentials:        adapters.PtrString("credentials"),
					HttpMethod:         adapters.PtrString("POST"),
					IntegrationResponses: map[string]types.IntegrationResponse{
						"200": {
							ContentHandling: types.ContentHandlingStrategyConvertToText,
							ResponseParameters: map[string]string{
								"param1": "value1",
							},
							ResponseTemplates: map[string]string{
								"template1": "value1",
							},
							SelectionPattern: adapters.PtrString("pattern"),
							StatusCode:       adapters.PtrString("200"),
						},
					},
					PassthroughBehavior: adapters.PtrString("WHEN_NO_MATCH"),
					RequestParameters: map[string]string{
						"param1": "value1",
					},
					RequestTemplates: map[string]string{
						"template1": "value1",
					},
					TimeoutInMillis: int32(29000),
					TlsConfig: &types.TlsConfig{
						InsecureSkipVerification: false,
					},
					Type: types.IntegrationTypeAwsProxy,
					Uri:  adapters.PtrString("uri"),
				},
				MethodResponses: map[string]types.MethodResponse{
					"200": {
						ResponseModels: map[string]string{
							"model1": "value1",
						},
						ResponseParameters: map[string]bool{
							"param1": true,
						},
						StatusCode: adapters.PtrString("200"),
					},
				},
				OperationName: adapters.PtrString("operation"),
				RequestModels: map[string]string{
					"model1": "value1",
				},
				RequestParameters: map[string]bool{
					"param1": true,
				},
				RequestValidatorId: adapters.PtrString("validator-id"),
			},
		},
	}

	item, err := resourceOutputMapper("rest-api-13", "scope", resource)
	if err != nil {
		t.Fatal(err)
	}

	if err := item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewResourceSource(t *testing.T) {
	config, account, region := adapters.GetAutoConfig(t)

	client := apigateway.NewFromConfig(config)

	source := NewResourceSource(client, account, region)

	test := adapters.E2ETest{
		Adapter:  source,
		Timeout:  10 * time.Second,
		SkipList: true,
	}

	test.Run(t)
}
