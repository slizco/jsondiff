package jsondiff

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
)

type Printer interface {

	// For setting context
	PrintNormalTag()
	PrintChangedTag()
	PrintAddedTag()
	PrintRemovedTag()
	PrintLastTag()
	SetResult(Difference)
	IncrementLevel()
	WriteTypeMaybe(interface{})

	// Get the difference
	Diff() Difference
	// Get the string diff
	String() string

	// For printing newline
	Newline(string)

	// For printing key
	PrintKey(string)
	// First line printed
	PrintFirst()
	// Write a string to buffer
	WriteString(string)

	// For printing lists
	PrintStartList()
	PrintEndList()
	PrintStartListItem()
	PrintListItem(interface{}, bool)

	// For printing maps
	PrintStartMap()
	PrintEndMap()
	PrintMapItem(string, interface{}, bool)
	PrintEndOfItem(bool)
}

func NewPrinter(opts *Options) Printer {
	if opts.Output == YAML {
		return &yamlPrinter{ctx: &context{opts: opts}}
	}
	return &jsonPrinter{ctx: &context{opts: opts}}
}

// General printing helpers
func printDiff(p Printer, a, b interface{}) {
	// print any starting lines/characters/messages
	p.PrintFirst()

	// print diff of a and b
	printDiffBody(p, a, b)
}

func printDiffBody(p Printer, a, b interface{}) {

	if a == nil || b == nil {
		if a == nil && b == nil {
			p.PrintNormalTag()
			printValue(p, a, false)
			p.SetResult(FullMatch)
		} else {
			printChanged(p, a, b)
			p.SetResult(NoMatch)
		}
		return
	}

	ka := reflect.TypeOf(a).Kind()
	kb := reflect.TypeOf(b).Kind()
	if ka != kb {
		printChanged(p, a, b)
		//p.Context().result(NoMatch)
		p.SetResult(NoMatch)
		return
	}
	switch ka {
	case reflect.Bool:
		if a.(bool) != b.(bool) {
			printChanged(p, a, b)
			//p.Context().result(NoMatch)
			p.SetResult(NoMatch)
			return
		}
	case reflect.String:
		switch aa := a.(type) {
		case json.Number:
			bb, ok := b.(json.Number)
			if !ok || aa != bb {
				printChanged(p, a, b)
				//p.Context().result(NoMatch)
				p.SetResult(NoMatch)
				return
			}
		case string:
			bb, ok := b.(string)
			if !ok || aa != bb {
				printChanged(p, a, b)
				//p.Context().result(NoMatch)
				p.SetResult(NoMatch)
				return
			}
		}
	case reflect.Slice:
		printSliceDiff(p, a, b)
		return
	case reflect.Map:
		printMapDiff(p, a, b)
		return
	}
	//p.Context().tag(&p.Context().opts.Normal)
	p.PrintNormalTag()
	printValue(p, a, true)
	//p.Context().result(FullMatch)
	p.SetResult(FullMatch)
}

func printChanged(p Printer, a, b interface{}) {
	//p.Context().tag(&p.Context().opts.Changed)
	p.PrintChangedTag()
	printMismatch(p, a, b)
}

func printMismatch(p Printer, a, b interface{}) {
	printValue(p, a, false)
	// p.WriteString(" => ")
	p.WriteString(" => ")
	printValue(p, b, false)
}

func printValue(p Printer, v interface{}, full bool) {
	switch vv := v.(type) {
	case bool:
		p.WriteString(strconv.FormatBool(vv))
	case json.Number:
		p.WriteString(string(vv))
	case string:
		vv = strconv.Quote(vv)
		p.WriteString(vv)
	case []interface{}:
		if full {
			printList(p, vv)
		} else {
			p.WriteString("[]")
		}
	case map[string]interface{}:
		if full {
			printMap(p, vv)
		} else {
			p.WriteString("{}")
		}
	default:
		p.WriteString("null")
	}

	p.WriteTypeMaybe(v)
}

func printList(p Printer, vv []interface{}) {
	if len(vv) == 0 {
		// if list is empty, write []
		// and return
		p.WriteString("[]")
		return
	}

	// start the list
	p.PrintStartList()

	for i, v := range vv {
		// print list items
		p.PrintListItem(v, i == len(vv)-1)
	}

	// print end of list
	p.PrintEndList()

}

func printMap(p Printer, vv map[string]interface{}) {
	if len(vv) == 0 {
		// if map is empty, write {}
		// and return
		p.WriteString("{}")
		return
	}

	p.PrintStartMap()

	i := 0
	for k, v := range vv {
		p.PrintMapItem(k, v, i == len(vv)-1)
		i++
	}
	p.PrintEndMap()
}

