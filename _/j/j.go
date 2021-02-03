// ..000 pointer(list)
// ....1 int  x>>1
// ...10 symbol x>>2
// ..100 operator x>>3
//
// 0     total memory (log2)
// 1     symbol list
// 2     value list
// 3     parse list
// 4..32 free list
//
// abc   symbol (max 6)
// 123   int (max 31 bit)
// [..]  list/quote
// #     length/non-list: -1
// +-*%\ arith(mod)
// <=>   compare
// &^    min max
//
// nyi:
// .exec 'each /over
// [a][b]: assign
// [c][t]? if
// a i@ index
// a i v$store
// ;putc
// !trace
// `trap
// n} break
// addpc depth{
// (comment)
//  go:embed j.j
//  var j []byte
package j

import (
	_ "embed"
	"math/bits"
)

var N uint32                // number
var S uint32                // symbol
var M []uint32              // heap
var P uint32                // current parse list
var F []func(uint32) uint32 // function table

func Step(x uint32) uint32 {
	if x >= '0' && x <= '9' {
		N *= 10
		N += x - '0'
		return 0
	}
	if N != 0 {
		P = cat(P, 1|N<<1)
		N = 0
	}
	if x >= 'a' && x <= 'z' {
		S *= 32
		S += x - 'a'
		return 0
	}
	if S != 0 {
		P = cat(P, 2|S<<2)
		S = 0
	}
	if x < 33 {
		if x == 10 {
			return Exec(M[3])
		}
		return 0
	}
	if x == 91 { // '['
		P = cat(P, mk(0))
		P = last(P)
		return 0
	}
	if x == 93 {
		P = parent(M[3], P)
		if P == 0 {
			panic("parse]")
		}
		return 0
	}
	P = cat(P, 4|(x-33)<<3)
	return 0
}
func Exec(x uint32) uint32 {
	stk := mk(0)
	n := nn(x)
	xp := P + 8
	for i := uint32(0); i < n; i++ {
		x := I(xp)
		if x&7 != 4 {
			stk = lcat(stk, rx(x))
		} else {
			stk = F[x>>3](stk)
		}
		xp += 4
	}
	return stk
}
func init() {
	finit()
	x := uint32(16)
	M = make([]uint32, 1<<(x-2)) // 64kB
	M[0] = uint32(x)
	p := uint32(128)
	for i := uint32(7); i < x; i++ {
		sI(4*i, p) // free pointer
		p *= 2
	}

	M[1] = mk(0)
	M[2] = mk(0)
	M[3] = mk(0)
	P = M[3]
	//dump(127)
}
func bk(n uint32) (r uint32) { // bucket type
	r = uint32(32 - bits.LeadingZeros32(7+4*n))
	if r < 4 {
		return 4
	}
	return r
}
func mk(x uint32) (r uint32) { // allocate
	t := bk(x)
	i := 4 * t
	m := 4 * M[0]
	for I(i) == 0 {
		if i >= m {
			panic("memory")
		}
		i += 4
	}
	a := I(i)
	sI(i, I(a))
	for j := i - 4; j >= 4*t; j -= 4 {
		u := a + 1<<(j>>2)
		sI(u, I(j))
		sI(j, u)
	}
	sI(a, 1)
	sI(a+4, x)
	return a
}
func rx(x uint32) uint32 {
	if x&7 == 0 {
		sI(x, I(x)+1)
	}
	return x
}
func dx(x uint32) uint32 {
	if x&7 == 0 {
		sI(x, I(x)-1)
		if I(x) == 0 {
			n := I(x + 4)
			p := x + 8
			for i := uint32(0); i < n; i++ {
				dx(I(p))
				p += 4
			}
			fr(x)
		}
		return x
	}
	return x
}
func fr(x uint32) {
	p := 4 * bk(I(4+x))
	sI(x, I(p))
	sI(p, x)
}
func nn(x uint32) uint32 { return I(4 + x) }
func lcat(x uint32, y uint32) (r uint32) {
	n := nn(x)
	r = mk(1 + n)
	xp, rp := x+8, r+8
	for i := uint32(0); i < n; i++ {
		sI(rp, rx(I(xp)))
		rp += 4
		xp += 4
	}
	sI(rp, y)
	dx(x)
	return r
}
func cat(x, y uint32) (r uint32) {
	p := parent(M[3], x)
	r = lcat(x, y)
	if x == M[3] {
		M[3] = r
		return r
	}
	sI(lastp(p), r)
	return r
}
func lastp(x uint32) uint32 {
	n := nn(x)
	if n == 0 {
		panic("empty")
	}
	return 4 + x + 4*n
}
func last(x uint32) (r uint32) {
	n := nn(x)
	if n == 0 {
		return 0
	}
	return lastp(x)
}
func last2(x uint32) (a, b uint32) {
	n := nn(x)
	if n < 2 {
		panic("stack-underflow")
	}
	x += 4 * n
	return I(x), I(x + 4)
}
func parent(x, y uint32) (r uint32) {
	if x&7 != 0 {
		panic("parent")
	}
	l := last(x)
	if l == y || l == 0 || x == y {
		return x
	}
	return parent(l, y)
}
func I(x uint32) uint32 { return M[x>>2] }
func sI(x, y uint32)    { M[x>>2] = y }

