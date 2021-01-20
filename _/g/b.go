package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/cmplx"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"text/tabwriter"
)

var nyi = errors.New("nyi")
var plotKeys uint32
var symbols = make(map[string]uint32)
var xfile string
var xline int

func ginit() {
	MT['r'+128] = read1
	MT['p'+128] = plot1
	MT['c'+128] = caption1
	MT['w'] = where2
	assign("read", 'r')
	assign("caption", 'c')
	assign("plot", 'p')
	assign("where", 'w')
	for _, s := range []string{"WIDTH", "HEIGHT", "COLUMNS", "LINES", "FFMT", "ZFMT"} {
		symbols[s] = ks(s)
	}
	assign("WIDTH", mki(800))
	assign("HEIGHT", mki(400))
	assign("COLUMNS", mki(80))
	assign("LINES", mki(20))
	assign("FFMT", kC([]byte("%.4g")))
	assign("ZFMT", kC([]byte("%.4ga%.0f")))
	plotKeys = kS([]string{"Type", "Style", "Limits", "Xlabel", "Ylabel", "Title", "Xunit", "Yunit", "Zunit", "Lines", "Foto", "Caption", "Data"})

}
func init() {
	exit = exitRepl
	Out = gOut

	// -s lines, cols (terminal size)
	s := func(a string, tail []string) ([]string, bool) {
		if a != "-s" || len(tail) < 2 {
			return tail, false
		}
		lines, cols := atoi(tail[0]), atoi(tail[1])
		assign("LINES", ki(lines))
		assign("COLUMNS", ki(cols))
		assign("WIDTH", ki(cols*11))
		assign("HEIGHT", ki(lines*20))
		return tail[2:], true
	}
	argvParsers = append(argvParsers, s)

	// \h(help)
	h := func(a string) bool {
		if a != `\h` && a != `\` {
			return false
		}
		fmt.Println(help)
		return true
	}
	le := func(a string) bool {
		if strings.HasPrefix(a, `\leak`) == false {
			return false
		}
		bleak()
		fmt.Println("no leak")
		return true
	}
	// \c(caption)
	c := func(a string) bool {
		if a != `\c` {
			return false
		}
		if lastCaption != nil {
			w, _ := clipTerminal()
			lastCaption.WriteTable(w, 0)
		}
		return true
	}
	replParsers = append(replParsers, h, le, c)
	kiniRunners = append(kiniRunners, ginit)
}

func exitRepl(x int) {
	if interactive {
		os.Exit(x)
	}
}
func bleak() {
	b := make([]uint64, len(MJ))
	copy(b, MJ)
	dx(plotKeys)
	for _, x := range symbols {
		dx(x)
	}
	leak()
	copy(MJ, b)
	msl()
}
func memstore() []uint32 {
	m := make([]uint32, len(MI))
	copy(m, MI)
	return m
}
func memcompare(m []uint32, s string) {
	if len(m) != len(MI) {
		panic(fmt.Sprintf("%s modified memory size: before %d now %d\n", s, len(m), len(MI)))
	}
	for i, u := range m {
		if u != MI[i] {
			panic(fmt.Sprintf("%s modified memory at %x(%d): 0x%x != 0x%x", s, i, i, m[i], MI[i]))
		}
	}
}
func gOut(x uint32) {
	o, lines := clipTerminal()

	//m := memstore()
	p := pk(x)
	//memcompare(m, "pk")
	if p != nil {
		showPlot(p)
		//memcompare(m, "showplot")
		return
	}
	if rows, cols := istab(x); rows > 1 {
		writeTable(x, o, rows, cols, lines)
		return
	} else if tp(x) == 7 {
		writeDict(x, o, lines)
		return
	}
	rx(x)
	o.Write(append(CK(kst(x)), 10))
}
func Loadfile(file string) error {
	b := make([]uint64, len(MJ))
	copy(b, MJ)
	defer func() {
		if r := recover(); r != nil {
			ics(I(140), I(144), os.Stdout)
			MJ = b
			msl()
		}
	}()
	if strings.HasSuffix(file, ".k") {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		xfile = file
		_, err = runscript(f)
		return err
	}
	return fmt.Errorf("loadfile: unknown file type: %s\n", file)
}
func read1(x uint32) uint32 {
	if tp(x) == 5 && nn(x) == 1 {
		x = cs(x)
	}
	if tp(x) != 1 {
		panic("type")
	}
	if nn(x) == 0 {
		dx(x)
		x = kC([]byte("./"))
	}
	if MC[7+x+nn(x)] == '/' {
		return readdir(string(CK(x)))
	}
	b, e := ioutil.ReadFile(string(CK(x)))
	fatal(e)
	return kC(b)
}
func readdir(s string) uint32 {
	fi, e := ioutil.ReadDir(s)
	fatal(e)
	keys, vals := make([]string, len(fi)), make([]int, len(fi))
	for i, f := range fi {
		keys[i] = f.Name()
		if f.IsDir() {
			keys[i] += "/"
		}
		vals[i] = int(f.Size())
	}
	return mkd(kS(keys), kI(vals))
}
func LoadDataFile(file, sym string) error {
	b, e := ioutil.ReadFile(file)
	if e != nil {
		return e
	}
	x := mk(1, uint32(len(b)))
	copy(MC[8+x:], b)
	assign(sym, x)
	return nil
}
func runscript(r io.Reader) (uint32, error) {
	xline = 0
	scn := bufio.NewScanner(includes(r))
	var x uint32
	for scn.Scan() {
		xline++
		s := strings.TrimSpace(scn.Text())
		if s == "" || strings.HasPrefix(s, "/") {
			continue
		}
		if idx := strings.Index(s, " /"); idx != -1 {
			s = s[:idx]
		}
		dx(x)
		x = val(kC([]byte(" " + s)))
	}
	xline = 0
	return x, nil
}

// clip to COLUMNS/LINES if interactive
func clipTerminal() (io.Writer, int) {
	if !interactive {
		return os.Stdout, 0
	}
	c := lupInt("COLUMNS")
	l := lupInt("LINES")
	if c <= 0 || l <= 0 {
		return os.Stdout, 0
	}
	return &clipWriter{Writer: os.Stdout, c: c - 2, l: l - 2}, l
}

func atoi(s string) int {
	i, e := strconv.Atoi(s)
	fatal(e)
	return i
}

func printStack(stack []byte) {
	if !interactive {
		debug.PrintStack()
		return
	}
	v := bytes.Split(stack, []byte{10})
	var o []string
	for _, b := range v {
		s := string(b)
		if strings.HasPrefix(s, "\t") {
			s = strings.TrimPrefix(s, "\t")
			if i := strings.Index(s, " +"); i > 0 {
				s = s[:i]
			}
			w := strings.Split(s, "/")
			if len(w) > 2 {
				w = w[len(w)-2:]
			}
			s = strings.Join(w, "/")
			if strings.HasPrefix(s, "debug") || strings.HasPrefix(s, "runtime") {
				continue
			}
			o = append(o, " "+s)
		}
	}
	if len(o) > 10 {
		o = o[:10]
	}
	if len(o) > 1 {
		for i := len(o) - 1; i >= 0; i-- {
			fmt.Println(o[i])
		}
	}
}

type clipWriter struct {
	io.Writer
	c, l int
	x, y int
}

func (cw *clipWriter) Write(p []byte) (n int, err error) {
	size := len(p)
	for {
		if len(p) == 0 {
			break
		} else if cw.l > 0 && cw.y >= cw.l {
			if cw.y == cw.l {
				cw.Writer.Write([]byte("..\n"))
			}
			cw.y++
			break
		}
		idx := bytes.IndexByte(p, '\n')
		if idx == -1 {
			if xx := cw.x + len(p); xx > cw.c {
				p = p[:cw.c-cw.x]
			}
			cw.Writer.Write(p)
			cw.x += len(p)
			break
		} else {
			if xx := cw.x + idx - 1; xx > cw.c {
				cw.Writer.Write(p[:cw.c-cw.x])
				cw.Writer.Write([]byte("..\n"))
			} else {
				cw.Writer.Write(p[:idx+1])
			}
			p = p[idx+1:]
			cw.x = 0
			cw.y++
		}
	}
	return size, nil
}

func lupInt(s string) int { // no modification
	r := lookup(s)
	if tp(r) != 2 || nn(r) != 1 {
		panic("var " + s + " is not int#1")
	}
	return int(MI[2+r>>2])
}
func lupString(s string) string { // no modification
	r := lookup(s)
	if tp(r) != 1 {
		panic("var " + s + " is not a char")
	}
	return string(Ck(r))
}
func lookup(s string) uint32 { // no modification
	x, o := symbols[s]
	if o == false {
		panic("var " + s + " is not a registered symbol")
	}
	return I(I(kval) + I(x+8))
}
func assign(s string, v uint32) { dx(asn(ks(s), v)) }
func kerr(e error) bool {
	if e == nil {
		return false
	}
	fmt.Fprintln(os.Stderr, e)
	if interactive == false {
		os.Exit(1)
	}
	return true
}
func perr(e error) {
	if e != nil {
		panic(e)
	}
}

func istab(x uint32) (rows, cols int) {
	if tp(x) == 7 {
		v := MI[3+x>>2]
		n := nn(v)
		if n == 0 {
			return 0, 0
		}
		n0 := nn(MI[2+v>>2])
		for i := uint32(0); i < n; i++ {
			if nn(MI[2+i+v>>2]) != n0 {
				return 0, -1
			}
		}
		return int(n0), int(nn(v))
	}
	return 0, -1
}

func writeDict(x uint32, w io.Writer, clip int) {
	k := Sk(I(8 + x))
	m := 1
	for i := range k {
		if n := len(k[i]); n > m {
			m = n
		}
	}
	rx(x)
	x = val(x)
	x = ech(x, 'k'+128)
	for i, s := range k {
		fmt.Fprintf(w, "%s%s|%s\n", s, strings.Repeat(" ", m-len(s)), string(Ck(I(8+x+uint32(4*i)))))
		if clip > 0 && i > clip {
			fmt.Fprintf(w, "..\n")
			break
		}
	}
	dx(x)
}
func writeTable(x uint32, ww io.Writer, rows, cols int, clip int) {
	tab := []byte{'\t'}
	nl := []byte{'\n'}
	ffmt := lupString("FFMT")
	zfmt := lupString("ZFMT")
	w := tabwriter.NewWriter(ww, 2, 8, 1, ' ', 0)
	keys, vals := Sk(MI[2+x>>2]), MI[3+x>>2]
	for i := range keys {
		w.Write([]byte(keys[i]))
		if i != int(cols-1) {
			w.Write(tab)
		}
	}
	w.Write(nl)
	for k := 0; k < rows; k++ {
		for i := 0; i < cols; i++ {
			w.Write([]byte(fmtVecAt(MI[2+uint32(i)+vals>>2], uint32(k), ffmt, zfmt)))
			if i != cols-1 {
				w.Write(tab)
			}
		}
		w.Write(nl)
		if clip > 0 && int(k) > clip {
			w.Write([]byte("..\n"))
			break
		}
	}
	w.Flush()
}

func fmtVecAt(x uint32, i uint32, ffmt, zfmt string) string {
	switch tp(x) {
	case 2:
		return strconv.Itoa(int(MI[2+i+x>>2]))
	case 3:
		return fmt.Sprintf(ffmt, MF[1+i+x>>3])
	case 4:
		z := complex(MF[1+2*i+x>>3], MF[2+2*i+x>>3])
		return absang(z, zfmt)
	case 5:
		return ski(MI[2+i+x>>2])
	default:
		return "?"
	}
}

// resolve backslash includes: \file.k
func includes(r io.Reader) io.Reader {
	b, e := ioutil.ReadAll(r)
	if e != nil {
		panic(e)
	}
	b = bytes.Replace(b, []byte("\r"), []byte{}, -1)
	v := bytes.Split(b, []byte("\n"))

	var buf bytes.Buffer
	for _, u := range v {
		if len(u) > 3 && u[0] == '\\' && bytes.HasSuffix(u, []byte(".k")) {
			if f, e := ioutil.ReadFile(string(u[1:])); e != nil {
				panic(e)
			} else {
				f = bytes.Replace(f, []byte("\r"), []byte{}, -1)
				buf.Write(f)
			}
		} else {
			buf.Write(u)
			buf.Write([]byte{'\n'})
		}
	}
	return &buf
}
func absang(x complex128, format string) string {
	if format == "" {
		format = "%va%v"
	}
	r, phi := cmplx.Polar(x)
	phi *= 180.0 / math.Pi
	if phi < 0 {
		phi += 360.0
	}
	if r == 0.0 {
		phi = 0.0
	}
	if phi == -0.0 || phi == 360.0 {
		phi = 0.0
	}
	return fmt.Sprintf(format, r, phi)
}
