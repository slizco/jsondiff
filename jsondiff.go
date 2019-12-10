package jsondiff

import (
	"bytes"
	"encoding/json"
)

type Difference int

const (
	FullMatch Difference = iota
	SupersetMatch
	NoMatch
	FirstArgIsInvalidJson
	SecondArgIsInvalidJson
	BothArgsAreInvalidJson

	// Supported output languages
	JSON = "JSON"
	YAML = "YAML"
)

func (d Difference) String() string {
	switch d {
	case FullMatch:
		return "FullMatch"
	case SupersetMatch:
		return "SupersetMatch"
	case NoMatch:
		return "NoMatch"
	case FirstArgIsInvalidJson:
		return "FirstArgIsInvalidJson"
	case SecondArgIsInvalidJson:
		return "SecondArgIsInvalidJson"
	case BothArgsAreInvalidJson:
		return "BothArgsAreInvalidJson"
	}
	return "Invalid"
}

type Tag struct {
	Begin string
	End   string
}

type Options struct {
	Output     string
	Normal     Tag
	Added      Tag
	Removed    Tag
	Changed    Tag
	Prefix     string
	Indent     string
	PrintTypes bool
}

// Provides a set of options that are well suited for console output. Options
// use ANSI foreground color escape sequences to highlight changes.
func DefaultConsoleOptions() Options {
	return Options{
		Added:   Tag{Begin: "\033[0;32m", End: "\033[0m"},
		Removed: Tag{Begin: "\033[0;31m", End: "\033[0m"},
		Changed: Tag{Begin: "\033[0;33m", End: "\033[0m"},
		Indent:  "    ",
		Output:  JSON,
	}
}

// Provides a set of options that are well suited for HTML output. Works best
// inside <pre> tag.
func DefaultHTMLOptions() Options {
	return Options{
		Added:   Tag{Begin: `<span style="background-color: #8bff7f">`, End: `</span>`},
		Removed: Tag{Begin: `<span style="background-color: #fd7f7f">`, End: `</span>`},
		Changed: Tag{Begin: `<span style="background-color: #fcff7f">`, End: `</span>`},
		Indent:  "    ",
		Output:  JSON,
	}
}

// WithYAMLOutput modifies the given options for writing YAML output
func (opts Options) WithYAMLOutput() Options {
	opts.Indent = "  " // indent by only two spaces in YAML
	opts.Output = YAML // set language to YAML
	return opts
}

// Compares two JSON documents using given options. Returns difference type and
// a string describing differences.
//
// FullMatch means provided arguments are deeply equal.
//
// SupersetMatch means first argument is a superset of a second argument. In
// this context being a superset means that for each object or array in the
// hierarchy which don't match exactly, it must be a superset of another one.
// For example:
//
//     {"a": 123, "b": 456, "c": [7, 8, 9]}
//
// Is a superset of:
//
//     {"a": 123, "c": [7, 8]}
//
// NoMatch means there is no match.
//
// The rest of the difference types mean that one of or both JSON documents are
// invalid JSON.
//
// Returned string uses a format similar to pretty printed JSON to show the
// human-readable difference between provided JSON documents. It is important
// to understand that returned format is not a valid JSON and is not meant
// to be machine readable.
func Compare(a, b []byte, opts *Options) (Difference, string) {
	var av, bv interface{}
	da := json.NewDecoder(bytes.NewReader(a))
	da.UseNumber()
	db := json.NewDecoder(bytes.NewReader(b))
	db.UseNumber()
	errA := da.Decode(&av)
	errB := db.Decode(&bv)
	if errA != nil && errB != nil {
		return BothArgsAreInvalidJson, "both arguments are invalid json"
	}
	if errA != nil {
		return FirstArgIsInvalidJson, "first argument is invalid json"
	}
	if errB != nil {
		return SecondArgIsInvalidJson, "second argument is invalid json"
	}

	if opts.Output == "" {
		// Default to JSON if not set
		opts.Output = JSON
	}

	ctx := context{opts: opts}
	ctx.printDiff(av, bv)
	if ctx.lastTag != nil {
		ctx.buf.WriteString(ctx.lastTag.End)
	}
	return ctx.diff, ctx.buf.String()
}
