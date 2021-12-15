#include<stdlib.h>
#include<string.h>
#include"../k.h"

#include<stdio.h> //printf


// dgesv solve linear system (real)
//  x: L columns (input matrix)
//  y: F or L    (rhs or multi-rhs)
//  r: F or L    (result)
// zgesv                     (complex)
//  x: L columns (input matrix)
//  y: Z or L    (rhs or multi-rhs)
//  r: Z or L    (result)
//
// dgels solve overdetermined system (real)
//  x: L columns (input matrix)
//  y: F or L    (rhs or multi-rhs)
//  r: F or L    (result)
// zgels                             (complex)
//  x: L columns (input matrix)
//  y: Z or L    (rhs or multi-rhs)
//  r: Z or L    (result)
//
// dsyev symmetric eigenvalues         (real)
//  x: L columns (square input matrix, using upper triag)
//  r: F eigenvalues
// dsyeV symmetric eigenvalues+vectors (real)
//  x: L columns (square input matrix, using upper triag)
//  r: L (F eigenvalues;L eigenvectors)
//
// todo: dsyeV zhhev zhheV dgees dgeeS zgees zgeeS dgesvd dgesvD zgesvd zgesvD  
//
// complex numbers
//  backends that do not support complex types natively use F:
//  r0 i0 r1 i1 ..

// Example
//  A:3^?12    /4x3 matrix col-major
//  B:?4
//  dgels[A;B]
// lapack versions dgesv zgesv dgels zgels should return the same as solve[x;y] from ktye/z.k

/*
A:4^?16;b:?4;B:(?4;?4);dgesv[A;b];dgesv[A;B]
A:4^?32;b:?8;B:(?8;?8);zgesv[A;b];zgesv[A;B]
A:3^?12;b:?4;B:(?4;?4);dgels[A;b];dgels[A;B]
A:3^?24;b:?8;B:(?8;?8);zgels[A;b];zgels[A;B]
A:+(1.96 -6.49 -0.47 -7.20 -0.65;-6.49 3.80 -6.39 1.50 -6.34;-0.47 -6.39 4.17 -1.51 2.67;-7.20 1.50 -1.51 5.70 1.80;-0.65 -6.34 2.67 1.80 -7.10)
*/

int LAPACKE_dgesv(int, int, int, double*, int, int*, double*, int);
int LAPACKE_zgesv(int, int, int, double*, int, int*, double*, int);
int LAPACKE_dgels(int, char, int, int, int, double*, int, double*, int);
int LAPACKE_zgels(int, char, int, int, int, double*, int, double*, int);
int LAPACKE_dsyev(int, char, char, int, double*, int, double*);


static K* Lk(K x, size_t *n){
	  *n = NK(x);
	K *r = (K*)malloc(sizeof(K)**n);
	LK(r, x);
	return r;
}
static void ul(K *l, size_t n){
	for(int i=0; i<n; i++) unref(l[i]);
	free(l);
}
static double *Fk(K x, size_t *n){
	if(TK(x) != 'F') return NULL;
	*n=NK(x);
	double *r = malloc(8**n);
	memcpy(r, (double*)dK(x), 8**n);
	unref(x);
	return r;
}
#ifdef KTYE
extern char *_M;
#endif
static K KZ(double *x, size_t n, size_t z) {
	K r = KF(x, n);
#ifdef KTYE
	if(z > 1){
		*(int32_t*)(_M + (int32_t)r - 12) = (int32_t)(n / 2);
		return (((K)22)<<59)|(K)(int32_t)r;
	}
#endif
	return r;
}

static size_t rect(K *r, K x, size_t *cols){ // r:,/x  rows:#*x  cols:#x
	if(TK(x) != 'L'){ unref(x); return 0; }
	*cols = NK(x);
	x = Kx(",/", x);
	size_t n = NK(x);
	size_t rows = n / *cols;
	if (n != rows**cols) {
		unref(x);
		return 0;
	}
	*r = x;
	return rows;
}
static size_t square(K *r, K x){ // r:,/x cols:#x
	size_t cols;
	size_t rows = rect(r, x, &cols);
	if(!rows) return rows;
	if(rows!=cols){ unref(x); return 0; }
	return rows;
}

