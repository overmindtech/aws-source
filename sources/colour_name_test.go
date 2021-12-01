package sources

import (
	"regexp"
	"testing"

	"github.com/overmindtech/sdp-go"
)

// This file contains tests for the ColourNameSource source. It is a good idea
// to write as many exhaustive tests as possible at this level to ensure that
// your source responds correctly to certain requests.
func TestGet(t *testing.T) {
	tests := []SourceTest{
		{
			Name:          "Getting a known colour",
			ItemContext:   "global",
			Query:         "GreenYellow",
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"name": "GreenYellow",
					},
				},
			},
		},
		{
			Name:        "Getting an unknown colour",
			ItemContext: "global",
			Query:       "UpsideDownBlack",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &ExpectedError{
				Type:             sdp.ItemRequestError_NOTFOUND,
				ErrorStringRegex: regexp.MustCompile("not recognized"),
				Context:          "global",
			},
			ExpectedItems: nil,
		},
		{
			Name:        "Getting an unknown context",
			ItemContext: "wonkySpace",
			Query:       "Red",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &ExpectedError{
				Type:             sdp.ItemRequestError_NOCONTEXT,
				ErrorStringRegex: regexp.MustCompile("colours are only supported"),
				Context:          "wonkySpace",
			},
			ExpectedItems: nil,
		},
	}

	RunSourceTests(t, tests, &ColourNameSource{})
}

func TestFind(t *testing.T) {
	tests := []SourceTest{
		{
			Name:          "Using correct context",
			ItemContext:   "global",
			Method:        sdp.RequestMethod_FIND,
			ExpectedError: nil,
			ExpectedItems: &ExpectedItems{
				NumItems: 147,
			},
		},
		{
			Name:        "Using incorrect context",
			ItemContext: "somethingElse",
			Method:      sdp.RequestMethod_FIND,
			ExpectedError: &ExpectedError{
				Type:             sdp.ItemRequestError_NOCONTEXT,
				ErrorStringRegex: regexp.MustCompile("colours are only supported"),
				Context:          "somethingElse",
			},
			ExpectedItems: nil,
		},
	}

	RunSourceTests(t, tests, &ColourNameSource{})
}
