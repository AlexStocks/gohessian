package com.ikurento.hessian;

import java.util.ArrayList;
import java.util.List;
import java.util.Arrays;

import com.caucho.hessian.client.HessianProxyFactory;


public class HessianClient {
    public static void main(String[] args) {
        HessianProxyFactory factory = new HessianProxyFactory();
        factory.setDebug(true);

        HessianClient client = new HessianClient();

        try {
            String url = "http://localhost:8000/echo";
            Echo echo = (Echo) factory.create(Echo.class, url);
            client.testEcho(echo);

            url = "http://localhost:8000/math";
            Math math = (Math) factory.create(Math.class, url);
            client.testMath(math);

        } catch (Exception e) {
            System.out.println(e);
        }
    }

    private void testEcho(Echo echo) {
        System.out.println("test echo interface");

        System.out.println(echo.nullEcho());
        System.out.println(echo.boolEcho(true));
        System.out.println(echo.longEcho(31415926));
        System.out.println(echo.floatEcho(314));
        System.out.println(echo.doubleEcho(3.1415926));
        System.out.println(echo.dateEcho(""));
        // System.out.println(new String(Arrays.toString(echo.binaryEcho("A03B"))));
        System.out.println(new String(echo.binaryEcho("A03B")));
        System.out.println(echo.stringEcho("ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890"));

        List<Object> l = new ArrayList<Object>();
        l.add(100);
        l.add(10.001);
        l.add("hello"); // java.Object
        l.add(new byte[]{0, 2, 4, 6, 8, 10}); // java.Object
        l.add(true);
        l.add(null);
        l.add(false);
        System.out.println(echo.listEcho(l));
    }

    private void testMath(Math math) {
        System.out.println("\ntest echo interface");
        System.out.println(math.Add(1, 2));
        System.out.println(math.Sub(1, 2));
        System.out.println(math.Mul(1, 2));
        System.out.println(math.Div(4, 2));
    }
}

