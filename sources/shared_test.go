package sources

import (
	"context"
	"regexp"
	"testing"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
)

// This file contains shared testing libraries to make testing sources easier

type ExpectedError struct {
	// The expected type of the error
	Type sdp.ItemRequestError_ErrorType

	// A pointer to a regex that will be used to validate the error message,
	// leave as `nil`if you don't want to check this
	ErrorStringRegex *regexp.Regexp

	// The context that the error should come from. Leave as "" if you don't
	// want to check this
	Context string
}

type ExpectedItems struct {
	// The expected number of items
	NumItems int

	// A list of expected attributes for the items, will be checked in order
	// with the first set of attributes neeing to match those of the first item
	// etc. Note that this doesn't need to have the same number of entries as
	// there are items
	ExpectedAttributes []map[string]interface{}
}

type SourceTest struct {
	// Name of the test for logging
	Name string
	// The context to be passed to the Get() request
	ItemContext string
	// The query to be passed
	Query string
	// The method that should be used
	Method sdp.RequestMethod
	// Details of the expected error, `nil` means no error
	ExpectedError *ExpectedError
	// The expected items
	ExpectedItems *ExpectedItems
}

func RunSourceTests(t *testing.T, tests []SourceTest, source discovery.Source) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var item *sdp.Item
			var items []*sdp.Item
			var err error

			switch test.Method {
			case sdp.RequestMethod_FIND:
				items, err = source.Find(context.Background(), test.ItemContext)
			case sdp.RequestMethod_SEARCH:
				searchable, ok := source.(discovery.SearchableSource)

				if !ok {
					t.Fatal("Supplied source did not fulfill discovery.SearchableSource interface. Cannot execute search tests against this source")
				}

				items, err = searchable.Search(context.Background(), test.ItemContext, test.Query)
			case sdp.RequestMethod_GET:
				item, err = source.Get(context.Background(), test.ItemContext, test.Query)
				items = []*sdp.Item{item}
			default:
				t.Fatalf("Test Method invalid: %v. Should be one of: sdp.RequestMethod_FIND, sdp.RequestMethod_SEARCH, sdp.RequestMethod_GET", test.Method)
			}

			// If an error was expected then validate that it was found
			if ee := test.ExpectedError; ee != nil {
				if err == nil {
					t.Error("expected error but got nil")
				}

				ire, ok := err.(*sdp.ItemRequestError)

				if !ok {
					t.Fatalf("error returned was type %T, expected *sdp.ItemRequestError", err)
				}

				if ee.Type != ire.ErrorType {
					t.Fatalf("error type was %v, expected %v", ire.ErrorType, ee.Type)
				}

				if ee.Context != "" {
					if ee.Context != ire.Context {
						t.Fatalf("error context was %v, expected %v", ire.Context, ee.Context)
					}
				}

				if ee.ErrorStringRegex != nil {
					if !ee.ErrorStringRegex.MatchString(ire.ErrorString) {
						t.Fatalf("error string did not match regex %v, raw value: %v", ee.ErrorStringRegex, ire.ErrorString)
					}
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
			}

			if ei := test.ExpectedItems; ei != nil {
				if len(items) != ei.NumItems {
					t.Fatalf("expected %v items, got %v", ei.NumItems, len(items))
				}

				for _, item := range items {
					RunItemValidationTest(t, item)
				}

				// Loop over the expected attributes and check
				for i, expectedAttributes := range ei.ExpectedAttributes {
					relevantItem := items[i]

					for key, expectedValue := range expectedAttributes {
						value, err := relevantItem.Attributes.Get(key)

						if err != nil {
							t.Error(err)
						}

						// Deal with comparing slices
						if expectedStringSlice, ok := expectedValue.([]interface{}); ok {
							if !interfaceSliceEqual(expectedStringSlice, value.([]interface{})) {
								t.Errorf("expected attribute %v to be %v, got %v", key, expectedValue, value)
							}
						} else {
							if value != expectedValue {
								t.Errorf("expected attribute %v to be %v, got %v", key, expectedValue, value)
							}
						}
					}
				}
			}
		})
	}
}

func interfaceSliceEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// RunItemValidationTest Checks an item to ensure it is a valid SDP item. This includes
// checking that all required attributes are populated
func RunItemValidationTest(t *testing.T, i *sdp.Item) {
	// Ensure that the item has the required fields set i.e.
	//
	// * Type
	// * UniqueAttribute
	// * Context
	// * Attributes
	if i.GetType() == "" {
		t.Errorf("Item %v has an empty Type", i.GloballyUniqueName())
	}

	if i.GetUniqueAttribute() == "" {
		t.Errorf("Item %v has an empty UniqueAttribute", i.GloballyUniqueName())
	}

	if i.GetContext() == "" {
		t.Errorf("Item %v has an empty Context", i.GloballyUniqueName())
	}

	attrMap := i.GetAttributes().AttrStruct.AsMap()

	if len(attrMap) == 0 {
		t.Errorf("Attributes for item %v are empty", i.GloballyUniqueName())
	}

	// Check the attributes themselves for validity
	for k := range attrMap {
		if k == "" {
			t.Errorf("Item %v has an attribute with an empty name", i.GloballyUniqueName())
		}
	}

	// Make sure that the UniqueAttributeValue is populated
	if i.UniqueAttributeValue() == "" {
		t.Errorf("UniqueAttribute %v for item %v is empty", i.GetUniqueAttribute(), i.GloballyUniqueName())
	}

	for index, linkedItem := range i.GetLinkedItems() {
		if linkedItem.GetType() == "" {
			t.Errorf("LinkedItem %v of item %v has empty type", index, i.GloballyUniqueName())
		}

		if linkedItem.GetUniqueAttributeValue() == "" {
			t.Errorf("LinkedItem %v of item %v has empty UniqueAttributeValue", index, i.GloballyUniqueName())
		}

		if linkedItem.GetContext() == "" {
			t.Errorf("LinkedItem %v of item %v has empty Context", index, i.GloballyUniqueName())
		}
	}

	for index, linkedItemRequest := range i.GetLinkedItemRequests() {
		if linkedItemRequest.GetType() == "" {
			t.Errorf("LinkedItemRequest %v of item %v has empty Type", index, i.GloballyUniqueName())
		}

		if linkedItemRequest.GetContext() == "" {
			t.Errorf("LinkedItemRequest %v of item %v has empty Context", index, i.GloballyUniqueName())
		}

		if linkedItemRequest.GetMethod() != sdp.RequestMethod_FIND {
			if linkedItemRequest.GetQuery() == "" {
				t.Errorf("LinkedItemRequest %v of item %v has empty query. This is not allowed unless the method is FIND", index, i.GloballyUniqueName())
			}
		}
	}
}
