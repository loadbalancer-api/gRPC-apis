#!/bin/bash
set +x

export JRE_HOME=""

export PUBLISHER="radware"

export DATA_DIR="/var/opt/radware/license"

export FNLS_USER="root"

export FNLS_GROUP="root"

export JVMOPTS="-server -Xms2g -Xmx2g -XX:CompressedClassSpaceSize=64m -XX:MetaspaceSize=256m               -XX:+UseG1GC -XX:NewRatio=3 -XX:MaxGCPauseMillis=75             -XX:G1HeapWastePercent=10 -XX:InitiatingHeapOccupancyPercent=75 -XX:+CMSScavengeBeforeRemark                -XX:+IgnoreUnrecognizedVMOptions --add-opens=java.base/java.lang=ALL-UNNAMED --add-opens=java.base/java.lang.invoke=ALL-UNNAMED             -XX:+ScavengeBeforeFullGC -Djava.security.egd=file:/dev/./urandom "

export DEFINES="-Dbase.dir=/var/opt/radware/license"

export OPTIONS=""

export LD_LIBRARY_PATH="/opt/radware/license/radware"

