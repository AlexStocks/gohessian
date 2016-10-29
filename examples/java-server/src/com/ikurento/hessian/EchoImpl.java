// *****************************************************
// MOD     : 2016-10-26 11:15
// FILE    : EchoImpl.java
// ******************************************************

package com.ikurento.hessian;

import java.util.Arrays;
import java.util.Date;
import java.util.List;
import java.util.Iterator;

import com.caucho.hessian.server.HessianServlet;

public class EchoImpl extends HessianServlet implements Echo {
    public Object nullEcho() {
        return null;
    }

    public boolean boolEcho(boolean b) {
        return b;
    }

    public long longEcho(long l) {
        return l;
    }

    public Date dateEcho(String dateStr) { // 2014-12-01 16:17:18
        return new Date();
    }

    public float floatEcho(float f) {
        return f;
    }

    public double doubleEcho(double d) {
        return d;
    }

    // like 0a 0b
    public byte[] binaryEcho(String hex) {
        System.out.println("fuck..............:" + hex);
        // return hexStringToByteArray(hex);
        return hex.getBytes();
    }

    public String stringEcho(String s) {
        return s;
    }

    public String listEcho(List<Object> l) {
        System.out.println("List:");
        Iterator<Object> it = l.iterator();
        String s = "List{";
        System.out.println("list size:" + l.size());
        while (it.hasNext()) {
            Object o = it.next();
            if (o == null) {
                // System.out.println(o.getClass() + "\t" + "null"); // object java.lang.NullPointerException
                System.out.println("\t" + "null");
                s += "null";
            } else if (o instanceof byte[] && ((byte[])o).length != 0) {
            // if (o.getClass().isArray()) {
            // if (o instanceof Object[]) {
                // Java byte[](int []) is a special Object which does not have toString().
                // System.out.println(o.getClass() + "\t" + o); // output：class [B        [B@15b93ce // output is o's memory address.
                System.out.println("byte[]" + "\t" + Arrays.toString((byte[])o)); // output: byte[]  [0, 2, 4, 6, 8, 10]
                // System.out.println("byte[]" + "\t" + new String((byte[])o));
                s += "byte[]";
            } else {
                System.out.println(o.getClass() + "\t" + o);
                s += o.getClass().getName();
            }
            if (it.hasNext()) {
                s += ",";
            }
        }
        s += "}";

        System.out.println("last result:" + s);
        return s;
    }

    private byte[] hexStringToByteArray(String s) {
        byte[] b = new byte[s.length() / 2];
        for (int i = 0; i < b.length; i++) {
            int index = i * 2;
            int v = Integer.parseInt(s.substring(index, index + 2), 16);
            b[i] = (byte) v;
        }
        return b;
    }
}
