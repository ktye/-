package main

import . "github.com/ktye/wg/module"

func zk() {
	Data(600, "``x`y`z`k`l`a`b`while`\"rf.\"`\"rz.\"`\"uqs.\"`\"uqf.\"`\"gdt.\"`\"lin.\"`\"odo.\"`\"grp.\"\n`\"x.\":{,/+\"0123456789abcdef\"@(x%16;16/x:256/256+x)}\n`\"t.\":`45\n`\"p.\":`46\n`\"b.\":(`46)[`b;]\n`\"c.\":(`46)[`c;]\n`\"i.\":(`46)[`i;]\n`\"s.\":(`46)[`s;]\n`\"f.\":(`46)[`f;]\n`\"z.\":(`46)[`z;]\n`\"uqs.\":{x@&~0b~':x:^x}\n`\"uqf.\":{x@&(!#x)=x?x}\n`\"gdt.\":{[t;g](!#t)($[g;{x@>y x};{x@<y x}])/|.t}\n`\"odo.\":{{y@(#y)/!x}/:[*/x;&'x#'|*\\-1_1,|x]}\n`\"grp.\":{(x@*'g)!g:(&~x~':x i)^i:<x}\nany:`30;abs:`32;sin:`44;cos:`39;find:`31;fill:`38;imag:`33;conj:`34;angle:`35;exp:`42;log:`43\n`\"pad.\":{(|/#'x)#'x}\n`\"l.\":{\nkt:{[x;y;k;T]x:$[`T~@x;T[x;k];`pad(\"\";\"-\"),$x];(x,'\"|\"),'T[y;k]}\nd:{[x;k;kt;T]r:!x;x:.x;$[`T~@x;kt[r;x;k;T];,'[,'[`pad(k'r);\"|\"];k'x]]}\nT:{[x;k]$[`L':@'.x;,k x;(,*x),(,(#*x)#\"-\"),1_x:\" \"/:'+`pad@'$(!x),'.x]}\nt:@x;k:`kxy 1\ndd:(\"\";,\"..\")20<#x:$[(@x)':`L`D`T;x;x~*x;x;[t:`L;,x]]\nx:$[x~*x;x;(20&#x)#x]\n$[`D~t;d[x;k;kt;T];`T~t;T[x;k];x~*x;,k x;k'x],dd}\n`\"str.\":{q:{c,(\"\\\\\"/:(0,i)^@[x;i;(qs!\"tnr\\\"\\\\\")x i:&x':qs:\"\\t\\n\\r\\\"\\\\\"]),c:_34}\n$[|/x':\"\\t\\n\\r\"__!31;\"0x\",`x@x;q x]}\n`\"kxy.\":{\na:{t:@x;x:$x;$[`c~t;`str x;`s~t;\"`\",x;x]}\nd:{[x;k]r:\"!\",k@.x;n:#!x;x:k@!x;$[(1~n)|(@.x)':`D`T;\"(\",x,\")\";x],r}\nv:{[x;k;m;n]m*:(.`\".kstm\")t:@x; dd:(\"\";\"..\")m<#x;x:(m&#x)#x\nx:$[`L~t;k'x;`C~t;x;$x]\nx:$[`B~t;(*'x),\"b\";`C~t;`str x;`S~t;c,(c:\"`\")/:x;`L~t;$[1~n;*x;\"(\",(\";\"/:x),\")\"];\" \"/:x]\n((\"\";\",\")(1~n)),x,dd}\nt:@y;n:#y;k:`kxy x;m:x\n$[n~0;(.`\".kst0\")@t;`T~t;\"+\",d[+y;k];`D~t;d[y;k];y~*y;a y;v[y;k;m;n]]}\n`\".kst0\":`B`C`I`S`F`Z`L!(\"0#0b\";c,c:_34;\"!0\";\"0#`\";\"0#0.\";\"0#0a\";\"()\")\n`\".kstm\":`B`C`I`S`F`Z`L!100 100 30 30 20 10 20\n`\"k.\":`kxy 1000000\n`\"rf.\": {.5+(x?0)%4294967295.}\n`\"rf1.\":{.5+(1.+x?0)%4294967295.}        \n`\"rz.\": {(%-2*log `rf1 x)@360.*`rf x}\n`\"lin.\":{$[`L~@z;(.`\"lin.\")[x;y]'z;[dx:0.+1_-':x;dy:0.+1_-':y;b:(-2+#x)&0|x'z;(y b)+(dy b)*(z-x b)%dx b]]}\n`\"split.\":{$[`L~@x;`split@'x;\" \"\\:$[\" \"=x@-1+#x:x@&~i&~':i:\" \"=x;-1_x;x]]}\n`\"edit.\":{et:{\"+\",(`k@!x),\"!\",el@.x}\nel:{t:(`S`B`C`I`F`Z!``b`c`i`f`z)@@'x; (`k@t),\"$'+`split@'\\\"\\\\n\\\"\\\\:-1_1_\\\"\\n\",(\"\\n\"/:\" \"/:'+`pad@'$x),\"\\n\\\"\"}\n$[`T~t:@x;et x;`L~t;el x;`C~t;\"-1_1_\\\"\\n\",x,\"\\n\\\"\";\"*\",el@,x]}\ndot:{[xt;y]{+/x*y}\\:[xt;y]}\nsolve:{qslv:{H:x 0;r:x 1;n:x 2;m:x 3;j:0;K:!m\nwhile[j<n;y[K]-:(+/(conj H[j;K])*y K)*H[j;K];K:1_K;j+:1]\ni:n-1;J:!n;y[i]%:r@i\nwhile[i;j:i_J;i-:1;y[i]:(y[i]-+/H[j;i]*y@j)%r@i]\nn#y}\nq:$[`i~@*|x;x;qr x];$[`L~@y;qslv/:[q;y];qslv[q;y]]}\nqr:{K:!m:#*x;I:!n:#x;j:0;r:n#0a;turn:$[`Z~@*x;{(-x)@angle y};{x*1. -1@y>0}]\nwhile[j<n;I:1_I\nr[j]:turn[s:abs@abs/j_x j;xx:x[j;j]]\nx[j;j]-:r[j]\nx[j;K]%:%s*(s+abs xx)\nx[I;K]-:{+/x*y}/:[(conj x[j;K]);x[I;K]]*\\:x[j;K]\nK:1_K;j+:1];(x;r;n;m)}\navg:{(+/x)%0.+#x}\nvar:{(+/x*x:(x-avg x))%-1+#x}\nstd:{%var x}\nrem:{x/x+x/y}\nej:{(y j),'x_z i j:&~0N=i:(z x)?y x}\n`\"pack.\":{w:{(`c@,#x),x};($t),$[`s~t:@x;`pack@$x;x~*x;w `c@,x;`L~@x;(`c@,#x),,/`pack@'x;(@x)':`D`T;(`pack@.x),`pack@!x;`S~t;,/`pack@$x;w `c x]}\n`\"unpack.\":{s:x;g:{[n]r:n#s;s::n_s;r};n:{*`i@g 4};u:{x;$[(t:*g 1)':\"bcifz\";*(`$t)g n[];t~\"s\";`$u 0;t~\"S\";`$u 0;t~\"L\";u'!n[];t~\"D\";(u 0)!u 0;t~\"T\";+(u 0)!u 0;(`$_t+32)g n[]]};u 0}\ncsv:{c:{s:`$'x@i:&x':\"ifzs\";n:`i$\" \"\\:-1_@[x;i;\" \"];y[a]:(y[a],''\"a\"),''y[1+a:&s=`z];s$'y n};s:$[\" \"~(*x);`split@;(*x)\\:];x:1_x;y:+s'$[`L~@y;y;\"\\n\"\\:y];$[#x;c[x;y];y]}\nucal:{[s;u;r](#u)#+solve[u,0a+r=/:?r;s]}\npcal:{[s;u]+solve[+u;+s]}\nuslv:{[qrk;s]qrsolve[qrk;s]}\nuidx:{[u;a]solve[((#u)#1a;1@a);u]}\nPW:800;PH:600;FH:20;FONT:\"monospace\"\n`\"pltnn.\":{wi:</</:;$[#i:&wi[*x;+\\plt`px`pw]&wi[x 1;+\\plt`py`ph];*i;0]}\n`\"pltco.\":{[p;x;y]w:p`fh;h:p`fh;X:p`px;Y:p`py;W:p`pw;H:p`ph;C:(X+W%2;Y+H%2);R:(W%2)&(H%2)-h;d:$[`xy~p`t;(X+w;X+W-w;Y+H-h;Y+h);((C-R),C+R)0 2 3 1];a:p`a; ((d 0 1)(a 0 1)'x;(d 2 3)(a 2 3)'y)}\n`\"pltcl.\":{[x;y]p:.`\"pltco.\";n:`pltnn(x;y); xy:p[plt n;x;y]; `<$[`polar~plt[n;`t];$imag/|xy;\"x:\",($xy 0), \" y:\",($xy 1)],_10 32}\n`\"pltzo.\":{[x;y;w;h]n:`pltnn(x;y);p:.`\"pltco.\";xy0:p[plt n;x;y+h]; xy1:p[plt n;x+w;y]; plt[`a;n]:(xy0,xy1)0 2 1 3; plt[`t;n]:`xy; draw[`plts@`plt;(PW;PH)]}\nplot:{n:#x;i:!0;r:!0\nabsa:{af:{v:.x;(!x)!(*v;y v 1)};am:{af[x;abs]};an:{af[x;angle]}\ni::^(&{$[1~#x;0b;`Z~@$[L~@v:(.x)1;*v;v]]}'x),!#x\nr::&1_~':i;x:x i;x[r]:am'x r;x[1+r]:an'x 1+r;x}\nmult:{x[`pw]%:n;x[`px]:i*x`pw;x[r;`ph]:3\\2*PH;x[1+r;`py`ph]:(3\\2*PH;3\\PH);x[`a;1+r;2 3]:,0 360.;x}\nplt::mult@(`plot@)'absa$[(@x)':`L`T;*(x;n:#x);,x]\nShow[draw[`plts@`plt;(PW;PH)];.`\"pltcl.\";.`\"pltzo.\"]}\n`\"plot.\":{[d]l:$!d;v:.d; t:$[2~#d;`xy;`polar];\ny:$[t~`xy; $[`L~@y:v 1;y;,y];          $[`L~@y:_*v;y;,y]]\nx:$[t~`xy; $[`L~@x:v 0;x;(,x)@(#y)#0]; $[`L~@x:imag@*v;x;,x]]\nxt:`tics(&/&/x;|/|/x);yt:`tics(&/&/y;|/|/y)\na:$[t~`xy;(xt 0;*-1#xt;yt 0;*-1#yt);(-a;a;-a;a:*|`tics@0.,|/|/abs@*v)]\nc:c@(#c:11826975 950271 2924588 2631638 12412820 4937356 12744675 8355711 2276796 13614615)/!#x\nstyle:$[t~`polar;\"..\";`i~@**y;\"||\";\"--\"]\nsize: $[t~`polar;2;style~\"||\";(--/((**x),-1#*x))%-1+#*x ;2]\nlines:{`style`size`color`x`y!(style;size;z;x;0.+y)}'[x;y;c]\npw:PW;ph:PH;`L`T`t`l`a`f`fh`px`py`pw`ph!(lines;\"\";t;l;a;FONT;FH;0;0;pw;ph)}\n`\"plts.\":{[sym];x:.sym;$[`D~@x;`Plot x;,/(`Plot@)'x]}\n`\"Plot.\":{[x];w:x`fh; h:x`fh; X:x`px; Y:x`py; W:x`pw; H:x`ph; a:x`a;T:x`T;grey:13882323\nC:(X+W%2;Y+H%2);R:(W%2)&(H%2)-h\ndst:$[`xy~x`t;(X+w;X+W-w;Y+H-h;Y+h);((C-R),C+R)0 2 3 1];rdst:(X+w;Y+h;W-2*w;H-2*h)\nxs:(a 0 1)(dst 0 1)'\nys:(a 2 3)(dst 2 3)'\nbars:{[l]$[\"|\"':l`style;(`color;l`color),,/{(`Rect;((-dx%2)+xs x;ys y;dx:-/xs(l`size;0.);(ys a 2)-ys y))}'[l`x;l`y];()]}\nline:{[l]$[\"-\"':l`style;(`linewidth;l`size;`color;l`color;`poly;(xs l`x;ys l`y));()]}\ndots:{[l]$[\".\"':l`style;(`color;l`color),,/{(`Circle;(xs x;ys y;1.5*l`size))}'[l`x;l`y];()]}\nc:(`clip;(X;Y;W;H);`font;(x`f;x`fh);`color;0;`text;((X+W%2;Y+h);1;T))\nxy:{[]c,:(`text;((X+w;Y+H);0;$a 0);`text;((X+W%2;Y+H);1;(x`l)0);`text;((X+W-w;Y+H);2;$a 1))\nc,:(`Text;((X+w;Y+H-h);0;$a 2);`Text;((X+w;Y+H%2);2;(x`l)1);`Text;((X+w;Y+h);2;$a 3))\nc,:(`color;0;`linewidth;2;`rect;rdst)\nc,:(`linewidth;1;`color;grey)\nc,:(`clip;rdst)\nc,:,/{(`line;0.+(x;dst 2;x;dst 3))}'xs`tics x[`a;0 1]\nc,:,/{(`line;0.+(dst 0;x;dst 1;x))}'ys`tics x[`a;2 3]}\npo:{[]c,:(`text;((C 0;Y+H);1;(x`l)0);`text;(C+.75*R;6;$(x`a)3))\nc,:(`font;(x`f;_h*.8)),,/{(`text;(C+R*(_;imag)@'x;y;z))}'[1@270.+a;0 0 6 6 4 4 2 2;$a:30 60 120 150 210 240 300 330]\nc,:(`color;0),/{(`line;,/+C+(R-w%2;R)*/:(_;imag)@'x)}'1@30.*!12\nc,:(`color;grey;`linewidth;1;`line;((-R)+*C;C 1;R+*C;C 1);`line;(*C;(-R)+C 1;*C;R+C 1))\nc,:,/{(`circle;0.+C,x)}'r:(xs@`tics 0.,x[`a;3])-*C\nc,:(`color;0;`linewidth;2;`circle;C,R)}\n$[`xy~x`t;xy[];po[]]\nc,:,/bars'x`L\nc,:,/line'x`L\nc,:,/dots'x`L}\n`\"tics.\":{[minmax]nice:{[x;r]f:x%0.+10^ex:_log[10;x];(1 2 5 10.@*&(~f>1 2 5 0w;f<1.5 3 7 0w)[r])*10^ex};e:nice[-/|minmax;0];s:nice[e%4.;1];n:_1.5+e%s;$[~(minmax 1)>*-2#r:s*(_(*minmax)%s)+!n;-1_r;r]}\n`\"ceg.\":{(x i)!0-':1+i:&(1_~~':x),1b}\nhist:{$[`i~@x;hist[(x;&/y;|/y);y];(Y;(`38)[0;(`ceg@^1+((d%2.0)+-1_Y:(x 1)+(d:(--/1_x)%-1.+n)*!n)'y)@!n:_0.+*x])]}\n")
	zn := int32(6756) // should end before 8k
	x := mk(Ct, zn)
	Memorycopy(int32(x), 600, zn)
	dx(Val(x))
}
