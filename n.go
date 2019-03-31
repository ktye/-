package i

import (
	"math"
	"math/cmplx"
	"reflect"
	"sort"
)

// atomic numeric monads
func rneg(a f) f { return -a }
func zneg(a z) z { return -a }
func rflr(a f) f { return math.Floor(a) }
func zflr(a z) z { return complex(math.Floor(cmplx.Abs(a)), 0) }
func rinv(a f) f { return 1.0 / a }
func zinv(a z) z { return complex(1, 0) / a }
func rsqr(a f) f { return math.Sqrt(a) }
func zsqr(a z) z { return cmplx.Sqrt(a) }
func rabs(a f) f { return math.Abs(a) }
func zabs(a z) z { return complex(cmplx.Abs(a), 0) }
func rnot(a f) f { return rter(a == 0, 1, 0) }
func znot(a z) z { return zter(a == 0, 1, 0) }
func ris0(a f) f {
	if math.IsNaN(a) {
		return 1.0
	}
	return 0.0
}
func zis0(a z) z {
	if cmplx.IsNaN(a) {
		return 1.0
	}
	return 0.0
}
func rexp(a f) f { return math.Exp(a) }
func zexp(a z) z { return cmplx.Exp(a) }
func rlog(a f) f { return math.Log(a) }
func zlog(a z) z { return cmplx.Log(a) }

// atomic numeric dyads
func radd(a, b f) f { return a + b }
func zadd(a, b z) z { return a + b }
func rsub(a, b f) f { return a - b }
func zsub(a, b z) z { return a - b }
func rmul(a, b f) f { return a * b }
func zmul(a, b z) z { return a * b }
func rdiv(a, b f) f { return a / b }
func zdiv(a, b z) z { return a / b }
func rmod(a, b f) f { return math.Mod(b, a) }
func zmod(a, b z) z { return complex(math.Mod(re(b), re(a)), 0) }
func rmin(a, b f) f { return rter(a < b, a, b) }
func zmin(a, b z) z { return zter(cmplx.Abs(a) < cmplx.Abs(b), a, b) } // what about equal abs? compare angle?
func rmax(a, b f) f { return rter(a > b, a, b) }
func zmax(a, b z) z { return zter(cmplx.Abs(a) > cmplx.Abs(b), a, b) }
func rles(a, b f) f { return rter(a < b, 1, 0) }
func zles(a, b z) z { return zter(cmplx.Abs(a) < cmplx.Abs(b), 1, 0) }
func rmor(a, b f) f { return rter(a > b, 1, 0) }
func zmor(a, b z) z { return zter(cmplx.Abs(a) > cmplx.Abs(b), 1, 0) }
func reql(a, b f) f { return rter(a == b, 1, 0) } // tolerance?
func zeql(a, b z) z { return zter(a == b, 1, 0) }
func rpow(a, b f) f { return math.Pow(a, b) }
func zpow(a, b z) z { return cmplx.Pow(a, b) }

func rter(c bool, a, b f) f {
	if c {
		return a
	}
	return b
}
func zter(c bool, a, b z) z {
	if c {
		return a
	}
	return b
}

type fr1 func(f) f
type fr2 func(f, f) f
type fz1 func(z) z
type fz2 func(z, z) z

func nm(x v, fr fr1, fz fz1) v {
	if d, ok := md(x); ok {
		for i, v := range d.v {
			d.v[i] = nm(v, fr, fz)
		}
		return d.mp()
	}
	if iv, ok := x.(l); ok {
		r := make(l, len(iv))
		for i := range r {
			r[i] = nm(iv[i], fr, fz)
		}
		return r
	}

	y, z, vec, t := nv(x)
	if y != nil {
		for i, x := range y {
			y[i] = fr(x)
		}
		return vn(y, nil, vec, t)
	}
	for i, x := range z {
		z[i] = fz(x)
	}
	return vn(nil, z, vec, t)
}

