package com.ikurento.person;

import org.springframework.stereotype.Service;

@Service
public class BasicServiceImpl implements BasicService {
    public BasicServiceImpl() {
    }

    public void test(String a) {
        System.out.println(a);
    }

    public String test(Integer a) {
        System.out.println(a);
        return a + "";
    }

    public void test2() {
        byte a = 1;
        System.out.println(a);
    }

    public String test(Integer a, String b) {
        System.out.println(a + b);
        return a + b;
    }

    public String[] demo(String[] s) {
        System.out.println(s);
        return s;
    }
}
