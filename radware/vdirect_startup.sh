#!/bin/bash
cd /opt/radware/vdirect/server
/usr/bin/java -Xmx1024M -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=./logs/application/ -Dvdirect.path=/opt/radware/vdirect/database -Dvdirect.quiet=true -cp vdirect-server-start.jar:lib/* com.radware.vdirect.server.Main
Restart=always &

