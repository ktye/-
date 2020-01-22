package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
)

type c = byte
type T c
type fn struct { // name:I:IIF::body..
	name string
	src  [2]int // line, col
	rety T
	args []T
	locl []T
	sign int
	bytes.Buffer
}

const (
	I = T(0x7f)
	J = T(0x7e)
	E = T(0x7d)
	F = T(0x7c)
)
const (
	sNewl = iota
	sFnam
	sRety
	sArgs
	sLocl
	sByte
	sBody
	sCmnt
)

var typs = map[c]T{'I': I, 'F': F}
var P = I // I(wasm32) J(wasm64)

type module []fn

func main() {
	var html bool
	flag.BoolVar(&html, "html", false, "html output")
	flag.Parse()
	b := run(os.Stdin)
	if html {
		b = page(b)
	}
	os.Stdout.Write(b)
}
func run(r io.Reader) []c {
	rd := bufio.NewReader(r)
	state := sNewl
	line, char, hi := 1, 0, true
	err := func(s string) { fatal(fmt.Errorf("%d:%d: %s", line, char, s)) }
	var m module
	var f fn
	var p c
	for {
		b, e := rd.ReadByte()
		if e == io.EOF {
			if f.name != "" {
				m = append(m, f)
			}
			return m.emit()
		}
		char++
		if b == '\n' {
			line++
			char = 1
		}
		fatal(e)
		switch state {
		case sNewl:
			if b == '/' {
				state = sCmnt
			} else if b == ' ' || b == '\n' || b == '\n' {
				if f.name == "" {
					err("parse name")
				}
				state = sByte
			} else {
				if f.name != "" {
					m = append(m, f)
				}
				f = fn{name: string(b)}
				state = sFnam
			}
		case sFnam:
			if craZ(b) || (len(f.name) > 0 && cr09(b)) {
				f.name += string(b)
			} else if b == ':' {
				state = sRety
			} else {
				err("parse function name")
			}
		case sRety:
			if f.rety == 0 {
				f.rety = typs[b]
				if f.rety == 0 {
					panic("parse return type")
				}
			} else if b == ':' {
				state = sArgs
			} else {
				err("parse return type")
			}
		case sArgs:
			if t := typs[b]; t == 0 && f.args == nil {
				err("parse args")
			} else if t != 0 {
				f.args = append(f.args, t)
			} else if b == ':' {
				state = sLocl
			} else {
				err("parse args")
			}
		case sLocl:
			if t := typs[b]; t == 0 && len(f.locl) == 0 && b == '{' {
				f.src = [2]int{line, char}
				state = sBody
				space = false
			} else if t != 0 {
				f.locl = append(f.locl, t)
			} else if b == ':' {
				state = sByte
			} else {
				err("parse locals")
			}
		case sBody:
			if b == '}' {
				state = sCmnt
				f.parse()
			} else {
				f.WriteByte(b)
			}
		case sByte:
			if b == '/' && hi {
				state = sCmnt
			} else if (b == ' ' || b == '\t') && hi {
				continue
			} else if (b == '\n') && hi {
				state = sNewl
			} else if crHx(b) {
				if hi {
					p, hi = xtoc(b)<<4, false
				} else {
					p, hi = p|xtoc(b), true
					f.WriteByte(p)
					p = 0
				}
			} else {
				err("parse body")
			}
		case sCmnt:
			if b == '\n' {
				state = sNewl
			}
		default:
			err("internal parse state")
		}
	}
}
func hxb(x c) (c, c) { h := "0123456789abcdef"; return h[x>>4], h[x&0x0F] }
func cr09(c c) bool  { return c >= '0' && c <= '9' }
func craZ(c c) bool  { return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') }
func cr0Z(c c) bool  { return cr09(c) || craZ(c) }
func crHx(c c) bool  { return cr09(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') }
func xtoc(x c) c {
	switch {
	case x < ':':
		return x - '0'
	case x < 'G':
		return 10 + x - 'A'
	default:
		return 10 + x - 'a'
	}
}
func fatal(e error) {
	if e != nil {
		panic(e)
	}
}

type parser struct {
	name       string
	line, char int
	b          []byte
}

func (f fn) parse() { // parse function body
	p := parser{name: f.name, line: f.src[0], char: f.src[1], b: strip(f.Bytes())}
}
func strip(b []c) []c { // strip comments
	lines := bytes.Split(b, []c{'\n'})
	for i, l := range lines {
		space := false
		for k, c := range l {
			if c == ' ' {
				space = true
			} else if c == '/' && space {
				lines[i] = l[:k]
				break
			}
		}
	}
	b = bytes.Join(lines, []c{'\n'})
}
func (p *parser) ex()

// intermediate representation for function bodies (typed expression tree)
type expr interface {
	rt() T // result type, maybe 0
	valid() bool
	bytes() []c
}
type sequence []expr // a;b;..
type v2 struct {     // x+y unitype
	s           string
	left, right expr
}
type v1 struct { // -y
	s   string
	arg expr
}

func (s sequence) rt() T { return s[len(s)-1].rt() }
func (s sequence) valid() bool { // all but the last expressions in a sequence must have no return type
	for i, e := range s {
		if i == len(s)-1 {
			return e.rt() != 0
		} else if e.rt() != 0 {
			return false
		}
	}
	return true
}
func (s sequence) bytes() (r []c) {
	for _, e := range s {
		r = append(r, e.bytes())
	}
	return r
}
func (v v2) rt() T { return v.right.rt() }
func (v v2) valid() bool {
	t := v.left.rt()
	return t == v.right.rt() && t != 0
}
func (v v2) bytes() []c {
	op := map[s][4]c{
		"+": [4]c{0x6a, 0x7c, 0x92, 0xa0}, // add
		"-": [4]c{0x6b, 0x7d, 0x93, 0xa1}, // sub
	}
	ops, ok := op[v.s]
	if ok == false {
		panic("type")
	}
	b := ops[typidx[v.rt()]]
	if b == 0 {
		panic("type")
	}
	typidx := map[T]int{I: 0, J: 1, E: 2, F: 3}
	return append(append(v.left.bytes(), v.right.bytes()...), b)
}
func (v v1) rt() T       { return v.arg.rt() }
func (v v1) valid() bool { return v.arg.rt() != 0 }
func (v v1) bytes() []c {
	op := map[s][4]c{
		"-": [4]c{0, 0, 0x8c, 0x9a}, // neg
	}
	ops, ok := op[v.s]
	if ok == false {
		panic("type")
	}
	b := ops[typidx[v.rt()]]
	if b == 0 {
		panic("type")
	}
	typidx := map[T]int{I: 0, J: 1, E: 2, F: 3}
	return append(v.arg.bytes(), v)
}

// emit byte code
func (m module) emit() []c {
	o := bytes.NewBuffer([]c{0, 0x61, 0x73, 0x6d, 1, 0, 0, 0}) // header
	// type section(1: function signatures)
	sec := NewSection(1)
	sigs, sigv := make(map[string]int), make([]string, 0)
	for i, f := range m {
		s := string(f.sig())
		if n, o := sigs[s]; o == false {
			n = len(sigs)
			sigs[s] = n
			sigv = append(sigv, s)
			m[i].sign = n
		}
	}
	sec.cat(lebu(len(sigv)))
	for _, s := range sigv {
		sec.cat([]c(s))
	}
	sec.out(o)
	// no import section(2)
	// function section(3: function signature indexes)
	sec = NewSection(3)
	sec.cat(lebu(len(m)))
	for _, f := range m {
		sec.cat(lebu(sigs[string(f.sig())]))
	}
	sec.out(o)
	// no table section(4)
	// linear memory section(5)
	sec = NewSection(5)
	sec.cat([]c{1, 0, 1}) // 1 initial memory segment, unshared, size 1 block
	sec.out(o)
	// no global section(6)
	// export section(7)
	sec = NewSection(7)
	sec.cat(lebu(len(m))) // number of exports (all)
	for i, f := range m {
		sec.cat(lebu(len(f.name)))
		sec.cat([]c(f.name))
		sec.cat1(0) // function-export
		sec.cat(lebu(i))
	}
	sec.out(o)
	// no start section(8)
	// no element section(9)
	// code section(10)
	sec = NewSection(10)
	sec.cat(lebu(len(m))) // number of functions
	for _, f := range m {
		b := f.code()
		sec.cat(lebu(len(b)))
		sec.cat(b)
	}
	sec.out(o)
	// no data section(11)
	return o.Bytes()
}

type section struct {
	t c
	b []c
}

func NewSection(t c) section { return section{t: t} }
func (s *section) cat(b []c) { s.b = append(s.b, b...) }
func (s *section) cat1(b c)  { s.b = append(s.b, b) }
func (s *section) out(w *bytes.Buffer) {
	w.WriteByte(s.t)
	w.Write(lebu(len(s.b)))
	w.Write(s.b)
}

func (f fn) sig() (r []c) {
	r = append(r, 0x60)
	r = append(r, lebu(len(f.args))...)
	for _, t := range f.args {
		r = append(r, c(t))
	}
	r = append(r, 1)
	r = append(r, c(f.rety))
	return r
}
func (f fn) code() (r []c) {
	logf("locl %d: %v\n", len(f.locl), f.locl)
	r = append(r, f.locs()...)
	r = append(r, f.Bytes()...)
	return append(r, 0x0b)
}
func (f fn) locs() (r []c) {
	var u []T
	var n []int
	for i, t := range f.locl {
		if i > 0 && t == f.locl[i-1] {
			n[len(n)-1]++
		} else {
			u, n = append(u, t), append(n, 1)
		}
	}
	r = lebu(len(u))
	for i, t := range u {
		r = append(r, lebu(n[i])...)
		r = append(r, c(t))
	}
	return r
}

func lebu(v int) []c { // encode unsigned leb128
	var b []c
	for {
		c := uint8(v & 0x7f)
		v >>= 7
		if v != 0 {
			c |= 0x80
		}
		b = append(b, c)
		if c&0x80 == 0 {
			break
		}
	}
	return b
}

/*
func lebs(b []b, v int64) []b { // encode signed leb128
	for {
		c := uint8(v & 0x7f)
		s := uint8(v & 0x40)
		v >>= 7
		if (v != -1 || s == 0) && (v != 0 || s != 0) {
			c |= 0x80
		}
		b = append(b, c)
		if c&0x80 == 0 {
			break
		}
	}
	return b
}
*/

func log(a ...interface{})            { fmt.Fprintln(os.Stderr, a...) }
func logf(f string, a ...interface{}) { fmt.Fprintf(os.Stderr, f, a...) }
func page(wasm []c) []c {
	var b bytes.Buffer
	b.Write([]c(head))
	b.WriteString(base64.StdEncoding.EncodeToString(wasm))
	b.Write([]c(tail))
	return b.Bytes()
}

const head = `<html>
<head><meta charset="utf-8"><title>w</title></head><body><script>
var us = function(s){var r=new Uint8Array(new ArrayBuffer(s.length));for(var i=0;i<s.length;i++)r[i]=s.charCodeAt(i);return r};
var s = "`
const tail = `"
var u = us(atob(s));
(async() => {
var r=await WebAssembly.instantiate(u)
window.k=r.instance.exports
})()
// browse to file://../index.html
</script>
<pre>
run wasm from js console, e.g:
 k.add(1,2)
</pre>
</body></html>
`
