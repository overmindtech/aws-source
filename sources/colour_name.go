package sources

import (
	"context"
	"fmt"

	"github.com/overmindtech/sdp-go"
)

// This is an example source that returns the details of given colours in HTML.
// The source needs to implement all of the methods that satisfy the
// discovery.Source interface:
//
// https://pkg.go.dev/github.com/overmindtech/discovery#Source
//

// Database of standard colour names
var Colours = map[string]string{
	"AliceBlue":            "#F0F8FF",
	"AntiqueWhite":         "#FAEBD7",
	"Aqua":                 "#00FFFF",
	"Aquamarine":           "#7FFFD4",
	"Azure":                "#F0FFFF",
	"Beige":                "#F5F5DC",
	"Bisque":               "#FFE4C4",
	"Black":                "#000000",
	"BlanchedAlmond":       "#FFEBCD",
	"Blue":                 "#0000FF",
	"BlueViolet":           "#8A2BE2",
	"Brown":                "#A52A2A",
	"BurlyWood":            "#DEB887",
	"CadetBlue":            "#5F9EA0",
	"Chartreuse":           "#7FFF00",
	"Chocolate":            "#D2691E",
	"Coral":                "#FF7F50",
	"CornflowerBlue":       "#6495ED",
	"Cornsilk":             "#FFF8DC",
	"Crimson":              "#DC143C",
	"Cyan":                 "#00FFFF",
	"DarkBlue":             "#00008B",
	"DarkCyan":             "#008B8B",
	"DarkGoldenRod":        "#B8860B",
	"DarkGray":             "#A9A9A9",
	"DarkGrey":             "#A9A9A9",
	"DarkGreen":            "#006400",
	"DarkKhaki":            "#BDB76B",
	"DarkMagenta":          "#8B008B",
	"DarkOliveGreen":       "#556B2F",
	"Darkorange":           "#FF8C00",
	"DarkOrchid":           "#9932CC",
	"DarkRed":              "#8B0000",
	"DarkSalmon":           "#E9967A",
	"DarkSeaGreen":         "#8FBC8F",
	"DarkSlateBlue":        "#483D8B",
	"DarkSlateGray":        "#2F4F4F",
	"DarkSlateGrey":        "#2F4F4F",
	"DarkTurquoise":        "#00CED1",
	"DarkViolet":           "#9400D3",
	"DeepPink":             "#FF1493",
	"DeepSkyBlue":          "#00BFFF",
	"DimGray":              "#696969",
	"DimGrey":              "#696969",
	"DodgerBlue":           "#1E90FF",
	"FireBrick":            "#B22222",
	"FloralWhite":          "#FFFAF0",
	"ForestGreen":          "#228B22",
	"Fuchsia":              "#FF00FF",
	"Gainsboro":            "#DCDCDC",
	"GhostWhite":           "#F8F8FF",
	"Gold":                 "#FFD700",
	"GoldenRod":            "#DAA520",
	"Gray":                 "#808080",
	"Grey":                 "#808080",
	"Green":                "#008000",
	"GreenYellow":          "#ADFF2F",
	"HoneyDew":             "#F0FFF0",
	"HotPink":              "#FF69B4",
	"IndianRed":            "#CD5C5C",
	"Indigo":               "#4B0082",
	"Ivory":                "#FFFFF0",
	"Khaki":                "#F0E68C",
	"Lavender":             "#E6E6FA",
	"LavenderBlush":        "#FFF0F5",
	"LawnGreen":            "#7CFC00",
	"LemonChiffon":         "#FFFACD",
	"LightBlue":            "#ADD8E6",
	"LightCoral":           "#F08080",
	"LightCyan":            "#E0FFFF",
	"LightGoldenRodYellow": "#FAFAD2",
	"LightGray":            "#D3D3D3",
	"LightGrey":            "#D3D3D3",
	"LightGreen":           "#90EE90",
	"LightPink":            "#FFB6C1",
	"LightSalmon":          "#FFA07A",
	"LightSeaGreen":        "#20B2AA",
	"LightSkyBlue":         "#87CEFA",
	"LightSlateGray":       "#778899",
	"LightSlateGrey":       "#778899",
	"LightSteelBlue":       "#B0C4DE",
	"LightYellow":          "#FFFFE0",
	"Lime":                 "#00FF00",
	"LimeGreen":            "#32CD32",
	"Linen":                "#FAF0E6",
	"Magenta":              "#FF00FF",
	"Maroon":               "#800000",
	"MediumAquaMarine":     "#66CDAA",
	"MediumBlue":           "#0000CD",
	"MediumOrchid":         "#BA55D3",
	"MediumPurple":         "#9370D8",
	"MediumSeaGreen":       "#3CB371",
	"MediumSlateBlue":      "#7B68EE",
	"MediumSpringGreen":    "#00FA9A",
	"MediumTurquoise":      "#48D1CC",
	"MediumVioletRed":      "#C71585",
	"MidnightBlue":         "#191970",
	"MintCream":            "#F5FFFA",
	"MistyRose":            "#FFE4E1",
	"Moccasin":             "#FFE4B5",
	"NavajoWhite":          "#FFDEAD",
	"Navy":                 "#000080",
	"OldLace":              "#FDF5E6",
	"Olive":                "#808000",
	"OliveDrab":            "#6B8E23",
	"Orange":               "#FFA500",
	"OrangeRed":            "#FF4500",
	"Orchid":               "#DA70D6",
	"PaleGoldenRod":        "#EEE8AA",
	"PaleGreen":            "#98FB98",
	"PaleTurquoise":        "#AFEEEE",
	"PaleVioletRed":        "#D87093",
	"PapayaWhip":           "#FFEFD5",
	"PeachPuff":            "#FFDAB9",
	"Peru":                 "#CD853F",
	"Pink":                 "#FFC0CB",
	"Plum":                 "#DDA0DD",
	"PowderBlue":           "#B0E0E6",
	"Purple":               "#800080",
	"Red":                  "#FF0000",
	"RosyBrown":            "#BC8F8F",
	"RoyalBlue":            "#4169E1",
	"SaddleBrown":          "#8B4513",
	"Salmon":               "#FA8072",
	"SandyBrown":           "#F4A460",
	"SeaGreen":             "#2E8B57",
	"SeaShell":             "#FFF5EE",
	"Sienna":               "#A0522D",
	"Silver":               "#C0C0C0",
	"SkyBlue":              "#87CEEB",
	"SlateBlue":            "#6A5ACD",
	"SlateGray":            "#708090",
	"SlateGrey":            "#708090",
	"Snow":                 "#FFFAFA",
	"SpringGreen":          "#00FF7F",
	"SteelBlue":            "#4682B4",
	"Tan":                  "#D2B48C",
	"Teal":                 "#008080",
	"Thistle":              "#D8BFD8",
	"Tomato":               "#FF6347",
	"Turquoise":            "#40E0D0",
	"Violet":               "#EE82EE",
	"Wheat":                "#F5DEB3",
	"White":                "#FFFFFF",
	"WhiteSmoke":           "#F5F5F5",
	"Yellow":               "#FFFF00",
	"YellowGreen":          "#9ACD32",
}

