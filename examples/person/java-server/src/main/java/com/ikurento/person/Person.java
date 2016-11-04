package com.ikurento.person;

import java.io.Serializable;

public class Person implements Serializable {
    private static final long serialVersionUID = 5282543653784523600L;
    private int age;
    private String name;
    private Address address;

    public Person(int age, String name, Address address) {
        this.age = age;
        this.name = name;
        this.address = address;
    }

    public int getAge() {
        return this.age;
    }

    public void setAge(int age) {
        this.age = age;
    }

    public String getName() {
        return this.name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Address getAddress() {
        return this.address;
    }

    public void setAddress(Address address) {
        this.address = address;
    }

    public String toString() {
        return "Age = " + age + ", Name = " + name + ", Address = " + address;
    }
}
