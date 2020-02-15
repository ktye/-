package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

type c = byte
type s = string
type T c
type fn struct { // name:I:IIF{body} or name.I:IIF{..} (unexported)
	name s
	ex   bool
	src  [2]int // line, col
	t    T      // return type
	args int
	locl []T
	lmap map[s]int // local index: args+locals
	sign int       // function signature index
	ast  expr
	bytes.Buffer
}
type sig struct {
	t T
	a []T
}
type module []fn

const (
	C = T(0x01) // i8
	I = T(0x7f) // i32
	J = T(0x7e) // i64
	F = T(0x7c) // f64
)

var typs = map[c]T{'C': C, 'I': I, 'J': J, 'F': F}
var tnum = map[T]int{I: 0, J: 1, F: 2} // no C
var styp = map[T]s{C: "C", I: "I", J: "J", F: "F"}
var alin = map[T]c{C: 0, I: 2, J: 3, F: 3}

func main() {
	var stdin io.Reader = os.Stdin
	var html, cout, gout bool
	var runfile string
	flag.BoolVar(&html, "html", false, "html output")
	flag.BoolVar(&cout, "c", false, "c output")
	flag.BoolVar(&gout, "go", false, "go output")
	flag.StringVar(&runfile, "r", "", "run file")
	flag.Parse()
	if runfile != "" {
		if f, e := ioutil.ReadFile(runfile); e != nil {
			panic(e)
		} else {
			stdin = bytes.NewReader(f)
		}
	}
	m, data := run(stdin)
	if html {
		os.Stdout.Write(page(m.wasm(data)))
	} else if cout {
		os.Stdout.Write(m.cout(data))
	} else if gout {
		os.Stdout.Write(m.gout(data))
	} else if runfile != "" {
		runWagon(m.wasm(data), flag.Args())
	} else {
		os.Stdout.Write(m.wasm(data))
	}
}
func (t T) String() s {
	if c := map[T]c{I: 'I', J: 'J', F: 'F'}[t]; c == 0 {
		return "0"
	} else {
		return s(c)
	}
}
func run(r io.Reader) (module, []c) {
	sFnam, sRety, sArgs, sBody, sData, sCmnt := 0, 1, 2, 3, 4, 5
	rd := bufio.NewReader(r)
	state := sFnam
	line, char := 1, 0
	err := func(s string) { panic(sf("%d:%d: %s", line, char, s)) }
	var m module
	var f fn
	var data []c
	for {
		b, e := rd.ReadByte()
		if e == io.EOF || (state == sFnam && b == '\\') {
			return m.compile(), decode(data)
		} else if e != nil {
			panic(e)
		}
		char++
		if b == '\n' {
			line++
			char = 1
		}
		switch state {
		case sFnam:
			if len(f.name) == 0 && b == ' ' || b == '\t' || b == '\n' {
				continue
			} else if len(f.name) == 0 && b == '/' {
				state = sCmnt
			} else if craZ(b) || (len(f.name) > 0 && cr09(b)) {
				f.name += s(b)
			} else if b == '.' {
				state = sRety
			} else if b == ':' {
				state = sRety
				f.ex = true
			} else {
				err("parse function name")
			}
		case sRety:
			if b == ':' {
				state = sArgs
				continue
			} else if f.t != 0 {
				err("parse return type")
			}
			if b == '{' && f.t == 0 {
				if f.name == "M" {
					state = sData
				}
				state = sBody // macro
				continue
			}
			f.t = typs[b]
			if f.t == 0 && b != '0' {
				err("parse return type")
			}
		case sArgs:
			if t := typs[b]; t == 0 && f.locl == nil {
				err("parse args")
			} else if t != 0 {
				f.locl = append(f.locl, t)
				f.args++
			} else if b == ' ' || b == '\t' {
				continue
			} else if b == '{' {
				state = sBody
				f.src = [2]int{line, char}
			} else {
				err("parse args")
			}
		case sBody:
			f.WriteByte(b)
			if b == '}' {
				state = sFnam
				m = append(m, f)
				f = fn{}
			}
		case sData:
			if b == '}' {
				state = sFnam
			} else {
				data = append(data, b)
			}
		case sCmnt:
			if b == '\n' {
				state = sFnam
			}
		default:
			err("internal parse state")
		}
	}
}
func decode(data []c) []c {
	if data != nil {
		panic("nyi")
	}
	return nil
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
func boolvar(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (m module) compile() (r module) {
	mac := make(map[s][]c)
	fns := make(map[s]int)
	var fsg []sig
	for _, f := range m {
		_, x := mac[f.name]
		_, y := fns[f.name]
		if x || y {
			panic(f.name + " already defined")
		}
		if f.args == 0 {
			b := f.Bytes()
			mac[f.name] = b[:len(b)-1] // strip '}'
		} else {
			r = append(r, f)
			n := len(r) - 1
			fns[f.name] = n
			sg := make([]T, f.args)
			copy(sg, f.locl)
			fsg = append(fsg, sig{t: f.t, a: sg})
		}
	}
	for i, f := range r {
		f.ast = f.parse(mac, fns, fsg)
		r[i] = f
	}
	return r
}

type parser struct {
	mac map[s][]c
	fns map[s]int
	fsg []sig
	*fn
	p   int
	b   []byte
	tok []byte
}

func (f *fn) parse(mac map[s][]c, fns map[s]int, fsg []sig) expr { // parse function body
	f.lmap = make(map[string]int)
	for i := 0; i < f.args; i++ {
		s := s('x' + c(i))
		if i > 2 {
			s = sf("x%d", i)
		}
		f.lmap[s] = i
	}
	p := parser{mac: mac, fns: fns, fsg: fsg, fn: f, b: strip(f.Bytes())}
	e := p.seq('}')
	e = p.locals(e, 0)
	if x, s := p.validate(e); x != nil {
		println(f.name)
		return p.xerr(x, s)
	}
	if t := e.rt(); t != p.fn.t {
		if !(t == 0 && p.fn.t == 255) {
			return p.err(sf("return type is %s not %s", t, p.fn.t))
		}
	}
	return e
}
func strip(b []c) []c { // strip comments
	lines := bytes.Split(b, []c{'\n'})
	for i, l := range lines {
		for k, c := range l {
			if c == '/' && (k == 0 || l[k-1] == ' ') {
				lines[i] = l[:k]
				break
			}
		}
	}
	return bytes.Join(lines, []c{'\n'})
}
func (p *parser) err(s s) expr {
	panic(s)
	return nil
}
func (p *parser) xerr(x expr, s s) expr {
	if i, o := x.(indicator); o {
		return p.indicate(i.indicate(), s)
	} else {
		return p.err(s)
	}
}
func (p *parser) indicate(pos int, e s) expr {
	s := s(p.fn.Bytes())
	lines := strings.Split(s, "\n")
	for _, l := range lines {
		if pos < len(l) {
			if pos > 0 {
				//pos--
			}
			return p.err("\n" + l + "\n" + strings.Repeat(" ", pos) + "^" + e)
		}
		pos -= len(l) + 1
	}
	return p.err(e)
}
func (p *parser) w() {
	for len(p.b) > 0 {
		if c := p.b[0]; c == ' ' || c == '\t' || c == '\n' {
			p.p++
			p.b = p.b[1:]
		} else {
			return
		}
	}
}
func (p *parser) t(f func([]c) int) bool { // test
	p.tok = nil
	if len(p.b) < 1 {
		return false
	}
	if n := f(p.b); n > 0 {
		p.tok = p.b[:n]
		p.b = p.b[n:]
		p.p += n
		return true
	}
	return false
}
func (p *parser) seq(term c) expr {
	var seq seq
	seq.pos = pos(p.p)
	sempos := p.p
	for {
		e := p.ex(p.noun())
		if e != nil {
			seq.argv = append(seq.argv, e)
		} else {
			p.w()
			if len(p.b) == 0 {
				p.xerr(seq, "missing "+s(term))
			}
			if p.b[0] == term {
				p.b = p.b[1:]
				break
			} else if p.b[0] != ';' {
				p.xerr(seq, "expected ;")
			} else {
				if p.p == sempos { // empty statement (for jump table)
					seq.argv = append(seq.argv, nop{})
				}
				sempos = p.p
				p.b = p.b[1:]
			}
		}
	}
	if len(seq.argv) > 1 { // suppress assignment expressions
		for i, e := range seq.argv[:len(seq.argv)-1] {
			if v, o := e.(las); o {
				v.tee = 0
				seq.argv[i] = v
			}
		}
	}
	if seq.argv == nil {
		return nil // empty?
	} else if len(seq.argv) == 1 {
		return seq.argv[0]
	}
	return seq
}
func (p *parser) ex(x expr) expr {
	if x == nil {
		return x
	}
	h := p.p
	v := p.noun()
	if op, o := x.(opx); o && s(op) == "-" { // fix neg. numbers
		if c, o := v.(con); o {
			c.i = -c.i
			c.f = -c.f
			x = c
			v = p.noun()
		}
	}
	if p.verb(x) {
		if y := p.ex(v); y == nil {
			return x // verb ?
		} else {
			return p.monadic(x, y, pos(h))
		}
	} else {
		if v == nil {
			return x // noun
		} else if p.verb(v) {
			h = p.p
			if y := p.ex(p.noun()); y == nil {
				return p.xerr(pos(h), sf("verb-verb (missing noun) x=%#v v=%#v", x, v))
			} else {
				return p.dyadic(v, x, y, pos(h))
			}
		} else if t, o := x.(typ); o {
			y := p.ex(v)
			if y == nil {
				y = v
			}
			return lod{t: t.t, argv: argv{y}, pos: pos(h)} // I x
		} else if s, o := v.(swc); o {
			s.argv = append(argv{x}, s.argv...)
			s.pos = pos(h)
			return s
		} else {
			return p.xerr(pos(h), sf("noun-noun (missing verb) %#v %#v", x, v))
		}
	}
}
func (p *parser) monadic(f, x expr, h pos) expr {
	switch v := f.(type) {
	case opx:
		if s(v) == "?" { // ?x
			return brif{argv: argv{x}, pos: h}
		}
		return v1{s: s(v), argv: argv{x}, pos: h}
	case fun:
		return cal{fun: v, argv: argv{x}}
	case asn:
		return ret{argv: argv{x}, pos: h}
	default:
		panic("nyi")
	}
}
func (p *parser) dyadic(f, x, y expr, h pos) expr {
	switch v := f.(type) {
	case asn:
		if v.opx == ":" { // memory
			return sto{argv: argv{x, y}, t: v.t, pos: h}
		} else { // local
			if v.opx != "" { // modified
				y = p.dyadic(opx(v.opx), x, y, h)
			}

			a := las{tee: 1, pos: h}
			xv, o := x.(loc)
			if o == false {
				return p.xerr(a, "assignment expects a symbol on the left")
			}
			if n, o := p.fn.lmap[xv.s]; o == false {
				xv.i = len(p.fn.lmap)
				p.fn.lmap[xv.s] = xv.i
				p.fn.locl = append(p.fn.locl, 0) // type is still unknown
			} else {
				xv.i = n
			}
			a.argv = []expr{xv, y}
			return a
		}
	case nlp:
		if a, o := x.(con); o && a.t == I && a.i == 1 { // 1/
			return whl{pos: h, argv: argv{x, untee(y)}}
		}
		return nlp{pos: h, argv: argv{x, untee(y)}}
	case opx:
		if _, o := v2Tab[s(v)]; o {
			return v2{s: s(v), argv: argv{x, y}, pos: h}
		}
		if _, o := cTab[s(v)]; o {
			return cmp{s: s(v), argv: argv{x, y}, pos: h}
		}
		if s(v) == "?" || s(v) == "?/" || s(v) == "?'" {
			if xt, o := x.(typ); o {
				sn := 0
				if s(v) == "?'" {
					sn = 1
				}
				return cvt{t: xt.t, argv: argv{y}, pos: h, sign: sn}
			}
			if s(v) == "?" {
				return iff{argv: argv{x, untee(y)}, pos: h}
			} else {
				return whl{argv: argv{x, untee(y)}, pos: h}
			}
		}
		return p.err("unknown operator(" + s(v) + ")")
	case fun:
		return cal{fun: v, argv: argv{x, y}}
	default:
		panic("nyi")
	}
}
func untee(x expr) expr {
	switch v := x.(type) {
	case las:
		v.tee = 0
		return v
	case seq:
		v.argv[len(v.argv)-1] = untee(v.argv[len(v.argv)-1])
		return v
	}
	return x
}
func (p *parser) verb(v expr) bool {
	switch v.(type) {
	case opx, nlp, asn, fun: // todo: others
		return true
	}
	return false
}
func (p *parser) noun() expr {
	p.w()
	if len(p.b) == 0 {
		return nil
	}
	switch {
	case p.t(sC('(')):
		return p.seq(')')
	case p.t(sTyp):
		return p.pTyp(p.tok)
	case p.t(sSym):
		if mc, o := p.mac[s(p.tok)]; o { // macro-expansion
			p.b = append(mc, p.b...)
			return p.noun()
		}
		return p.pSym(p.tok)
	case p.t(sC('/')):
		return nlp{}
	case p.t(sCon):
		return p.pCon(p.tok)
	case p.t(sOp):
		e := p.pOp(p.tok)
		if len(p.b) > 0 && p.b[0] == ':' { // w/o space
			p.t(sC(':'))
			t := T(0)
			if len(p.b) > 0 && p.b[0] == '\'' { // ::' (i32.store_8)
				p.t(sC('\''))
				t = C
			}
			return asn{e.(opx), t}
		} else if s(e.(opx)) == ":" {
			return asn{opx(""), 0}
		} else if s(e.(opx)) == "?" && len(p.b) > 0 && p.b[0] == '[' {
			p.t(sC('['))
			s := untee(p.seq(']').(seq)).(seq)
			return swc{argv: s.argv} // ?[a;b;..] (jump)
		}
		return e
	case p.t(sCnd):
		s := p.seq(']').(seq)
		return cnd(s)
	default:
		return nil
	}
}
func (p *parser) locals(e expr, lv int) expr {
	switch l := e.(type) {
	case las:
		l.argv[1] = p.locals(l.argv[1], lv)
		x := l.argv[0].(loc)
		if x.t == 0 {
			x.t = p.locl[x.i]
			l.argv[0] = x
		}
		yt := l.argv[1].rt()
		if yt == 0 {
			fmt.Fprintf(os.Stderr, "%#v\n", l)
			return p.xerr(e, "cannot assign zero type")
		}
		if x.t == 0 {
			p.locl[x.i] = yt
			x.t = yt
			l.argv[0] = x
		} else if x.t != yt {
			return p.xerr(e, sf("local reassignment of type %s with %s", x.t, yt))
		}
		return l
	case loc:
		if n, o := p.fn.lmap[l.s]; o {
			l.i = n
			l.t = p.locl[n]
		} else {
			return p.xerr(l, "undeclared("+l.s+")")
		}
		if l.t == 0 {
			l.t = p.locl[l.i]
		}
		return l
	case nlp:
		l.argv[0] = p.locals(l.argv[0], lv+1)
		switch x := l.argv[0].(type) {
		case las:
			l.n = x.argv[0].(loc).i
		case loc:
			l.n = x.i
		default:
			l.n = p.nloc(s('i'+lv)+"n", I) // create limit in jn ..
		}
		l.c = p.nloc(s('i'+lv), I) // set/create loop counter
		l.argv[1] = p.locals(l.argv[1], lv+1)
		return l
	case v2:
		l.argv[0] = p.locals(l.argv[0], lv)
		l.argv[1] = p.locals(l.argv[1], lv)
		t := l.rt()
		for i := 0; i < 2; i++ {
			if x, o := l.argv[i].(loc); o {
				if x.t == 0 {
					x.t = t // uninitialized local
					l.argv[i] = x
				}
			}
		}
		return l
	case sto:
		l.argv[0] = p.locals(l.argv[0], lv)
		l.argv[1] = p.locals(l.argv[1], lv)
		if l.t == 0 {
			l.t = l.argv[1].rt()
		}
		return l
	default:
		if av, o := e.(argvec); o {
			v := av.args()
			for i, a := range v {
				v[i] = p.locals(a, lv)
			}
		}
		return e
	}
}
func (p *parser) validate(e expr) (expr, s) {
	if av, o := e.(argvec); o {
		for _, e := range av.args() {
			if r, s := p.validate(e); r != nil {
				return r, s
			}
		}
	}
	s := e.valid()
	if s != "" {
		return e, s
	}
	return nil, ""
}
func (p *parser) nloc(s s, t T) int { // local index by name, may create new
	n, o := p.fn.lmap[s]
	if o {
		if p.fn.locl[n] != t {
			p.err(s + " exists with different type")
		}
	} else {
		n = len(p.fn.lmap)
		p.fn.lmap[s] = n
		p.fn.locl = append(p.fn.locl, t)
	}
	return n
}
func sTyp(b []c) int { // C I J F
	if _, o := typs[b[0]]; o == false {
		return 0
	}
	if len(b) > 0 && (craZ(b[1]) || cr09(b[1])) {
		return 0
	}
	return 1
}
func (p *parser) pTyp(b []c) expr { return typ{t: typs[b[0]]} }
func sSym(b []c) int { // [aZ][a9]*
	c := b[0]
	if craZ(c) == false {
		return 0
	}
	for i, c := range b {
		if craZ(c) == false && cr09(c) == false {
			return i
		}
	}
	return len(b)
}
func (p *parser) pSym(b []c) expr {
	if n, o := p.fns[s(b)]; o {
		return fun{s: s(b), n: n, sig: p.fsg[n], pos: pos(p.p)}
	}
	return loc{pos: pos(p.p), s: s(b), i: -1}
}
func sCon(b []c) int { // 123 123i 123j .123 123. -..
	dot := false
	if !cr09(b[0]) {
		return 0
	}
	for i, c := range b {
		if cr09(c) {
			continue
		} else if dot == false && (c == 'i' || c == 'j') {
			return i + 1
		} else if dot == false && c == '.' {
			dot = true
		} else {
			return i
		}
	}
	return len(b)
}
func (p *parser) pCon(b []c) expr {
	var r con
	if bytes.IndexByte(b, '.') != -1 {
		if f, e := strconv.ParseFloat(s(b), 64); e != nil {
			return p.err(e.Error())
		} else {
			r.t = F
			r.f = f
			return r
		}
	}
	r.t = I
	if c := b[len(b)-1]; c == 'i' || c == 'j' {
		b = b[:len(b)-1]
		if c == 'j' {
			r.t = J
		}
	}
	if i, e := strconv.ParseInt(s(b), 10, 64); e != nil {
		return p.err(e.Error())
	} else {
		r.i = i
	}
	return r
}
func sOp(b []c) int {
	if b[0] == '?' && len(b) > 1 && b[1] == '/' { // ?/ while
		return 2
	} else if b[0] == ':' || b[0] == '?' {
		return 1
	}
	for _, n := range []int{3, 2, 1} { // longest match first
		if len(b) >= n && allops[s(b[:n])] {
			return n
		}
	}
	return 0
}
func (p *parser) pOp(b []c) expr { return opx(s(b)) }
func sCnd(b []c) int {
	if len(b) > 1 && b[0] == '$' && b[1] == '[' {
		return 2
	}
	return 0
}
func sC(x c) func(b []c) int { return func(b []c) int { return boolvar(b[0] == x) } }

// intermediate representation for function bodies (typed expression tree)
type expr interface {
	rt() T    // result type, maybe 0
	valid() s // ok("") or err
	bytes() []c
}
type argvec interface {
	args() []expr
}
type argv []expr
type cstringer interface {
	cstr() s
}
type gstringer interface {
	gstr() s
}
type seq struct { // a;b;..
	pos
	argv
}
type cnd struct { // $[a;b;..]
	pos
	argv
}
type v2 struct { // x+y unitype
	pos
	argv
	s s // +-*%
}
type v1 struct { // -y
	pos
	argv
	s s
	p int
}
type cmp struct { // x<y..
	pos
	argv
	s s
}
type typ struct { // type C I J F
	pos
	t T
}
type con struct { // numeric constant
	pos
	t T
	i int64
	f float64
}
type cvt struct { // J? convert
	pos
	argv
	t    T
	sign int
}
type fun struct { // f
	pos
	s     s
	n     int
	sig   sig
	indir bool
}
type cal struct { // f x
	fun
	argv
}
type loc struct { // local get
	pos
	t T
	s s
	i int
}
type las struct { // local set
	pos
	argv
	tee c // 01
}
type sto struct { // x::y x::'y (C)
	pos
	argv
	t T
}
type lod struct { // I x  (I'x signed)
	pos
	argv
	t T
}
type ret struct { // :x (return)
	pos
	argv
}
type iff struct { // x?y
	pos
	argv
}
type swc struct { // x?[a;b;..] jump table
	pos
	argv
	// t T todo as expression with return value (how to do in c/go?)
}
type nlp struct { // x/y loop
	pos
	argv
	n int // index in locl for loop limit
	c int // index in locl for loop counter
}
type whl struct { // x?/y while  1/ while(1)
	pos
	argv
}
type brif struct { // ?x
	pos
	argv
}
type opx s        // operator
type asn struct { // assignments :(local) ::(memory) +:(modified local)
	opx
	t T // C(::')
}
type nop struct{}
type pos int // src position
type indicator interface {
	indicate() int
}

func (p pos) indicate() int { return int(p) }

func getop(tab map[s]code, op s, t T) (r c) {
	ops, ok := tab[op]
	if !ok {
		panic("unknown operator: " + op)
	}
	switch t {
	case I:
		r = ops.I
	case J:
		r = ops.J
	case F:
		r = ops.F
	default:
		panic("type(" + op + ")")
	}
	if r == 0 {
		panic("type(" + op + ")")
	}
	return r
}
func cop(tab map[s]code, op s, t T) (o, u s) {
	ops, ok := tab[op]
	if !ok {
		panic("type")
	}
	o = ops.c
	if strings.Index(o, ";") != -1 {
		v := strings.Split(o, ";")
		o = v[tnum[t]]
	}
	if o[0] == 'U' {
		if t != F {
			u = "(u" + styp[t] + ")"
		}
		o = o[1:]
	}
	return o, u
}
func gop(tab map[s]code, op s, t T) (o, u s) {
	ops, ok := tab[op]
	if !ok {
		panic("type")
	}
	o = ops.g
	if strings.Index(o, ";") != -1 {
		v := strings.Split(o, ";")
		o = v[tnum[t]]
	}
	if o[0] == 'U' {
		if t != F {
			u = "u" + styp[t]
		}
		o = o[1:]
	}
	return o, u
}

func (a argv) args() []expr { return a }
func (a argv) x() expr      { return a[0] }
func (a argv) y() expr      { return a[1] }
func (s seq) rt() T         { return s.argv[len(s.argv)-1].rt() }
func (s seq) valid() s { // all but the last expressions in a sequence must have no return type
	for i, e := range s.argv {
		if t := e.rt(); i < len(s.argv)-1 && t != 0 {
			return sf("statement %d/%d has nonzero type %s: %#v", i+1, len(s.argv), t, e)
		}
	}
	return ""
}
func (s seq) bytes() (r []c) {
	for _, e := range s.argv {
		r = append(r, e.bytes()...)
	}
	return r
}
func (s seq) cstr() (r s) {
	t := s.rt()
	if t != 0 {
		r = "("
		for i, a := range s.argv {
			r += cstring(a)
			if i < len(s.argv)-1 {
				r += ","
			}
		}
		return r + ")"
	}
	for _, a := range s.argv {
		r += cstring(a)
	}
	return r
}
func (s seq) gstr() (r s) {
	t := s.rt()
	if t != 0 {
		r = "func()" + styp[t] + "{"
	}
	for i, a := range s.argv {
		if i == len(s.argv) && t != 0 {
			r += "return "
		}
		r += gstring(a)
		if i < len(s.argv)-1 {
			r += ";"
		}
	}
	if t != 0 {
		r += "}()"
	}
	return r
}
func (v cnd) rt() T { return v.argv[len(v.argv)-1].rt() }
func (v cnd) valid() s {
	n := len(v.argv)
	if n < 3 || n%2 == 0 { // only odd are allowed (with else statement)
		return sf("conditional $[..] has wrong number of cases: %d", n)
	}
	rt := v.rt()
	for i := 0; i < n-1; i += 2 {
		if t := v.argv[i].rt(); t != I {
			return sf("conditional must be I (%s)", t)
		}
	}
	for i := 1; i < n; i += 2 {
		if v.argv[i].rt() != rt {
			return sf("conditional has mixed types")
		}
	}
	return ""
}
func (v cnd) bytes() (r []c) {
	t := v.rt()
	a := v.argv
	for i := 0; i < len(a)-1; i += 2 {
		r = catb(r, a[i].bytes(), []c{0x04, c(t)}, a[i+1].bytes(), []c{0x05})
	}
	return catb(r, a[len(a)-1].bytes(), bytes.Repeat([]c{0x0b}, len(a)/2))
}
func (v cnd) cstr() (r s) {
	s := ""
	a := v.argv
	for i := 0; i < len(a)-1; i += 2 {
		if i > 0 {
			s = ":"
		}
		r += s + cstring(a[i]) + "?" + cstring(a[i+1])
	}
	return r + ":" + cstring(a[len(a)-1])
}
func (v cnd) gstr() (r s) {
	if t := v.rt(); t == 0 {
		r = "func(){"
	} else {
		r = "func()" + styp[t] + "{"
	}
	s, a := "", v.argv
	for i := 0; i < len(a)-1; i += 2 {
		if i > 0 {
			s = "else"
		}
		r += s + " if " + gstring(a[i]) + "{return " + gstring(a[i+1]) + ";}"
	}
	return "else{" + gstring(a[len(a)-1]) + "}}()"
}
func (v v2) rt() T {
	t := v.x().rt()
	if t == 0 { // e.g. uninitialized local (r+:x)
		return v.y().rt()
	}
	return t
}
func (v v2) valid() s {
	if tx, ty := v.x().rt(), v.y().rt(); tx == 0 {
		return sf("left argument has zero type")
	} else if ty == 0 {
		fmt.Fprintf(os.Stderr, "%#v\n", v)
		return sf("right argument has zero type")
	} else if tx != ty {
		return sf("types mismatch %s %s", tx, ty)
	}
	return ""
}
func (v v2) bytes() []c {
	return append(append(v.x().bytes(), v.y().bytes()...), getop(v2Tab, v.s, v.rt()))
}
func (v v2) cstr() s  { return c2str(v2Tab, v.s, v.rt(), v.x(), v.y()) }
func (v v2) gstr() s  { return g2str(v2Tab, v.s, v.rt(), v.x(), v.y()) }
func (v v1) rt() T    { return v.x().rt() }
func (v v1) valid() s { return ifex(v.x().rt() == 0, "argument has zero type") }
func (v v1) bytes() []c {
	if t := v.rt(); v.s == "-" && t == I {
		return catb([]c{0x41, 0x00}, v.x().bytes(), []c{0x6b}) // 0-x
	} else if v.s == "-" && t == J {
		return catb([]c{0x42, 0x00}, v.x().bytes(), []c{0x7d}) // 0-x
	}
	return append(v.x().bytes(), getop(v1Tab, v.s, v.rt()))
}
func (v v1) cstr() s { o, u := cop(v1Tab, v.s, v.rt()); return jn(o, "(", u, cstring(v.x()), ")") }
func (v v1) gstr() s {
	o, u := gop(v1Tab, v.s, v.rt())
	return jn(o, u, "((", gstring(v.x()), "))")
}
func (v cmp) rt() T    { return I }
func (v cmp) valid() s { return v2(v).valid() }
func (v cmp) bytes() []c {
	return append(append(v.x().bytes(), v.y().bytes()...), getop(cTab, v.s, v.rt()))
}
func (v cmp) cstr() s  { return c2str(cTab, v.s, v.rt(), v.x(), v.y()) }
func (v cmp) gstr() s  { return g2str(cTab, v.s, v.rt(), v.x(), v.y()) }
func (v con) rt() T    { return v.t }
func (v con) valid() s { return ifex(v.t == 0, "constant has zero type") }
func (v con) bytes() (r []c) {
	r = append([]c{0x41}, lebs(int(v.i))...)
	if v.t == J {
		r[0]++
	} else if v.t == F {
		b := make([]byte, 9)
		b[0] = 0x44
		binary.LittleEndian.PutUint64(b[1:], math.Float64bits(v.f))
		return b
	}
	return r
}
func (v con) cstr() s {
	if v.t == F {
		s := sf("%v", v.f)
		if strings.Index(s, ".") == -1 {
			s += ".0"
		}
		return s
	}
	return sf("%d", v.i)
}
func (v con) gstr() s  { return v.cstr() }
func (v cvt) rt() T    { return v.t }
func (v cvt) valid() s { return ifex(v.t == 0, "convert: illegal target type") }
func (v cvt) bytes() []c {
	tab := map[T]s{
		I: "\x00\x00\xa7\xa7\xab\xaa",
		J: "\xad\xac\x00\x00\xb1\xb0",
		F: "\xb8\xb7\xba\xb9\x00\x00",
	}
	return append(v.x().bytes(), c(tab[v.t][2*tnum[v.x().rt()]+v.sign]))
}
func (v cvt) cstr() s    { return jn("(", styp[v.t], ")", cstring(v.x())) } // todo: signed?
func (v cvt) gstr() s    { return jn(styp[v.t], "(", gstring(v.x()), ")") } // todo signed
func (v typ) rt() T      { return 0 }
func (v typ) valid() s   { return "illegal type" }
func (v typ) bytes() []c { return nil }
func (v fun) rt() T      { return 0 }
func (v fun) valid() s   { return "unapplied func " + v.s }
func (v fun) bytes() []c { return nil }
func (v cal) rt() T      { return v.sig.t }
func (v cal) valid() s {
	if len(v.sig.a) != len(v.argv) {
		return sf("func %s has wrong argn: %d", v.s, len(v.argv))
	}
	for i, a := range v.argv {
		if a.rt() != v.sig.a[i] {
			return sf("func %s arg %d has wrong type", v.s, i+1)
		}
	}
	return ""
}
func (v cal) bytes() (r []c) {
	for _, a := range v.argv {
		r = append(r, a.bytes()...)
	}
	return append(append(r, 0x10), leb(v.n)...)
}
func (v cal) cstr() s {
	av, s := make([]s, len(v.argv)), ""
	for i, a := range v.argv {
		av[i] = cstring(a)
	}
	if v.sig.t == 0 {
		s = ";"
	}
	return jn(v.s, "(", strings.Join(av, ","), ")", s)
}
func (v cal) gstr() s {
	av := make([]s, len(v.argv))
	for i, a := range v.argv {
		av[i] = gstring(a)
	}
	return jn(v.s, "(", strings.Join(av, ","), ")")
}
func (v loc) rt() T      { return v.t }
func (v loc) valid() s   { return ifex(v.t == 0, "local has zero type") }
func (v loc) bytes() []c { return append([]c{0x20}, leb(v.i)...) }
func (v loc) cstr() s    { return locstr(v) }
func (v loc) gstr() s    { return locstr(v) }
func (v las) rt() T      { return T(v.tee) * v.y().rt() }
func (v las) valid() s {
	tx, ty := v.x().rt(), v.y().rt()
	return ifex(tx == 0 || tx != ty, sf("assignment with mismatched types %s %s", tx, ty))
}
func (v las) bytes() []c {
	return append(v.y().bytes(), append([]c{0x21 + v.tee}, leb(v.x().(loc).i)...)...)
}
func (v las) cstr() (r s) {
	r = jn(locstr(v.x()), "=", cstring(v.y()))
	if v.tee != 0 {
		return "(" + r + ")"
	} else {
		return r + ";"
	}
}
func (v las) gstr() s {
	if v.tee > 0 {
		return jn("as", styp[v.rt()], "(&", locstr(v.x()), ",", gstring(v.y()), ")")
	}
	return jn(locstr(v.x()), "=", s(v.y().bytes()), ";")
}
func (v lod) rt() T {
	if v.t == C {
		return I
	}
	return v.t
}
func (v lod) valid() s {
	if t := v.x().rt(); t != I {
		return sf("load has wrong addr type %s", t)
	} else if v.t == 0 {
		return sf("load has wrong store type")
	}
	return ""
}
func (v lod) bytes() (r []c) {
	op := map[T]c{C: 0x2d, I: 0x28, J: 0x29, F: 0x2b}[v.t]
	al := alin[v.t]
	return append(v.x().bytes(), []c{op, al, 0}...)
}
func (v lod) cstr() s { return jn("M", styp[v.t], "[", cstring(v.x()), gmemshift(v.t), "]") }
func (v lod) gstr() s { return jn("M.", styp[v.t], "[", gstring(v.x()), gmemshift(v.t), "]") }
func (v sto) rt() T   { return 0 }
func (v sto) valid() s {
	if yt := v.y().rt(); yt == 0 || (v.t == C && yt != I) {
		return "store: y is has wrong type"
	} else if v.x().rt() != I {
		return "store addr has wrong type"
	}
	return ""
}
func (v sto) bytes() (r []c) {
	op := map[T]c{C: 0x3a, I: 0x36, J: 0x37, F: 0x39}[v.t]
	al := alin[v.t]
	return catb(v.x().bytes(), v.y().bytes(), []c{op, al, 0})
}
func (v sto) cstr() s {
	return jn("M", styp[v.t], "[", cstring(v.x()), gmemshift(v.t), "]=(", styp[v.t], ")", cstring(v.y()), ";")
}
func (v sto) gstr() s {
	return jn("M.", styp[v.t], "[", gstring(v.x()), gmemshift(v.t), "]=", styp[v.t], "(", gstring(v.y()), ")")
}
func (v ret) rt() T      { return 0 /*v.x().rt()*/ }
func (v ret) valid() s   { return ifex(v.x().rt() == 0, "return zero type") }
func (v ret) bytes() []c { return append(v.x().bytes(), 0x0f) }
func (v ret) cstr() s    { return jn("return ", cstring(v.x()), ";") }
func (v ret) gstr() s    { return jn("return ", gstring(v.x()), ";") }
func (v iff) rt() T      { return 0 }
func (v iff) valid() s {
	if t := v.x().rt(); t != I {
		return sf("conditional has wrong type %s", t)
	}
	if t := v.y().rt(); t != 0 {
		return "if statement must not return a value"
	}
	return ""
}
func (v iff) bytes() (r []c) { return catb(v.x().bytes(), []c{0x04, 0x40}, v.y().bytes(), []c{0x0b}) }
func (v iff) cstr() s        { return jn("if(", cstring(v.x()), ")", cstring(v.y())) }
func (v iff) gstr() s        { return jn("if ", gstring(v.x()), "{", gstring(v.y()), "}") }
func (v swc) rt() T          { return 0 }
func (v swc) valid() s {
	if v.x().rt() != I {
		return "swc wrong type for x"
	}
	for i, a := range v.argv[1:] {
		if a.rt() != 0 {
			return sf("swc arg %d must have zero type", i+1)
		}
	}
	return ""
}
func (v swc) bytes() (r []c) { // (block(block(... x br_table 0..n) argv[i];break)...)
	//fmt.Fprintf(os.Stderr, "%#v\n", v)
	n := len(v.argv) - 1
	r = catb(bytes.Repeat([]c{0x02, 0x40}, n+1), v.x().bytes(), []c{0x0e}, leb(n-1))
	for i := 0; i < n; i++ {
		r = append(r, leb(i)...)
	}
	r = append(r, 0x0b)
	for i, a := range v.argv[1:] {
		if _, o := a.(nop); o {
			r = catb(r, []c{0x0c}, leb(n-2-i), []c{0x0b})
		} else {
			r = catb(r, a.bytes(), []c{0x0c}, leb(n-1-i), []c{0x0b})
		}
	}
	return r
}
func (v swc) cstr() (r s) {
	r = jn("switch(", cstring(v.x()), "){")
	for i, a := range v.argv[1 : len(v.argv)-1] {
		if _, o := a.(nop); !o {
			r += sf("case %d:%sbreak;", i, cstring(a))
		}
	}
	return r + sf("default:%s}", cstring(v.argv[len(v.argv)-1]))
}
func (v swc) gstr() (r s) {
	r = jn("switch ", gstring(v.x()), "{")
	for i, a := range v.argv[1 : len(v.argv)-1] {
		if _, o := a.(nop); !o {
			r += sf("case %d:%s;", i, gstring(a))
		}
	}
	return r + sf("default:%s;}", gstring(v.argv[len(v.argv)-1]))
}
func (v nlp) rt() T { return 0 }
func (v nlp) valid() s {
	if xt, yt := v.x().rt(), v.y().rt(); xt != I {
		return sf("loop range is not I: %s", xt)
	} else if yt != 0 {
		return sf("loop body has nonzero type %s", yt)
	}
	return ""
}
func (v nlp) bytes() (r []c) {
	r = v.x().bytes()
	if isexpr(v.x()) {
		r = append(append(r, 0x22), leb(v.n)...) // tee.n for general expressions
	}
	i, n := s(leb(v.c)), s(leb(v.n))
	//                    if           0   →i   loop
	r = catb(r, []c(sf("\x04\x40\x41\x00\x21%s\x03\x40", i)))
	//                                        i       1   +  tee→i    n   <  continue
	return catb(r, v.y().bytes(), []c(sf("\x20%s\x41\x01\x6a\x22%s\x20%s\x49\x0d\x00\x0b\x0b", i, i, n)))
}
func (v nlp) cstr() (r s) {
	if isexpr(v.x()) {
		r = sf("x%d=%s;", v.n, cstring(v.x()))
	}
	return r + sf("for(x%d=0;x%d<x%d;x%d++){%s}", v.c, v.c, v.n, v.c, cstring(v.y()))
}
func (v nlp) gstr() (r s) {
	if isexpr(v.x()) {
		r = sf("x%d=%s;", v.n, gstring(v.x()))
	}
	return r + sf("for x%d=0;x%d<x%d;x%d++{%s}", v.c, v.c, v.n, v.c, gstring(v.y()))
}
func (v whl) rt() T { return 0 }
func (v whl) valid() s {
	if t := v.x().rt(); t != I {
		return sf("while conditional has wrong type %d", t)
	} else if v.y().rt() != 0 {
		return sf("while body must have no type")
	}
	return ""
}
func (v whl) bytes() (r []c) {
	cnd := sf("%s\x45\x0d\x01", s(v.x().bytes()))
	if _, o := v.x().(con); o {
		cnd = "" // 1/
	}
	//             block   loop     ? y  continue
	return []c(sf("\x02\x40\x03\x40%s%s\x0c\x00\x0b\x0b", cnd, s(v.y().bytes())))
}
func (v whl) cstr() s { return jn("while(", cstring(v.x()), "){", cstring(v.y()), "}") }
func (v whl) gstr() s {
	x := gstring(v.x())
	if x == "1" {
		x = ""
	}
	return jn("for ", x, "{", gstring(v.y()), "}")
}
func (v brif) rt() T      { return 0 }
func (v brif) valid() s   { return ifex(v.x().rt() != I, "brif has wrong conditional type") }
func (v brif) bytes() []c { return append(v.x().bytes(), 0x0d, 0x01) } // break outer block
func (v brif) cstr() s    { return jn("if(", cstring(v.x()), ")break;") }
func (v brif) gstr() s    { return jn("if ", gstring(v.x()), "{break}") }
func (v opx) rt() T       { return 0 }
func (v opx) valid() s    { return ifex(s(v) != "!", "nonapplied operator") }
func (v opx) bytes() []c  { return []c{0x00} }
func (v opx) cstr() s     { return "trap();" }
func (v opx) gstr() s     { return `panic("trap")` }
func (v pos) rt() T       { return 0 }
func (v pos) valid() s    { return "position(dummy expr)" }
func (v pos) bytes() []c  { return nil }
func (v nop) rt() T       { return 0 }
func (v nop) valid() s    { return "" }
func (v nop) bytes() []c  { return nil }
func (v nop) cstr() s     { return "()" }
func (v nop) gstr() s     { return "()" }

func locstr(v expr) s { return sf("x%d", v.(loc).i) }
func isexpr(x expr) bool { // general expr that needs an explicit assignment
	switch x.(type) {
	case las:
	case loc:
	default:
		return true
	}
	return false
}
func ifex(c bool, s s) s {
	if c {
		return s
	}
	return ""
}
func gmemshift(t T) s {
	shift := ""
	if t == I {
		shift = ">>2"
	} else if t == J || t == F {
		shift = ">>3"
	}
	return shift
}

type code struct {
	I, J, F c
	c, g    s
}

func c2str(tab map[s]code, op s, t T, x, y expr) s {
	o, u := cop(tab, op, t)
	if len(o) > 2 {
		return jn(u, o, "(", cstring(x), ",", cstring(y), ")")
	} else {
		return jn("((", u, cstring(x), ")", o, "(", u, cstring(y), "))")
	}
}
func g2str(tab map[s]code, op s, t T, x, y expr) s {
	o, u := cop(tab, op, t)
	u += "("
	if len(o) > 2 {
		return jn(u, o, "(", cstring(x), ",", cstring(y), "))")
	} else {
		return jn("((", u, cstring(x), "))", o, "(", u, cstring(y), ")))")
	}
}
func cstring(x expr) s { xs := x.(cstringer); return xs.cstr() }
func gstring(x expr) s { xs := x.(gstringer); return xs.gstr() }

var v1Tab = map[s]code{
	"-": code{0, 0, 0x9a, "-", "-"},                                                                            // neg (-I -J is replaced)
	"+": code{0, 0, 0x99, "fabs", "math.Abs"},                                                                  // abs (+I +J is not allowed)
	"~": code{0x45, 0x50, 0, "!", "!"},                                                                         // eqz
	"_": code{1, 1, 0x9c, ";;floor", "math.Floor"},                                                             // floor (ceil, trunc, nearest?)
	"*": code{0x67, 0x79, 0, "__builtin_clz;__builtin_clzll;", "Ubits.LeadingZeros32;Ubits.LeadingZeros64;"},   // clz
	"|": code{0x68, 0x79, 0, "__builtin_ctz;__builtin_ctzll;", "Ubits.TrailingZeros32;Ubits;TrailingZeros64;"}, // ctz
	"%": code{0, 0, 0x9f, "sqrt", "math.Sqrt"},                                                                 // sqr
}
var v2Tab = map[s]code{
	`+`:   code{0x6a, 0x7c, 0xa0, "+", "+"},     // add
	`-`:   code{0x6b, 0x7d, 0xa1, "-", "-"},     // sub
	`*`:   code{0x6c, 0x7e, 0xa2, "*", "*"},     // mul
	`%`:   code{0x6e, 0x80, 0xa3, "U/", "U/"},   // div/div_u
	`%'`:  code{0x6d, 0x7f, 0xa3, "/", "/"},     // div_s
	`\`:   code{0x70, 0x82, 0, "U%", "%U"},      // rem_u
	`\'`:  code{0x6f, 0x81, 0, "%", "%"},        // rem_s
	`&`:   code{0x71, 0x83, 0, "&", "&"},        // and
	`|`:   code{0x72, 0x84, 0, "|", "|"},        // or
	`^`:   code{0x73, 0x85, 0, "^", "^"},        // xor
	`<<`:  code{0x74, 0x86, 0, "<<", "<<"},      // shl
	`>>`:  code{0x76, 0x88, 0, "U>>", "U>>"},    // shr_u
	`>>'`: code{0x75, 0x87, 0, ">>", ">>"},      // shl_s
	`<|'`: code{0x77, 0x89, 0, "", ""},          // rotl
	`>|'`: code{0x78, 0x8a, 0, "", ""},          // rotr
	`&'`:  code{0, 0, 0xa4, "fmin", "math.Max"}, // min
	`|'`:  code{0, 0, 0xa5, "fmax", "math.Min"}, // max
}
var cTab = map[s]code{
	"<":   code{0x49, 0x54, 0x63, "U<", "U<"},   // lt/lt_u
	"<'":  code{0x48, 0x53, 0x63, "<", "<"},     // lt_s
	">":   code{0x4b, 0x56, 0x64, "U>", "U>"},   // gt/gt_u
	">'":  code{0x4a, 0x55, 0x64, ">", ">"},     // gt_s
	"<=":  code{0x4d, 0x58, 0x65, "U<=", "U<="}, // le/le_u
	"<='": code{0x4c, 0x57, 0x65, "<=", "<="},   // le_s
	">=":  code{0x4f, 0x5a, 0x66, "U>=", "U>="}, // ge/ge_u
	">='": code{0x4e, 0x59, 0x66, ">=", ">="},   // ge/ge_s
	"~":   code{0x46, 0x51, 0x61, "==", "=="},   // eq
	"!":   code{0x47, 0x52, 0x62, "!=", "!="},   // ne
}
var allops map[s]bool

func init() {
	allops = make(map[s]bool)
	for _, t := range []map[s]code{v1Tab, v2Tab, cTab} {
		for s := range t {
			allops[s] = true
		}
	}
}

// emit wasm byte code
func (m module) wasm(data []c) []c {
	o := bytes.NewBuffer([]c{0, 0x61, 0x73, 0x6d, 1, 0, 0, 0}) // header
	// type section(1: function signatures)
	sec := NewSection(1)
	sigs, sigv := make(map[s]int), make([]s, 0)
	for i, f := range m {
		s := s(f.sig())
		if n, o := sigs[s]; o == false {
			n = len(sigs)
			sigs[s] = n
			sigv = append(sigv, s)
			m[i].sign = n
		}
	}
	sec.cat(leb(len(sigv)))
	for _, s := range sigv {
		sec.cat([]c(s))
	}
	sec.out(o)
	// no import section(2)
	// function section(3: function signature indexes)
	sec = NewSection(3)
	sec.cat(leb(len(m)))
	for _, f := range m {
		sec.cat(leb(sigs[s(f.sig())]))
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
	idx, exp := m.exports()
	sec.cat(leb(len(exp))) // number of exports
	for i, f := range exp {
		sec.cat(leb(len(f.name)))
		sec.cat([]c(f.name))
		sec.cat1(0) // function-export
		sec.cat(leb(idx[i]))
	}
	sec.out(o)
	// no start section(8)
	// no element section(9)
	// code section(10)
	sec = NewSection(10)
	sec.cat(leb(len(m))) // number of functions
	for _, f := range m {
		b := f.code()
		sec.cat(leb(len(b)))
		sec.cat(b)
	}
	sec.out(o)
	// no data section(11)
	return o.Bytes()
}
func (m module) exports() (idx []int, fns []fn) {
	for i, f := range m {
		if f.ex {
			idx = append(idx, i)
			fns = append(fns, f)
		}
	}
	return idx, fns
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
	w.Write(leb(len(s.b)))
	w.Write(s.b)
}

func (f fn) sig() (r []c) {
	r = append(r, 0x60)
	r = append(r, leb(f.args)...)
	for i := 0; i < f.args; i++ {
		r = append(r, c(f.locl[i]))
	}
	if f.t == 0 {
		return append(r, 0)
	}
	r = append(r, 1)
	r = append(r, c(f.t))
	return r
}
func (f fn) code() (r []c) {
	r = append(r, f.locs()...)
	r = append(r, f.ast.bytes()...)
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
	r = leb(len(u))
	for i, t := range u {
		r = append(r, leb(n[i])...)
		r = append(r, c(t))
	}
	return r
}
func leb(v int) []c { return lebu(v) }
func lebu(v int) []c { // encode unsigned leb128
	if v < 0 {
		panic("lebu")
	}
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
func lebs(v int) []c { // encode signed leb128
	var b []c
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
func catb(x ...[]c) (r []c) {
	for _, b := range x {
		r = append(r, b...)
	}
	return r
}
func jn(a ...s) s                { return strings.Join(a, "") }
func log(a ...interface{})       { fmt.Fprintln(os.Stderr, a...) }
func logf(f s, a ...interface{}) { fmt.Fprintf(os.Stderr, f, a...) }
func sf(f s, a ...interface{}) s { return fmt.Sprintf(f, a...) }
func page(wasm []c) []c {
	var b bytes.Buffer
	b.WriteString(head)
	b.WriteString(base64.StdEncoding.EncodeToString(wasm))
	b.WriteString(tail)
	return b.Bytes()
}

func (m module) cout(data []c) []c {
	var b bytes.Buffer
	b.WriteString(chead)
	for _, f := range m {
		st := styp[f.t]
		if f.t == 0 {
			st = "void "
		}
		sig := ""
		for i := 0; i < f.args; i++ {
			if i > 0 {
				sig += ","
			}
			sig += styp[f.locl[i]]
		}
		fmt.Fprintf(&b, "%s %s(%s);", st, f.name, sig)
	}
	fmt.Fprintf(&b, "\n")
	for _, f := range m {
		sig, loc := "", ""
		for i := 0; i < f.args; i++ {
			if i > 0 {
				sig += ","
			}
			sig += styp[f.locl[i]] + " " + "x" + s('0'+byte(i))
		}
		for i := f.args; i < len(f.locl); i++ {
			t := f.locl[i]
			if t == F {
				loc += sf("%s x%d=.0;", styp[t], i)
			} else {
				loc += sf("%s x%d=0;", styp[t], i)
			}
		}
		st := styp[f.t]
		if f.t == 0 {
			st = "void "
		}
		fmt.Fprintf(&b, "%s %s(%s){%s", st, f.name, sig, loc)
		if sq, o := f.ast.(seq); o {
			for i, e := range sq.argv {
				nl := ""
				if i == len(sq.argv)-1 {
					if f.t != 0 {
						b.WriteString("R ")
					}
					nl = ";}\n"
				}
				b.WriteString(cstring(e))
				b.WriteString(nl)
			}
		} else {
			if f.t != 0 {
				b.WriteString("R ")
			}
			b.WriteString(cstring(f.ast))
			b.WriteString(";}\n")
		}
	}
	return b.Bytes()
}
func (m module) gout(data []c) []c {
	var b bytes.Buffer
	b.WriteString(ghead)
	for _, f := range m {
		sig := ""
		for i := 0; i < f.args; i++ {
			if i > 0 {
				sig += ","
			}
			sig += "x" + s('0'+byte(i)) + " " + styp[f.locl[i]]
		}
		fmt.Fprintf(&b, "func %s(%s) %s {", f.name, sig, styp[f.t])
		if sq, o := f.ast.(seq); o {
			for i, e := range sq.argv {
				if i < len(sq.argv)-1 {
					b.WriteString("return ")
				}
				b.WriteString(gstring(e))
				b.WriteString("}")
			}
		} else {
			b.WriteString("return ")
			b.WriteString(gstring(f.ast))
			b.WriteString("}\n")
		}
	}
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

var chead = ``
var ghead = `
type I=int32;type J=int64;type F=float64
func asI(x *I,y I)I{*x=y;return y};func lsJ(x *J,y J)J{*x=y;return y};func asF(x *F,y F)F{*x=y;return y};
`