func finit() {
	f := func(c byte, g func(uint32) uint32) { F[c-33] = g }
	F = make([]func(uint32) uint32, 128)
	f('~', swp)
	f('"', dup)
	f('_', pop)
	f('|', rol)
	f('#', cnt)
	f('+', add)
	f('-', sub)
	f('*', mul)
	f('%', dif)
	f('\\', mod)
	f('=', eql)
	f('>', gti)
	f('<', lti)
	f('&', min)
	f('^', max)
}
func swp(s uint32) uint32 {
	x := lastp(s)
	if x < s+12 {
		panic("swp underflow")
	}
	t := I(x)
	sI(x, I(x-4))
	sI(x-4, t)
	return s
}
func dup(s uint32) uint32 { p := last(s); return lcat(s, rx(p)) }
func rol(s uint32) uint32 {
	p := lastp(s)
	if p < s+16 {
		panic("rol underflow")
	}
	a := I(p)
	sI(p, I(p-4))
	sI(p-4, I(p-8))
	sI(p-8, a)
	return s
}
func cnt(s uint32) uint32 {
	x := last(s)
	r := uint32(0xffffffff)
	if x&7 == 0 {
		r = 1 + 2*nn(x)
	}
	return v1(s, r)
}
func v1(s, x uint32) uint32 {
	sp := s + 4 + 4*nn(s)
	dx(I(sp))
	sI(sp, x)
	return s
}
func ints(s uint32) (j, k int32) {
	a, b := last2(s)
	if a&1 == 0 || b&1 == 0 {
		panic("ints")
	}
	return int32(a) >> 1, int32(b) >> 1
}
func add(s uint32) uint32 { a, b := ints(s); return i2(s, a+b) }
func sub(s uint32) uint32 { a, b := ints(s); return i2(s, a-b) }
func mul(s uint32) uint32 { a, b := ints(s); return i2(s, a*b) }
func dif(s uint32) uint32 { a, b := ints(s); return i2(s, a/b) }
func mod(s uint32) uint32 { a, b := ints(s); return i2(s, a%b) }
func eql(s uint32) uint32 { a, b := last2(s); return i2(s, ib(a == b)) }
func gti(s uint32) uint32 { a, b := ints(s); return i2(s, ib(a > b)) }
func lti(s uint32) uint32 { a, b := ints(s); return i2(s, ib(a < b)) }
func max(s uint32) uint32 {
	a, b := ints(s)
	if a > b {
		return i2(s, a)
	}
	return i2(s, b)
}
func min(s uint32) uint32 {
	a, b := ints(s)
	if a < b {
		return i2(s, a)
	}
	return i2(s, b)
}
func i2(s uint32, a int32) uint32 {
	s = pop(s)
	n := nn(s)
	sI(s+4+4*n, uint32(1|(a<<1)))
	return s
}
func ib(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func pop(x uint32) (r uint32) {
	n := nn(x)
	if n == 0 {
		panic("pop:underflow")
	}
	if bk(n) == bk(n-1) {
		sI(x+4, n-1)
	} else {
		n--
		r = mk(n)
		rp := r + 8
		xp := x + 8
		for i := uint32(0); i < n; i++ {
			sI(rp, rx(I(xp)))
			xp += 4
			rp += 4
		}
		dx(x)
		return r
	}
	return x
}