func printSliceDiff(p Printer, a interface{}, b interface{}) {
	sa, sb := a.([]interface{}), b.([]interface{})
	salen, sblen := len(sa), len(sb)
	max := salen
	if sblen > max {
		max = sblen
	}
	p.PrintNormalTag()
	if max == 0 {
		// if list is empty, write []
		// record type
		// and return
		p.WriteString("[]")
		return
	}

	p.IncrementLevel()
	// start list
	p.PrintStartList()

	for i := 0; i < max; i++ {
		p.PrintStartListItem()

		if i < salen && i < sblen {
			// print change in items
			printDiffBody(p, sa[i], sb[i])
		} else if i < salen {
			// print item that has been removed
			p.PrintRemovedTag()
			printValue(p, sa[i], true)
			p.SetResult(SupersetMatch)
		} else if i < sblen {
			// print item that has been added
			p.PrintAddedTag()
			printValue(p, sb[i], true)
			p.SetResult(NoMatch)
		}
		p.PrintNormalTag()
		p.PrintEndOfItem(i == max-1)
	}
	p.PrintEndList()

	p.WriteTypeMaybe(a)
}

func printMapDiff(p Printer, a interface{}, b interface{}) {
	ma, mb := a.(map[string]interface{}), b.(map[string]interface{})
	keysMap := make(map[string]bool)
	for k := range ma {
		keysMap[k] = true
	}
	for k := range mb {
		keysMap[k] = true
	}
	keys := make([]string, 0, len(keysMap))
	for k := range keysMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	p.PrintNormalTag()
	if len(keys) == 0 {
		// if list is empty, write {}
		// record type
		// and return
		p.WriteString("{}")
		p.WriteTypeMaybe(a)
		return
	}

	p.PrintStartMap()

	for i, k := range keys {
		va, aok := ma[k]
		vb, bok := mb[k]
		if aok && bok {
			// print change in items
			p.PrintKey(k)
			printDiffBody(p, va, vb)
		} else if aok {
			// print item that has been removed
			p.PrintRemovedTag()
			p.PrintKey(k)
			printValue(p, va, true)
			p.SetResult(SupersetMatch)
		} else if bok {
			// print item that has been added
			p.PrintAddedTag()
			p.PrintKey(k)
			printValue(p, vb, true)
			p.SetResult(NoMatch)
		}
		p.PrintNormalTag()
		p.PrintEndOfItem(i == len(keys)-1)
	}
	p.PrintEndMap()

	p.WriteTypeMaybe(a)
}

// JSON-specific printing helpers

type jsonPrinter struct {
	ctx *context
}

func (p *jsonPrinter) Context() *context {
	return p.ctx
}

func (p *jsonPrinter) PrintNormalTag() {
	p.ctx.tag(&p.ctx.opts.Normal)
}

func (p *jsonPrinter) PrintChangedTag() {
	p.ctx.tag(&p.ctx.opts.Changed)
}

func (p *jsonPrinter) PrintAddedTag() {
	p.ctx.tag(&p.ctx.opts.Added)
}

func (p *jsonPrinter) PrintRemovedTag() {
	p.ctx.tag(&p.ctx.opts.Removed)
}

func (p *jsonPrinter) PrintLastTag() {
	if p.ctx.lastTag != nil {
		p.ctx.buf.WriteString(p.ctx.lastTag.End)
	}
}

func (p *jsonPrinter) SetResult(r Difference) {
	p.ctx.result(r)
}

func (p *jsonPrinter) WriteString(str string) {
	p.ctx.buf.WriteString(str)
}

func (p *jsonPrinter) WriteTypeMaybe(v interface{}) {
	if p.ctx.opts.PrintTypes {
		p.ctx.buf.WriteString(" ")
		p.ctx.writeType(v)
	}
}

func (p *jsonPrinter) IncrementLevel() {
	p.ctx.level++
}

func (p *jsonPrinter) Diff() Difference {
	return p.ctx.diff
}

func (p *jsonPrinter) String() string {
	return p.ctx.buf.String()
}

func (p *jsonPrinter) Newline(s string) {
	p.ctx.buf.WriteString(s)
	if p.ctx.lastTag != nil {
		p.ctx.buf.WriteString(p.ctx.lastTag.End)
	}
	p.ctx.buf.WriteString("\n")
	p.ctx.buf.WriteString(p.ctx.opts.Prefix)
	for i := 0; i < p.ctx.level; i++ {
		p.ctx.buf.WriteString(p.ctx.opts.Indent)
	}
	if p.ctx.lastTag != nil {
		p.ctx.buf.WriteString(p.ctx.lastTag.Begin)
	}
}

