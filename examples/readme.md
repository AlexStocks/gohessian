# caculator #
---
*witten by AlexStocks on 2016/10/25*

---
## readme
Caculator package is used to test hessian2 protocol(github.com/AlexStocks/dubbogo/codec/hessian).

*这个程序主要是为了测试github.com/AlexStocks/codec/hessian/client_test.go,之所以取名calculator是因为服务端有math 8interface,无其他意义。*

测试的时候注意修改 github.com/AlexStocks/dubbogo-examples/calculator/java-server/src/com/ikurento/hessian:HessianServer.java中与url相关的设置。具体来说是，port(HessianServer.java:line 21) 和 path(HessianServer.java:line 40-42).

无论是java-server还是java-client，只需执行sh build-winidows.sh(linux: sh build-linux.sh)即可。

---
## java server & java client

The java-server(Echo Interface) & java-client refers to [chenluyj/gohessian/test](https://github.com/chenluyj/gohessian/tree/master/test). Thanks to [chenluyj](https://github.com/chenluyj)'s great work.

Echo Interface Client: github.com/AlexStocks/gohessian/examples/java-client

DataType Interface Client: github.com/AlexStocks/gohessian/client_test.go:TestDataType

Math Interface Client: github.com/AlexStocks/gohessian/client_test.go:TestMath
