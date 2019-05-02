ARG         repo=gitlab.com/sirlatrom/docker-secretprovider-plugin-vault/presentty-vault/installer

FROM        golang:1.12-alpine3.9 AS builder
ARG         repo
WORKDIR     /go/src
COPY        installer/vendor/ ./
RUN         CGO_ENABLED=0 go install -v ./...
WORKDIR     /go/src/$repo
COPY        installer/plugin_installer.go .
RUN         CGO_ENABLED=0 go install -v

FROM        registry.gitlab.com/almbrand/docker-stack-waiter AS docker-stack-waiter

FROM        docker:rc-git
RUN         apk add --no-cache iptables supervisor vim tmux bash
RUN         ln -f -s vim /usr/bin/vi
VOLUME      /var/lib/docker
COPY        supervisord.conf /etc/
ENTRYPOINT  ["/usr/bin/supervisord", "-n"]
COPY        .vimrc /root/
COPY        tmux.conf /root/.tmux.conf
ENV         DOCKER_BUILDKIT=1
COPY --from=builder /go/bin/installer /usr/local/bin/installer
COPY --from=docker-stack-waiter /docker-stack-waiter /usr/local/bin/docker-stack-waiter
COPY        demo.sh cleanup.sh docker-compose.yml /root/
WORKDIR     /root
