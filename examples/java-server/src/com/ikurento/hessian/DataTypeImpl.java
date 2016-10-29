// *****************************************************
// DESC    : implementation of DTType interface
// AUTHOR  : Alex Stocks
// EMAIL   : alexstocks@foxmail.com
// MOD     : 2016-10-25 21:43
// FILE    : DataTypeImpl.java
// ******************************************************

package com.ikurento.hessian;

import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.Iterator;
import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;

import com.caucho.hessian.server.HessianServlet;

public class DataTypeImpl extends HessianServlet implements DataType {
    public String DTVoid() { return "void"; }

    public String DTNull()  { return "null"; }

    public String DTBoolean(boolean b) {
        return "boolean";
    }

    public String DTInt(int i) { return "int"; }

    public String DTDouble(double d) {
        return "double";
    }

    public String DTBytes(byte[] b) {
        return "byte[]";
    }

    public String DTString(String s) { return "string"; }

    public String DTList(List<Object> l) {
        String s = "List{";
        System.out.println("List:");
        Iterator<Object> it = l.iterator();
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

        /*
        File file;
        FileOutputStream fop;
        try {
            file = new File("C:\\Users\\AlexStocks\\share\\securecrt\\test.txt");
            fop = new FileOutputStream(file);
            if (!file.exists()) {
                file.createNewFile();
            }
            fop.write(s.getBytes());
            fop.flush();
            fop.close();
        } catch (IOException e) {
            e.printStackTrace();
        } finally {
        }
        */

        return s;
    }

    public String DTMap(Map<Object, Object> m)  {
        String s = "Map{}";
        System.out.println("Map: size" + m.size());
        for (Map.Entry<Object, Object> entry:m.entrySet()) {
            System.out.println("\t" + entry.getKey()+":"+entry.getValue());
        }

        return s;
    }
}
