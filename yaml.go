package jsondiff

func (ctx *context) yamlQuote(str string) string {
	return str
}

func (ctx *context) yamlWriteStartList() {
	ctx.newline("")
}

func (ctx *context) yamlWriteEndList() {
	return
}

func (ctx *context) yamlWriteListItem(i int, v interface{}, len int) {
	ctx.buf.WriteString("- ")
	ctx.writeValue(v, true)
	if i != len-1 {
		ctx.newline("")
	} else {
		ctx.level--
	}
}

func (ctx *context) yamlWriteStartMap() {
	ctx.level++
	ctx.newline("")
}

func (ctx *context) yamlWriteEndMap() {
	return
}

func (ctx *context) yamlWriteMapItem(i int, k string, v interface{}, len int) {
	ctx.key(k)
	ctx.writeValue(v, true)
	if i != len-1 {
		ctx.newline("")
	}
}

func (ctx *context) yamlWriteEndOfItem(i, len int) {
	if i != len-1 {
		ctx.newline("")
	} else {
		ctx.level--
	}
}

func (ctx *context) yamlWriteStartListItem() {
	ctx.buf.WriteString("- ")
}