K solve(K x, K y, size_t z, int sq){
	size_t rows, cols;
	rows = rect(&x, x, &cols);
	if(cols == 0){ unref(y); return KE("gels: rect A"); }
	
	if(z*(rows/z) != rows){ unref(x); unref(y); return KE("gels: zlength A"); }
	rows /= z;
	if(rows < cols){ unref(x); unref(y); return KE("gels: length A"); }
	if((rows != cols)&&sq){ unref(x); unref(y); return KE("gesv: A square"); }
	
	double *A, *B;
	size_t xn;
	A = Fk(x, &xn);
	if(A == NULL){ unref(y); return KE("gels: type A"); }

	int rl = 0;
	size_t nrhs = 1;
	char yt = TK(y);
	if(yt == 'L'){ //multi-rhs
		rl = 1;
		size_t yn = rect(&y, y, &nrhs);
		if(nrhs == 0) { return KE("gels: rect B"); }
		yn /= z;
		if(yn != rows){ unref(y); return KE("gels: rows B"); }
		yt = TK(y);
	}
	K r;
	if(yt == 'F'){
		size_t yn;
		B = Fk(y, &yn);
		if(B == NULL){ free(A); return KE("gels: type B"); }
		yn /= z;
		if(yn/nrhs != rows){ free(A); free(B); return KE("gels: length B"); }
		int info;
		if(sq){
			int *ipiv = malloc(rows*sizeof(int));
			if(z > 1) info = LAPACKE_zgesv(102, (int)rows, (int)nrhs, A, (int)rows, ipiv, B, (int)rows);
			else      info = LAPACKE_dgesv(102, (int)rows, (int)nrhs, A, (int)rows, ipiv, B, (int)rows);
			free(ipiv);
		}else{
			if(z > 1) info = LAPACKE_zgels(102, 'N', (int)rows, (int)cols, (int)nrhs, A, (int)rows, B, (int)rows);
			else      info = LAPACKE_dgels(102, 'N', (int)rows, (int)cols, (int)nrhs, A, (int)rows, B, (int)rows);
		}
		if(info!=0){ free(A); free(B); return KE("lapack-dgels"); }
		free(A);
		if(rl) {
			K *l = (K*)malloc(nrhs * sizeof(K));
			for(int i=0;i<nrhs;i++){
				l[i] = KZ(B+i*z*rows, z*cols, z);
			}
			r = KL(l, nrhs);
			free(l);
		} else {
			r = KZ(NULL, z*cols, z);
			memcpy(dK(r), B, 8*z*cols);
		}
		free(A); free(B);
		return r;
	} else {
		unref(y); free(A); return KE("type dgels-B?");
	}
}
K eig(K x, size_t z, char jobz){
	size_t cols = square(&x, x);
	if(cols == 0){ return KE("eig A square"); }
	
	size_t xn;
	double *A = Fk(x, &xn);
	if(A == NULL){ return KE("eig: type A"); }
	
	int info;
	double *w = (double*)malloc(z*cols*sizeof(double));
	info = LAPACKE_dsyev(102, jobz, 'U', (int)cols, A, (int)cols, w);
	
	K r;
	K rw = KZ(w, z*cols, z);
	free(w);
	if(jobz == 'V'){
		K *l = (K*)malloc(cols*sizeof(K));
		for(int i=0;i<cols;i++){
			l[i] = KZ(A+i*z*cols, z*cols, z); 
		}
		K rl[2] = {rw, KL(l, cols)};
		free(l);
		r = KL(rl, 2);
	} else  r = rw;
	free(A);
	return r;
}
K dgesv(K x, K y){ return solve(x, y, 1, 1); }
K zgesv(K x, K y){ return solve(x, y, 2, 1); }
K dgels(K x, K y){ return solve(x, y, 1, 0); }
K zgels(K x, K y){ return solve(x, y, 2, 0); }
K dsyev(K x, K y){ return eig(x, 1, 'N'); }
K dsyeV(K x, K y){ return eig(x, 1, 'V'); }

void loadmat(){
 KR("dgesv", (void*)dgesv, 2);
 KR("zgesv", (void*)zgesv, 2);
 KR("dgels", (void*)dgels, 2);
 KR("zgels", (void*)zgels, 2);
 KR("dsyev", (void*)dsyev, 1);
 KR("dsyeV", (void*)dsyeV, 1);
}
