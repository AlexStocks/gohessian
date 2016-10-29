#!/usr/bin/env bash
# ******************************************************
# DESC    : build script for java-server
# AUTHOR  : Alex Stocks
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-10-26 20:28
# FILE    : build-linux.sh
# ******************************************************

#rm -f *.class
rm ./HessianServer.jar
cp="./"
for j in $(ls ../java-lib)
do
  cp+=":../java-lib/$j"   # windows中用分号，linux中用冒号
done

JAVA_OPT=" -DDEBUG -server -Dorg.eclipse.jetty.util.log.class=org.eclipse.jetty.util.log.Slf4Log "
JAVA_OPT=" $JAVA_OPT -Dorg.eclipse.jetty.LEVEL=INFO -DServiceImpl.LEVEL=DEBUG -DService.LEVEL=DEBUG "

javac -classpath $cp ./src/com/ikurento/hessian/*.java -d ./ # ./src/com/ikurento/hessian/Service.java ./src/com/ikurento/hessian/ServiceImpl.java -d ./
# jar cvfm HessianServer.jar ./com/ikurento/hessian/*.class
# java $JAVA_OPT -classpath $cp HessianServer
jar cvmf ./META-INF/MANIFEST.MF HessianServer.jar ./com/ikurento/hessian/* # ../java-lib/*
# jar -xf HessianServer.jar
rm -rf ./com
cp+=":./HessianServer.jar"
java $JAVA_OPT -classpath $cp com.ikurento.hessian.HessianServer
rm ./HessianServer.jar