func (p *jsonPrinter) PrintKey(k string) {
	p.ctx.buf.WriteString(strconv.Quote(k))
	p.ctx.buf.WriteString(": ")
}

func (p *jsonPrinter) PrintFirst() {
	return
}

func (p *jsonPrinter) PrintStartList() {
	p.Newline("[")
}

func (p *jsonPrinter) PrintEndList() {
	p.ctx.buf.WriteString("]")
}

func (p *jsonPrinter) PrintStartListItem() {
	return
}

func (p *jsonPrinter) PrintListItem(v interface{}, last bool) {
	printValue(p, v, true)
	if !last {
		p.Newline(",")
	} else {
		p.ctx.level--
		p.Newline("")
	}
}

func (p *jsonPrinter) PrintStartMap() {
	p.ctx.level++
	p.Newline("{")
}

func (p *jsonPrinter) PrintEndMap() {
	p.Newline("")
	p.ctx.buf.WriteString("}")
}

func (p *jsonPrinter) PrintMapItem(k string, v interface{}, last bool) {
	p.PrintKey(k)
	printValue(p, v, true)

	if !last {
		p.Newline(",")
	} else {
		p.ctx.level--
		p.Newline("")
	}
}

func (p *jsonPrinter) PrintEndOfItem(last bool) {
	if !last {
		p.Newline(",")
	} else {
		p.ctx.level--
	}
}

// YAML-specific printing helpers

type yamlPrinter struct {
	ctx *context
}

func (p *yamlPrinter) Context() *context {
	return p.ctx
}

func (p *yamlPrinter) PrintNormalTag() {
	p.ctx.tag(&p.ctx.opts.Normal)
}

func (p *yamlPrinter) PrintChangedTag() {
	p.ctx.tag(&p.ctx.opts.Changed)
}

func (p *yamlPrinter) PrintAddedTag() {
	p.ctx.tag(&p.ctx.opts.Added)
}

func (p *yamlPrinter) PrintRemovedTag() {
	p.ctx.tag(&p.ctx.opts.Removed)
}

func (p *yamlPrinter) PrintLastTag() {
	if p.ctx.lastTag != nil {
		p.ctx.buf.WriteString(p.ctx.lastTag.End)
	}
}

func (p *yamlPrinter) SetResult(r Difference) {
	p.ctx.result(r)
}

func (p *yamlPrinter) WriteString(str string) {
	p.ctx.buf.WriteString(str)
}

func (p *yamlPrinter) Diff() Difference {
	return p.ctx.diff
}

func (p *yamlPrinter) String() string {
	return p.ctx.buf.String()
}

func (p *yamlPrinter) Newline(s string) {
	p.ctx.buf.WriteString(s)
	if p.ctx.lastTag != nil {
		p.ctx.buf.WriteString(p.ctx.lastTag.End)
	}
	p.ctx.buf.WriteString("\n")
	p.ctx.buf.WriteString(p.ctx.opts.Prefix)
	for i := 0; i < p.ctx.level; i++ {
		p.ctx.buf.WriteString(p.ctx.opts.Indent)
	}
	if p.ctx.lastTag != nil {
		p.ctx.buf.WriteString(p.ctx.lastTag.Begin)
	}
}

func (p *yamlPrinter) WriteTypeMaybe(v interface{}) {
	if p.ctx.opts.PrintTypes {
		p.ctx.buf.WriteString(" ")
		p.ctx.writeType(v)
	}
}

func (p *yamlPrinter) PrintKey(k string) {
	p.ctx.buf.WriteString(k)
	p.ctx.buf.WriteString(": ")
}

func (p yamlPrinter) PrintFirst() {
	p.ctx.buf.WriteString("---")
}

func (p *yamlPrinter) IncrementLevel() {
	p.ctx.level++
}

func (p *yamlPrinter) PrintStartList() {
	p.Newline("")
}

func (p *yamlPrinter) PrintEndList() {
	return
}

func (p *yamlPrinter) PrintStartListItem() {
	p.WriteString("- ")
}

func (p *yamlPrinter) PrintListItem(v interface{}, last bool) {
	p.ctx.buf.WriteString("- ")
	printValue(p, v, true)
	if !last {
		p.Newline("")
	} else {
		p.ctx.level--
	}
}

func (p *yamlPrinter) PrintStartMap() {
	p.ctx.level++
	p.Newline("")
}

func (p *yamlPrinter) PrintEndMap() {
	return
}

func (p *yamlPrinter) PrintMapItem(k string, v interface{}, last bool) {
	p.PrintKey(k)
	printValue(p, v, true)
	if !last {
		p.Newline("")
	}
}

func (p *yamlPrinter) PrintEndOfItem(last bool) {
	if last {
		p.Newline("")
	} else {
		p.ctx.level--
	}
}
