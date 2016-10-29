#!/usr/bin/env bash
# ******************************************************
# DESC    : build script for java-client
# AUTHOR  : Alex Stocks
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-10-26 20:27
# FILE    : build-linux.sh
# ******************************************************

#rm -f *.class
rm HessianClient.jar
cp="./"
for j in $(ls ../java-lib)
do
  cp+=":../java-lib/$j"   # windows���÷ֺţ�linux����ð��
done

JAVA_OPT=" -DDEBUG -server -Dorg.eclipse.jetty.util.log.class=org.eclipse.jetty.util.log.Slf4Log "
JAVA_OPT=" $JAVA_OPT -Dorg.eclipse.jetty.LEVEL=INFO -DServiceImpl.LEVEL=DEBUG -DService.LEVEL=DEBUG "

javac -classpath $cp ./src/com/ikurento/hessian/*.java -d ./
# jar cvfm HessianClient.jar ./com/ikurento/hessian/*.class
# java $JAVA_OPT -classpath $cp HessianClient
jar cvmf ./META-INF/MANIFEST.MF HessianClient.jar ./com/ikurento/hessian/* # ../java-lib/*
# jar -xf HessianClient.jar
rm -rf ./com
cp+=":./HessianClient.jar"
java $JAVA_OPT -classpath $cp com.ikurento.hessian.HessianClient
rm HessianClient.jar
