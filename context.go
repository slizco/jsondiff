package jsondiff

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
)

type context struct {
	opts    *Options
	buf     bytes.Buffer
	level   int
	lastTag *Tag
	diff    Difference
}

func (ctx *context) Output() string {
	return ctx.opts.Output
}

func (ctx *context) newline(s string) {
	ctx.buf.WriteString(s)
	if ctx.lastTag != nil {
		ctx.buf.WriteString(ctx.lastTag.End)
	}
	ctx.buf.WriteString("\n")
	ctx.buf.WriteString(ctx.opts.Prefix)
	for i := 0; i < ctx.level; i++ {
		ctx.buf.WriteString(ctx.opts.Indent)
	}
	if ctx.lastTag != nil {
		ctx.buf.WriteString(ctx.lastTag.Begin)
	}
}

func (ctx *context) key(k string) {
	ctx.buf.WriteString(ctx.quote(k))
	ctx.buf.WriteString(": ")
}

func (ctx *context) writeValue(v interface{}, full bool) {
	switch vv := v.(type) {
	case bool:
		ctx.buf.WriteString(strconv.FormatBool(vv))
	case json.Number:
		ctx.buf.WriteString(string(vv))
	case string:
		vv = ctx.quote(vv)
		ctx.buf.WriteString(vv)
	case []interface{}:
		if full {
			ctx.writeList(vv)
		} else {
			ctx.buf.WriteString("[]")
		}
	case map[string]interface{}:
		if full {
			ctx.writeMap(vv)
		} else {
			ctx.buf.WriteString("{}")
		}
	default:
		ctx.buf.WriteString("null")
	}

	ctx.writeTypeMaybe(v)
}

func (ctx *context) writeTypeMaybe(v interface{}) {
	if ctx.opts.PrintTypes {
		ctx.buf.WriteString(" ")
		ctx.writeType(v)
	}
}

func (ctx *context) writeType(v interface{}) {
	switch v.(type) {
	case bool:
		ctx.buf.WriteString("(boolean)")
	case json.Number:
		ctx.buf.WriteString("(number)")
	case string:
		ctx.buf.WriteString("(string)")
	case []interface{}:
		ctx.buf.WriteString("(array)")
	case map[string]interface{}:
		ctx.buf.WriteString("(object)")
	default:
		ctx.buf.WriteString("(null)")
	}
}

func (ctx *context) writeMismatch(a, b interface{}) {
	ctx.writeValue(a, false)
	ctx.buf.WriteString(" => ")
	ctx.writeValue(b, false)
}

func (ctx *context) tag(tag *Tag) {
	if ctx.lastTag == tag {
		return
	} else if ctx.lastTag != nil {
		ctx.buf.WriteString(ctx.lastTag.End)
	}
	ctx.buf.WriteString(tag.Begin)
	ctx.lastTag = tag
}

func (ctx *context) result(d Difference) {
	if d == NoMatch {
		ctx.diff = NoMatch
	} else if d == SupersetMatch && ctx.diff != NoMatch {
		ctx.diff = SupersetMatch
	} else if ctx.diff != NoMatch && ctx.diff != SupersetMatch {
		ctx.diff = FullMatch
	}
}

func (ctx *context) quote(str string) string {
	switch ctx.Output() {
	case JSON:
		return ctx.jsonQuote(str)
	case YAML:
		return ctx.yamlQuote(str)
	default:
		return str
	}
}

func (ctx *context) writeStartList() {
	switch ctx.Output() {
	case JSON:
		ctx.jsonWriteStartList()
	case YAML:
		ctx.yamlWriteStartList()
	default:
		return
	}
}

func (ctx *context) writeEndList() {
	switch ctx.Output() {
	case JSON:
		ctx.jsonWriteEndList()
	case YAML:
		ctx.yamlWriteEndList()
	default:
		return
	}
}

func (ctx *context) writeListItem(i int, v interface{}, len int) {
	switch ctx.Output() {
	case JSON:
		ctx.jsonWriteListItem(i, v, len)
	case YAML:
		ctx.yamlWriteListItem(i, v, len)
	default:
		return
	}
}

func (ctx *context) writeStartMap() {
	switch ctx.Output() {
	case JSON:
		ctx.jsonWriteStartMap()
	case YAML:
		ctx.yamlWriteStartMap()
	default:
		return
	}
}

func (ctx *context) writeEndMap() {
	switch ctx.Output() {
	case JSON:
		ctx.jsonWriteEndMap()
	case YAML:
		ctx.yamlWriteEndMap()
	default:
		return
	}
}

func (ctx *context) writeMapItem(i int, k string, v interface{}, len int) {
	switch ctx.Output() {
	case JSON:
		ctx.jsonWriteMapItem(i, k, v, len)
	case YAML:
		ctx.yamlWriteMapItem(i, k, v, len)
	default:
		return
	}
}

func (ctx *context) writeEndOfItem(i, len int) {
	switch ctx.Output() {
	case JSON:
		ctx.jsonWriteEndOfItem(i, len)
	case YAML:
		ctx.yamlWriteEndOfItem(i, len)
	default:
		return
	}
}

