package jsondiff

import (
	"strconv"
)

func (ctx *context) jsonQuote(str string) string {
	return strconv.Quote(str)
}

func (ctx *context) jsonWriteStartList() {
	ctx.level++
	ctx.newline("[")
}

func (ctx *context) jsonWriteEndList() {
	ctx.buf.WriteString("]")
}

func (ctx *context) jsonWriteListItem(i int, v interface{}, len int) {
	ctx.writeValue(v, true)
	ctx.jsonWriteEndOfItem(i, len)
}

func (ctx *context) jsonWriteStartMap() {
	ctx.level++
	ctx.newline("{")
}

func (ctx *context) jsonWriteEndMap() {
	ctx.buf.WriteString("}")
}

func (ctx *context) jsonWriteMapItem(i int, k string, v interface{}, len int) {
	ctx.key(k)
	ctx.writeValue(v, true)
	ctx.jsonWriteEndOfItem(i, len)
}

func (ctx *context) jsonWriteEndOfItem(i, len int) {
	if i != len-1 {
		ctx.newline(",")
	} else {
		ctx.level--
		ctx.newline("")
	}
}

func (ctx *context) jsonWriteStartListItem() {
	return
}
