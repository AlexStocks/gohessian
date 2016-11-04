package com.ikurento.person;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import org.springframework.stereotype.Service;
import org.apache.log4j.Logger;

@Service
public class ClassServiceImpl implements ClassService {
     public static final Logger LOG = Logger.getLogger(ClassServiceImpl.class);

    public ClassServiceImpl() {
    }

    public String sayHello(String name) {
        System.out.println("length:" + name.length());
        return name;
    }

    public String helloWorld() {
        return "Hello World!";
    }

    public Person getPerson(String name, Address address, int age) {
        Person p = new Person(age, name, address);
        return p;
    }

    public List<Person> getPersons() {
        ArrayList ps = new ArrayList();

        for(int i = 0; i < 10; ++i) {
            Address address = new Address("SH" + i, "CN");
            Person p = new Person(20 + i, "foo" + i, address);
            ps.add(p);
        }

        return ps;
    }

    public Map<String, Person> getPersonMap() {
        HashMap map = new HashMap();
        Address address = new Address("SH", "CN");
        Address address2 = new Address("SH", "CN");
        Person p = new Person(21, "foo1", address);
        Person p2 = new Person(22, "foo2", address2);
        map.put("key1", p);
        map.put("key2", p2);
        return map;
    }

    public Person getPerson2(Person p) {
        LOG.warn("getPerson2(@p:%s)" + p);
        p.setName("foo");
        Address address = p.getAddress();
        address.setCountry("CN");
        p.setAddress(address);
        return p;
    }

    public Address getAddress(Address address) {
        System.out.println(address.getCountry() + address.getCity());
        address.setCity("SH");
        address.setCountry("CN");
        return address;
    }

    public List<Person> getPersons2(List<Person> persons) {
        int i = 0;
        Iterator var3 = persons.iterator();
        LOG.warn("@persons size:" + persons.size());

        while(var3.hasNext()) {
            Person person = (Person)var3.next();
            LOG.warn("getPerson2(@person:%s)" + person);
            System.out.println("getPerson2(@person:" + person + ")");
            StringBuilder var10001 = (new StringBuilder()).append("foo");
            ++i;
            person.setName(var10001.append(i).toString());
            ++i;
            person.setAge(100 + i);
        }

        return persons;
    }

    public Person getPersonDefalut() {
        Address address = new Address("SH", "CN");
        Person p = new Person(20, "foo", address);
        return p;
    }

    public List<Address> getAddresses() {
        ArrayList ps = new ArrayList();

        for(int i = 0; i < 10; ++i) {
            Address address = new Address("SH" + i, "CN");
            ps.add(address);
        }

        return ps;
    }

    public Address getAddressDefalut() {
        Address address = new Address("SH", "CN");
        return address;
    }
}
