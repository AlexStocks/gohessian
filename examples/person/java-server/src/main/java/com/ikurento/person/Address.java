package com.ikurento.person;

import java.io.Serializable;

public class Address implements Serializable {
    private static final long serialVersionUID = -2290457777731757372L;
    private String city;
    private String country;

    public Address(String city, String country) {
        this.city = city;
        this.country = country;
    }

    public String getCity() {
        return this.city;
    }

    public void setCity(String city) {
        this.city = city;
    }

    public String getCountry() {
        return this.country;
    }

    public void setCountry(String country) {
        this.country = country;
    }

    public String toString() {
        return "{country:" +  country + ", city:" + city + "}";
    }
}
