ini:I:I{0::1130366807310592j;128::x;p:256;i:8;(i<x)?/((4*i)::p;p::i;p*:2;i+:1);x} /x:16(64k)
mk:I:II{t:x bk y;i:4*t;(~I i)?/i+:4;(128~i)?!;i::I 4+a:I i;j:i-4;(j>=4*t)?/(j-:4;u:a+1<<j%4;u::I j;j::u);a::y|x<<29;(a+4)::1;a}
mki:I:I{r:2 mk 1;(r+4)::1;(r+8)::x;r}mkd:I:II{r:6 mk 1;(8+r)::x;(12+r)::y;decr x;decr y;r}
fr:0:I{v1;t:4*xt bk xn;x::I t;t::x}bk:I:II{32-*7+y*C x}decr:0:I{(x>255)?(xr:I x+4;(x+4)::xr-1;(1~xr)?fr x)}dxr:{decr x;r}inc:I:I{(x+4)::1+I x+4;x}lnc:0:II{x+:8;y/(x:inc x;x+:4)}
v1:{xt:(I x)>>29;xn:(I x)&536870911;xp:8+x}
lrc:I:II{!;1}drc:I:II{!;1} /nyi
til:I:I{v1;(~2~xt)?!;r:xt mk n:I xp;rp:8+r;n/(rp::i;rp+:4);dxr}
fst:I:I{v1;(~xt)? :x;(7~xt)? :x;r:xt mk 1;xt?[;(r+8)::C xp;;(r+8)::J xp;((r+8)::J xp;(r+16)::J xp+8);;;!;(r+8)::I xp];(xt~6)?xn lnc x+8;dxr}
rev:I:I{v1;xt?[!;;;;;;r:((inc xp)mkd inc xp+4);r:xt mk xn];xt?[!;(rp:r+7+xn;xn/((rp-i)::C xp+i));;(rp:r+8*xn;xn/(rp::J xp;rp-:8;xp+:8));(rp:r-8+16*xn;xn/(rp::J xp;(8+rp)::J 1+xp;rp-:16;xp+:16));;((r+8)::rev r+8;(r+12)::rev r+12);;;(rp:r+4+4*xn;xn/(rp::I xp;xp+:4;rp-:4))];dxr}
cnt:I:I{v1;decr x;mki xn}
tip:I:I{v1;r:2 mk 2;(8+r)::xt;(12+r)::xn;dxr}
sumi:I:II{y/r+:I x+1;r}
wer:I:I{v1;(~2~xt)?!;xn/(0>'I xp+i)?!;r:2 mk xn sumi 8+xp;rp:8+r;xn/(I xp)/(rp::i;rp+:4);dxr}
enl:I:I{v1;r:1 mk 6;(8+r)::x;r}
neg:I:I{v1;xt?[!;!;;;;!; :45 lrc x; :45 drc x;r:xt mk xn];decr x;(2~xt)?(rp:r+8;xn/(rp::0-I xp;rp+:4;xp+:4); :r);(4~xt)?(xt:3;xn*:2);rp:8+r;xn/(rp::-F xp;rp+:4;xp+:4);r}

\
01234567   xt:x>>29       xn:x&536870911 (-1+1<<29)
Fcifzsld   xt~0(function) x<256(basic) 
0148x440   ft~xn&0xff00  (derived, proj, lambda, native)  composition==lambda?
	   fn~xn&0xff    (argn)

+ add abs                 memory
- sub neg                   0..  7   type sizes   0 1 4 8 16 4 4 0
* mul fst                   8.. 11   k-tree/key   pointer     todo
% div sqr                  12.. 15   k-tree/value pointer     todo
& min wer                  16..127   free pointers (4*i) for bt i, i:4..31
| max rev                 128..131   memsize log2
< les gup                 
> mor gdn                 
= eql grp                 
~ mtc not                 
! key til                 
, cat enl                 
^ exc asc                 
$ str cst   
# tak cnt
_ drp flr                 
? fnd unq                 
@ atx typ                 
. cal val                 