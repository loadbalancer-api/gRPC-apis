FROM golang:1.13.6 AS builder
ARG AUSER
ARG AKEY
ARG APackage
ENV creds=$AUSER:$AKEY
ENV artifactory=$APackage
WORKDIR /workspace
ADD https://$creds@$artifactory/vdirect-server-install-deb-4-12-0-1.deb  /workspace/

ADD https://$creds@$artifactory/license-server-2-3-0-1.tgz  /workspace/

RUN apt-get clean && rm -rf /var/lib/apt/lists/* && apt-get update \
    && apt-get install -y --no-install-recommends \
    supervisor \
    default-jre \
    zip \
    python3 python3-pip \
    curl /workspace/vdirect-server-install-deb-4-12-0-1.deb

RUN pip3 install requests

WORKDIR /workspace/api
COPY api/lbservice/ ./lbservice

WORKDIR /workspace/radware/server
COPY radware/server/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o ./server server.go

WORKDIR /workspace/radware
COPY radware/vdirect_startup.sh ./
RUN chmod 777 vdirect_startup.sh
COPY radware/lls_startup.sh ./
COPY radware/env_setup.sh ./
RUN chmod 777 lls_startup.sh
RUN chmod 777 env_setup.sh

WORKDIR /workspace/radware/workflow_templates
COPY radware/workflow_templates/ ./

RUN echo root:C\!sc0123 | chpasswd

WORKDIR /workspace
COPY supervisord.conf ./
#Install local license server "LLS"
RUN tar -zxf  license-server-2-3-0-1.tgz
RUN echo root:C\!sc0123 | chpasswd

ENTRYPOINT ["/bin/bash", "-c", "supervisord -c /workspace/supervisord.conf"]
