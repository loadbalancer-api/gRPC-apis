#!/bin/bash

cd /workspace/license-server-2.3.0-1 

/usr/bin/python3 /workspace/license-server-2.3.0-1/install.py --vdirect-username "root" --vdirect-password "C\!sc0123" --vdirect-port "2189" --ha-role "standalone" --cloud-sync "offline" --plugin-path "/workspace/license-server-2.3.0-1/plugins/"

source /workspace/radware/env_setup.sh

cd /workspace/license-server-2.3.0-1/server

ln -s /opt/radware/license/radware /workspace/license-server-2.3.0-1/server/radware

/usr/bin/java $JVMOPTS $DEFINES -jar flexnetls.jar $OPTIONS

