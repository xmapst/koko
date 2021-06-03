FROM golang:1.15-alpine as stage-build
LABEL stage=stage-build

WORKDIR /opt/koko
ARG TARGETARCH=amd64
ARG GOPROXY=https://goproxy.io
ARG VERSION

ENV TARGETARCH=$TARGETARCH \
    GOPROXY=$GOPROXY \
    VERSION=$VERSION \
    GO111MODULE=on \
    GOOS=linux \
    CGO_ENABLED=0

COPY . .
RUN wget https://github.com/jumpserver/koko/releases/download/${Version}/koko-${Version}-linux-${TARGETARCH}.tar.gz -O kubectl.tar.gz && tar -xzf kubectl.tar.gz \
    && chmod +x kubectl && mv kubectl rawkubectl \
    && wget http://download.jumpserver.org/public/kubectl_aliases.tar.gz -O kubectl_aliases.tar.gz && tar -xzvf kubectl_aliases.tar.gz \
    && cd utils && sh -ixeu build.sh

FROM debian:stretch-slim
ENV LANG en_US.utf8
RUN sed -i  's/deb.debian.org/mirrors.163.com/g' /etc/apt/sources.list \
    && sed -i  's/security.debian.org/mirrors.163.com/g' /etc/apt/sources.list \
    && apt-get update -y \
    && apt-get install -y locales \
    && localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8 \
    && apt-get install -y --no-install-recommends gnupg dirmngr openssh-client procps curl \
    && apt-get install -y --allow-unauthenticated --no-install-recommends mariadb-client \
    && apt-get install -y --no-install-recommends gdb ca-certificates jq iproute2 less bash-completion unzip sysstat acl net-tools iputils-ping telnet dnsutils wget vim git \
    && rm -rf /var/lib/apt/lists/*

ENV TZ Asia/Shanghai
WORKDIR /opt/koko/
COPY --from=stage-build /opt/koko/release/koko /opt/koko
COPY --from=stage-build /opt/koko/release/koko/kubectl /usr/local/bin/kubectl
COPY --from=stage-build /opt/koko/rawkubectl /usr/local/bin/rawkubectl
COPY --from=stage-build /opt/koko/utils/coredump.sh .
COPY --from=stage-build /opt/koko/entrypoint.sh .
COPY --from=stage-build /opt/koko/utils/init-kubectl.sh .
COPY --from=stage-build /opt/koko/.kubectl_aliases /opt/kubectl-aliases/.kubectl_aliases

RUN chmod 755 entrypoint.sh && chmod 755 init-kubectl.sh

EXPOSE 2222 5000
CMD ["./entrypoint.sh"]