func nd(x, y v, fr fr2, fz fz2) v {
	if r, ok := ndic(x, y, fr, fz); ok {
		return r
	}

	xn, yn := ln(x), ln(y)
	switch {
	case xn >= 0 && yn >= 0 && xn != yn:
		return e("length")
	case xn < 0 && yn >= 0:
		x, xn = rsh(yn, x), yn
	case yn < 0 && xn >= 0:
		y, yn = rsh(xn, y), xn
	}
	xl, yl := false, false
	if xn >= 0 && rtyp(x).Elem().Kind() == reflect.Interface {
		xl = true
	}
	if yn >= 0 && rtyp(y).Elem().Kind() == reflect.Interface {
		yl = true
	}
	if xl || yl {
		r := make(l, xn) // TODO: make custom interface type, if both have the same type
		for i := range r {
			r[i] = nd(at(x, i), at(y, i), fr, fz)
		}
		return r
	}

	xr, xz, xvec, xt := nv(x)
	yr, yz, yvec, yt := nv(y)
	if xz != nil || yz != nil {
		if xr != nil {
			xz = toZ(xr)
			xr = nil
		} else if yr != nil {
			yz = toZ(yr)
			yr = nil
		}
	}
	n1, n2 := len(xr), len(yr)
	if xr == nil {
		n1, n2 = len(xz), len(yz)
	}
	if n1 == 0 || n2 == 0 {
		if xt == yt && xt != nil {
			return ms(xt, 0).Interface()
		}
		return l{}
	}
	if n1 == 1 && n2 > 1 {
		xr, xz = nrsh(xr, xz, n2)
		n1 = n2
	} else if n1 > 1 && n2 == 1 {
		yr, yz = nrsh(yr, yz, n1)
		n1 = n2
	}
	if n1 != n2 {
		e("length")
	}
	if xr != nil {
		for i := range xr {
			xr[i] = fr(xr[i], yr[i])
		}
	} else {
		for i := range xz {
			xz[i] = fz(xz[i], yz[i])
		}
	}
	vec := false
	if xvec || yvec {
		vec = true
	}
	if xt == yt {
		return vn(xr, xz, vec, xt)
	}
	return vn(xr, xz, vec, nil)
}
func ndic(x, y v, fr fr2, fz fz2) (v, bool) {
	xd, isx := md(x)
	yd, isy := md(y)
	if isx == false && isy == false {
		return nil, false
	}
	if isx && isy {
		// That could be an identity element, depending on the verb.
		// oK just fills the other value without applying the function.
		zero := 0.0 // If there is no agreement, i can do what i want.
		for i, k := range xd.k {
			yi, v := yd.at(k)
			if yi < 0 {
				xd.v[i] = nd(xd.v[i], zero, fr, fz)
			} else {
				xd.v[i] = nd(xd.v[i], v, fr, fz)
			}
		}
		for i, k := range yd.k {
			if idx, _ := xd.at(k); idx < 0 {
				xd.k = append(xd.k, k)
				xd.v = append(xd.v, nd(zero, yd.v[i], fr, fz))
			}
		}
		if xd.t != yd.t {
			xd.t = nil
		}
		return xd.mp(), true
	}
	// d+v is not allowed, but d+a is.
	d := xd
	a := y
	flip := false
	if isx == false {
		d = yd
		a = x
		flip = true
	}
	if ln(a) >= 0 {
		e("type") // d+v is not allowed
	}
	for i, _ := range d.k {
		x, y = d.v[i], a
		if flip {
			x, y = y, x
		}
		d.v[i] = nd(x, y, fr, fz)
	}
	return xd.mp(), true
}

