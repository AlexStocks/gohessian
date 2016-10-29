// *****************************************************
// DESC    : implementation of Math interface
// AUTHOR  : Alex Stocks
// EMAIL   : alexstocks@foxmail.com
// MOD     : 2016-10-25 21:43
// FILE    : MathImpl.java
// ******************************************************

package com.ikurento.hessian;

import com.caucho.hessian.server.HessianServlet;

public class MathImpl extends HessianServlet implements Math {
    public int Add(int x, int y) {
        System.out.println("add result:" + (x+y));
        return x + y;
    }

    public int Sub(int x, int y) {
        return x - y;
    }

    public int Mul(int x, int y) {
        return x * y;
    }

    public int Div(int x, int y) {
        if (y == 0) {
            return 0;
        }

        return x / y;
    }
}
