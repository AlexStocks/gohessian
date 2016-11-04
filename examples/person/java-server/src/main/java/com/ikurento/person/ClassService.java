package com.ikurento.person;

import java.util.List;
import java.util.Map;

public interface ClassService {
    String sayHello(String var1);

    String helloWorld();

    Address getAddressDefalut();

    Address getAddress(Address var1);

    Person getPersonDefalut();

    Person getPerson(String var1, Address var2, int var3);

    Person getPerson2(Person var1);

    List<Address> getAddresses();

    List<Person> getPersons();

    Map<String, Person> getPersonMap();

    List<Person> getPersons2(List<Person> var1);
}