func (ctx *context) writeStartListItem() {
	switch ctx.Output() {
	case JSON:
		ctx.jsonWriteStartListItem()
	case YAML:
		ctx.yamlWriteStartListItem()
	default:
		return
	}
}

func (ctx *context) writeList(vv []interface{}) {
	if len(vv) == 0 {
		// if list is empty, write []
		// and return
		ctx.buf.WriteString("[]")
		return
	}

	// start the list
	ctx.writeStartList()

	for i, v := range vv {
		// write list items
		ctx.writeListItem(i, v, len(vv))
	}

	// write end of list
	ctx.writeEndList()

}

func (ctx *context) writeMap(vv map[string]interface{}) {
	if len(vv) == 0 {
		// if map is empty, write {}
		// and return
		ctx.buf.WriteString("{}")
		return
	}

	ctx.writeStartMap()

	i := 0
	for k, v := range vv {
		ctx.writeMapItem(i, k, v, len(vv))
		i++
	}
	ctx.writeEndMap()
}

func (ctx *context) printSliceDiff(a interface{}, b interface{}) {
	sa, sb := a.([]interface{}), b.([]interface{})
	salen, sblen := len(sa), len(sb)
	max := salen
	if sblen > max {
		max = sblen
	}
	ctx.tag(&ctx.opts.Normal)
	if max == 0 {
		// if list is empty, write []
		// record type
		// and return
		ctx.buf.WriteString("[]")
		ctx.writeTypeMaybe(a)
		return
	}

	// start list
	ctx.writeStartList()

	for i := 0; i < max; i++ {
		ctx.writeStartListItem()

		if i < salen && i < sblen {
			// print change in items
			ctx.printDiff(sa[i], sb[i])
		} else if i < salen {
			// print item that has been removed
			ctx.tag(&ctx.opts.Removed)
			ctx.writeValue(sa[i], true)
			ctx.result(SupersetMatch)
		} else if i < sblen {
			// print item that has been added
			ctx.tag(&ctx.opts.Added)
			ctx.writeValue(sb[i], true)
			ctx.result(NoMatch)
		}
		ctx.tag(&ctx.opts.Normal)
		ctx.writeEndOfItem(i, max)
	}
	ctx.writeEndList()

	ctx.writeTypeMaybe(a)
}

func (ctx *context) printMapDiff(a interface{}, b interface{}) {
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
	ctx.tag(&ctx.opts.Normal)
	if len(keys) == 0 {
		// if list is empty, write {}
		// record type
		// and return
		ctx.buf.WriteString("{}")
		ctx.writeTypeMaybe(a)
		return
	}

	ctx.writeStartMap()

	for i, k := range keys {
		va, aok := ma[k]
		vb, bok := mb[k]
		if aok && bok {
			// print change in items
			ctx.key(k)
			ctx.printDiff(va, vb)
		} else if aok {
			// print item that has been removed
			ctx.tag(&ctx.opts.Removed)
			ctx.key(k)
			ctx.writeValue(va, true)
			ctx.result(SupersetMatch)
		} else if bok {
			// print item that has been added
			ctx.tag(&ctx.opts.Added)
			ctx.key(k)
			ctx.writeValue(vb, true)
			ctx.result(NoMatch)
		}
		ctx.tag(&ctx.opts.Normal)
		ctx.writeEndOfItem(i, len(keys))
	}
	ctx.writeEndMap()

	ctx.writeTypeMaybe(a)
}

func (ctx *context) printMismatch(a, b interface{}) {
	ctx.tag(&ctx.opts.Changed)
	ctx.writeMismatch(a, b)
}

func (ctx *context) printDiff(a, b interface{}) {
	if a == nil || b == nil {
		if a == nil && b == nil {
			ctx.tag(&ctx.opts.Normal)
			ctx.writeValue(a, false)
			ctx.result(FullMatch)
		} else {
			ctx.printMismatch(a, b)
			ctx.result(NoMatch)
		}
		return
	}

	ka := reflect.TypeOf(a).Kind()
	kb := reflect.TypeOf(b).Kind()
	if ka != kb {
		ctx.printMismatch(a, b)
		ctx.result(NoMatch)
		return
	}
	switch ka {
	case reflect.Bool:
		if a.(bool) != b.(bool) {
			ctx.printMismatch(a, b)
			ctx.result(NoMatch)
			return
		}
	case reflect.String:
		switch aa := a.(type) {
		case json.Number:
			bb, ok := b.(json.Number)
			if !ok || aa != bb {
				ctx.printMismatch(a, b)
				ctx.result(NoMatch)
				return
			}
		case string:
			bb, ok := b.(string)
			if !ok || aa != bb {
				ctx.printMismatch(a, b)
				ctx.result(NoMatch)
				return
			}
		}
	case reflect.Slice:
		ctx.printSliceDiff(a, b)
		return
	case reflect.Map:
		ctx.printMapDiff(a, b)
		return
	}
	ctx.tag(&ctx.opts.Normal)
	ctx.writeValue(a, true)
	ctx.result(FullMatch)
}