type ColourNameSource struct {
	// If you were writing a source that needed to store some state or config,
	// you would store that config in here. Since this is just static and
	// doesn't connect to any other systems that might warrant configuration,
	// we'll leave this blank
}

// Type The type of items that this source is capable of finding
func (s *ColourNameSource) Type() string {
	return "colour"
}

// Descriptive name for the source, used in logging and metadata
func (s *ColourNameSource) Name() string {
	return "colour-name"
}

// List of contexts that this source is capable of find items for. If the
// source supports all contexts the special value `AllContexts` ("*")
// should be used
func (s *ColourNameSource) Contexts() []string {
	// Names of colours are globally unique, there isn't a difference between
	// "red" in one context and "red" in another context as they are defined by
	// w3.org: https://www.w3.org/TR/css-color-4/
	//
	// Some types of items have a specific context, for example a user named
	// "admin" on one computer isn't the same user as a user named "admin" on a
	// different computer, even though they have the same name, they have a
	// different context.
	return []string{
		"global", // This is a reserved word meaning that the items should be considered globally unique
	}
}

// Get Get a single item with a given context and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *ColourNameSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != "global" {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: "colours are only supported in the 'global' context",
			Context:     itemContext,
		}
	}

	// NOTE: In this source there isn't anything that we need to pass `ctx` too
	// as it is just manipulating objects in memory and will be extremely fast.
	// If however we needed to run an external command or make a call over the
	// network, we should make sure to re-use the ctx object when making those
	// calls to ensure that they will be calls to ensure that they will be
	// stopped if the query is cancelled or reaches a timeout

	// Look up the colour in the database
	hexValue, found := Colours[query]

	if !found {
		// If it wasn't found then return an error
		//
		// Sources should return errors of type sdp.ItemRequestError. Details of
		// what these errors should contain can be found in the SDP
		// documentation: https://github.com/overmindtech/sdp#errors
		//
		// Or the Go Docs:
		// https://pkg.go.dev/github.com/overmindtech/sdp-go#ItemRequestError_ErrorType
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("colour %v not recognized", query),
			Context:     "global",
		}
	}

	// Convert the attributes from a golang map, to the structure required for
	// the SDP protocol
	attributes, err := sdp.ToAttributes(map[string]interface{}{
		"name": query,
		"hex":  hexValue,
	})

	// Return a OTHER error of something goes wrong here
	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     "global",
		}
	}

	// Finally construct the item
	item := sdp.Item{
		Type:            "colour",
		UniqueAttribute: "name",
		Attributes:      attributes,
		Context:         "global",
		// If this item had linked items we would supply requests that the user
		// (or system) could execute here. However for the purposes of this
		// example we are going to say that colours aren't "related" to each
		// other at all
		LinkedItemRequests: []*sdp.ItemRequest{},
	}

	return &item, nil
}

// Find Finds all items in a given context
func (s *ColourNameSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	// Loop over all the colours and use a Get() request to resolve them to an
	// item
	for name := range Colours {
		item, err := s.Get(ctx, itemContext, name)

		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// sen on, so the one with the higher weight value will win.
func (s *ColourNameSource) Weight() int {
	return 100
}

// Hidden Reports whether the items returned by this source should be marked as
// hidden in their metadata. Hidden items won't be shown in GUIs or stored in
// the database. This method is optional and since most sources aren't going to
// be hidden, you usually don't need it and it can simply be removed.
//
// Hidden sources will also be excluded from requests involving wildcards
//
// Uncomment this if you need the source to be hidden, if you're not sure, just
// delete this whole comment section
//
// func (s *ColourNameSource) Hidden() bool {
//  return false
// }
