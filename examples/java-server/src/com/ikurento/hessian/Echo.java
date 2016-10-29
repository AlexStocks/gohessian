// *****************************************************
// MOD     : 2016-10-26 11:15
// FILE    : Echo.java
// ******************************************************

package com.ikurento.hessian;

import java.util.Date;
import java.util.List;

// client: github.com/AlexStocks/dubbogo-examples/calculator/java-client
public interface Echo {
    public Object nullEcho();
    public boolean boolEcho(boolean b);
    public long longEcho(long l);
    public Date dateEcho(String dateStr);// 2014-12-01 16:17:18
    public float floatEcho(float f);
    public double doubleEcho(double d);
    public byte[] binaryEcho(String hex);
    public String stringEcho(String s);
    public String listEcho(List<Object> ls);
}
