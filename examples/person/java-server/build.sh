#!/usr/bin/env bash
# ******************************************************
# EMAIL   : alexstocks@foxmail.com
# FILE    : build.sh
# ******************************************************

# mvn -DdownloadSources=true eclipse:eclipse
# mvn -Dmaven.test.skip=true clean assembly:assembly
mvn clean package -Dmaven.test.skip=true
