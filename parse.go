package k

import (
	. "github.com/ktye/wg/module"
)

func parse(x K) (r K) {
	x = tok(x)
	pp = int32(x)
	pe = pp + 8*nn(x)

	r = e(t())
	return r
}
func e(x K, xv int32) (r K) { // Lt
	if x == 0 {
		return 0
	}
	y, yv := t()
	if y == 0 {
		return x
	}
	if yv != 0 && xv == 0 {
		r = e(t())
		r = ucat(r, x)
		r = ucat(r, y)
		return dyadic(r) // dyadic
	}
	r = e(y, yv)
	if xv == 0 || (r == y && xv+yv == 2) {
		return cat1(cat1(ucat(r, x), 19), 1) // juxtaposition or train
	}
	r = ucat(r, x)
	return monadic(r) // monadic
}
func t() (r K, verb int32) { // Lt
	var ln int32
	r = next()
	if r == 0 {
		return 0, 0
	}
	if r < 127 && is(int32(r), 32) {
		pp -= 8
		return 0, 0
	}
	if r == K('(') {
		r, ln = plist(41)
		if ln == 1 {
			r = Fst(r)
		} else {
			r = cat3(flat(Rev(r)), Ki(ln), 27, 0)
		}
	} else {
		r, verb = l1(r), ib(tp(r) == 0)
	}
	for {
		n := next()
		if n == 0 {
			break
		}
		a := int32(n)
		if tp(n) == 0 && a > 20 && a < 27 {
			r, verb = cat1(cat1(r, n), 0), 1
		} else if n == 91 { // [
			n, ln = plist(93)
			verb = 0
			if ln == 1 {
				r = cat1(cat1(Cat(Fst(n), r), 19), 1)
			} else {
				n = cat3(flat(Rev(n)), Ki(ln), 27, 0)
				r = cat1(cat1(Cat(n, r), 20), 1)
			}
		} else {
			pp -= 8
			break
		}
	}
	return r, verb
}

func plist(c K) (r K, n int32) {
	r = mk(Lt, 0)
	b := next()
	if b == 0 || b == c {
		return r, 0
	}
	pp -= 8
	for {
		n++
		x := e(t())
		r = cat1(r, x)
		b = next()
		if b == c {
			break
		}
		if b != 59 { // ;
			trap(Parse)
		}
	}
	return r, n
}

func next() (r K) {
	if pp == pe {
		return 0
	}
	r = K(I64(pp))
	pp += 8
	return r
}
func lastp(x K) K { return K(I64(int32(x) + 8*(nn(x)-1))) }
func dyadic(x K) K {
	l := lastp(x)
	if l < 2 {
		x = cat1(x, 20) // .
	}
	return cat1(x, 1)
}
func monadic(x K) K {
	l := lastp(x)
	if l < 2 {
		return cat1(cat1(x, 19), 1) // @
	}
	return cat1(x, 0)
}