func nv(x v) (fv, zv, bool, rT) { // import any number or numeric vector types
	switch t := x.(type) {
	case f:
		return fv{t}, nil, false, rTf
	case z:
		return nil, zv{t}, false, rTz
	case fv:
		r := make(fv, len(t))
		copy(r, t)
		return r, nil, true, rTf
	case zv:
		r := make(zv, len(t))
		copy(r, t)
		return nil, t, true, rTz
	case s:
		e("type")
	}
	v := rval(x)
	if v.Kind() == reflect.Slice {
		n := v.Len()
		z := reflect.Zero(v.Type().Elem())
		if z.Type().ConvertibleTo(rTf) {
			r := make(fv, n)
			for i := range r {
				r[i] = v.Index(i).Convert(rTf).Float()
			}
			return r, nil, true, z.Type()
		} else if z.Type().ConvertibleTo(rTz) {
			r := make(zv, n)
			for i := range r {
				r[i] = v.Index(i).Convert(rTz).Complex()
			}
			return nil, r, true, z.Type()
		} else if z.Type().ConvertibleTo(rTb) {
			r := make(fv, n)
			for i := range r {
				b := v.Index(i).Convert(rTb).Bool()
				if b {
					r[i] = 1
				}
			}
			return r, nil, true, z.Type()
		}
		e("type")
	}

	if v.Type().ConvertibleTo(rTf) {
		return fv{v.Convert(rTf).Float()}, nil, false, v.Type()
	}
	if v.Type().ConvertibleTo(rTz) {
		return nil, zv{v.Convert(rTz).Complex()}, false, v.Type()
	}
	if v.Type().ConvertibleTo(rTb) {
		b := v.Convert(rTb).Bool()
		r := 0.0
		if b {
			r = 1.0
		}
		return fv{r}, nil, false, v.Type()
	}
	e("type")
	return nil, nil, false, rTf
}
func vn(x fv, z zv, vec bool, t rT) interface{} { // convert numbers back to original type
	if x != nil && (t == rTf || t == nil) {
		if vec {
			return x
		}
		return x[0]
	}
	if z != nil && (t == rTz || t == nil) {
		if vec {
			return z
		}
		return z[0]
	}
	if vec == false {
		if x != nil {
			if t.ConvertibleTo(rTb) {
				b := false
				if x[0] != 0 {
					b = true
				}
				return rval(b).Convert(t).Interface()
			}
			return rval(x[0]).Convert(t).Interface()
		}
		return rval(z[0]).Convert(t).Interface()
	}
	n := len(x)
	if x == nil {
		n = len(z)
	}
	r := ms(t, n)
	for i := 0; i < n; i++ {
		if x != nil {
			if t.ConvertibleTo(rTb) {
				b := false
				if x[i] != 0 {
					b = true
				}
				r.Index(i).Set(rval(b).Convert(t))
			} else {
				r.Index(i).Set(rval(x[i]).Convert(t))
			}
		} else {
			r.Index(i).Set(rval(z[i]).Convert(t))
		}
	}
	return r.Interface()
}
func sn(v v) (fv, bool, bool) { // import strings as numbers; for =<>
	s, n, _, o := sy(v)
	if o == false {
		return nil, false, false
	}
	if n < 0 {
		return fv{0}, false, true
	}
	m := strmap(s)
	r := make(fv, n)
	for i := range s {
		r[i] = m[s[i]]
	}
	return r, true, true
}
func sn2(x, y v) (v, v) { // map strings to floats
	sx, nx, _, o := sy(x)
	if o == false {
		return x, y
	}
	sy, ny, _, o := sy(y)
	if o == false {
		return x, y
	}
	vec := true
	if nx < 0 && ny < 0 {
		vec, nx, ny = false, 1, 1
	} else if nx < 0 {
		sx, nx = rsh(ny, sx).(sv), ny
	} else if ny < 0 {
		sy, ny = rsh(nx, sy).(sv), nx
	} else if nx != ny {
		e("length")
	}
	b := make(sv, nx+ny)
	copy(b, sx)
	copy(b[nx:], sy)
	m := strmap(b)
	rx := make(fv, nx)
	for i := range sx {
		rx[i] = m[sx[i]]
	}
	ry := make(fv, ny)
	for i := range sy {
		ry[i] = m[sy[i]]
	}
	if !vec {
		return rx[0], ry[0]
	}
	return rx, ry
}
func strmap(x sv) map[s]f { // map s to f uniq and comparable
	n := len(x)
	idx := til(f(n)).(fv)
	c := cp(x).(sv)
	u := grades{sort.StringSlice(c), idx}
	sort.Sort(u)
	m := make(map[s]f)
	w := 0.0
	for i := range u.idx {
		if i == 0 || c[i] != c[i-1] {
			m[c[i]] = w
			w += 1.0
		}
	}
	return m
}

func toZ(x fv) zv {
	z := make(zv, len(x))
	for i, r := range x {
		z[i] = complex(r, 0)
	}
	return z
}
func nrsh(x fv, z zv, n int) (fv, zv) {
	if x == nil {
		r := make(zv, n)
		for i := range r {
			r[i] = z[0]
		}
		return nil, r
	}
	r := make(fv, n)
	for i := range r {
		r[i] = x[0]
	}
	return r, nil
}
func nl(x v) l {
	v := rval(x)
	if v.Kind() == reflect.Slice {
		r := make(l, v.Len())
		for i := range r {
			rval(x).Index(i).Set(v.Index(i))
		}
		return r
	} else {
		return l{x}
	}
}
