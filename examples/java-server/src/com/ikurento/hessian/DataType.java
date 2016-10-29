// *****************************************************
// DESC    : interafce for http://localhost:8000/datatype
// AUTHOR  : Alex Stocks
// EMAIL   : alexstocks@foxmail.com
// MOD     : 2016-10-25 21:37
// FILE    : DataType.java
// ******************************************************

package com.ikurento.hessian;

import java.util.List;
import java.util.Map;

// client: github.com/AlexStocks/dubbogo/codec/hessian/client_test.go:TestDataType
public interface DataType {
    public String DTVoid();
    public String DTNull();
	public String DTBoolean(boolean b);
    public String DTInt(int i);
    public String DTDouble(double d);
    public String DTBytes(byte[] b);
    public String DTString(String s);
    public String DTList(List<Object> l);
    public String DTMap(Map<Object, Object> m);
}
