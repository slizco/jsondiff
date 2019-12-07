package jsondiff

import (
	"sort"
)

func (ctx *context) writeStartList() {
	if !ctx.Y() {
		// JSON
		ctx.newline("[")
	} else {
		ctx.newline("")
	}
}

func (ctx *context) writeEndList() {
	if !ctx.Y() {
		// JSON
		ctx.buf.WriteString("]")
	}
}

func (ctx *context) writeListItem(i int, v interface{}, len int) {
	if ctx.Y() {
		// YAML
		ctx.buf.WriteString("- ")
		ctx.writeValue(v, true)
		if i != len-1 {
			ctx.newline("")
		} else {
			ctx.level--
		}
	} else {
		// JSON
		ctx.writeValue(v, true)
		if i != len-1 {
			ctx.newline(",")
		} else {
			ctx.level--
			ctx.newline("")
		}
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

func (ctx *context) writeStartMap() {
	ctx.level++
	if !ctx.Y() {
		// JSON
		ctx.newline("{")
	} else {
		ctx.newline("")
	}
}

func (ctx *context) writeEndMap() {
	if !ctx.Y() {
		// JSON
		ctx.newline("")
		ctx.buf.WriteString("}")
	}
}

func (ctx *context) writeMapItem(i int, k string, v interface{}, len int) {
	ctx.key(k)
	ctx.writeValue(v, true)
	if ctx.Y() {
		// YAML
		if i != len-1 {
			ctx.newline("")
		}
	} else {
		// JSON
		if i != len-1 {
			ctx.newline(",")
		} else {
			ctx.level--
			ctx.newline("")
		}
	}
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

func (ctx *context) printEndOfItem(i, len int) {
	if i != len-1 {
		if !ctx.Y() {
			// JSON
			ctx.newline(",")
		} else {
			// YAML
			ctx.newline("")
		}
	} else {
		ctx.level--
	}
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
		return
	}

	// start list
	ctx.level++
	ctx.writeStartList()

	for i := 0; i < max; i++ {
		if ctx.Y() {
			ctx.buf.WriteString("- ")
		}
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
		ctx.printEndOfItem(i, max)
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
		ctx.printEndOfItem(i, len(keys))
	}
	ctx.writeEndMap()

	ctx.writeTypeMaybe(a)
}
