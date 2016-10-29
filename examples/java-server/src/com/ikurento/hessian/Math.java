// *****************************************************
// DESC    : interafce for http://localhost:8000/math
// AUTHOR  : Alex Stocks
// EMAIL   : alexstocks@foxmail.com
// MOD     : 2016-10-25 21:37
// FILE    : Math.java
// ******************************************************

package com.ikurento.hessian;

// client: github.com/AlexStocks/dubbogo/codec/hessian/client_test.go:TestMath
public interface Math {
	public int Add(int x, int y);
	public int Sub(int x, int y);
	public int Mul(int x, int y);
	public int Div(int x, int y);
}
