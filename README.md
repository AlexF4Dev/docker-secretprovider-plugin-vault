# Sample Docker/Vault secrets plugin

TODO:

- [ ] Docs
- [ ] Make TLS configurabl
- [ ] CI/CD
- [ ] Use docker-compose for teardown/bootstrap

Preliminary instructions:

Run `./rebuild.sh`, you should get the following

(Note the different tokens in the two task instances of the `snitch` service).

```console
$ ./rebuild.sh
Error: No such container: vault
Error: No such service: snitch
Error: No such service: vault-helper-service
Error: No such secret: secret-zero
Error: No such secret: haha
Error: No such secret: VAULT_TOKEN_SORT_OF
Error response from daemon: plugin "<redacted_registry>/absukl/secrets-plugin:latest" not found
Error: No such plugin: <redacted_registry>/absukl/secrets-plugin:latest
Sending build context to Docker daemon  14.82MB
Step 1/11 : FROM golang:1.11rc1-alpine AS build
 ---> c5ca90186c47
Step 2/11 : ARG repo=${REGISTRY}/absukl/secrets-plugin
 ---> Using cache
 ---> 08add4d33da0
Step 3/11 : WORKDIR /go/src
 ---> Using cache
 ---> 386553a66e78
Step 4/11 : COPY vendor/ ./
 ---> Using cache
 ---> fdf99e385796
Step 5/11 : RUN CGO_ENABLED=0 go install -v ./...
 ---> Running in 5ced50828431
github.com/docker/docker/api/types/network
github.com/docker/docker/api
github.com/docker/docker/api/types/events
github.com/docker/docker/api/types/image
github.com/hashicorp/hcl/hcl/strconv
golang.org/x/sys/windows
golang.org/x/sys/unix
github.com/Microsoft/go-winio
github.com/pkg/errors
google.golang.org/grpc/codes
github.com/golang/protobuf/proto
golang.org/x/net/context
github.com/opencontainers/go-digest
github.com/opencontainers/image-spec/specs-go
net
github.com/opencontainers/image-spec/specs-go/v1
github.com/docker/distribution/digestset
github.com/docker/distribution/reference
github.com/docker/docker/api/types/blkiodev
golang.org/x/crypto/ssh/terminal
github.com/docker/docker/api/types/mount
github.com/docker/docker/api/types/strslice
github.com/docker/go-units
github.com/sirupsen/logrus
github.com/docker/docker/api/types/versions
github.com/docker/docker/api/types/filters
github.com/containerd/containerd/log
github.com/gogo/protobuf/proto
github.com/docker/docker/api/types/time
github.com/docker/docker/pkg/stdcopy
github.com/golang/snappy
github.com/golang/protobuf/ptypes/any
github.com/golang/protobuf/ptypes/duration
github.com/golang/protobuf/ptypes/timestamp
google.golang.org/genproto/googleapis/rpc/status
crypto/x509
github.com/docker/go-connections/nat
github.com/golang/protobuf/ptypes
github.com/docker/docker/api/types/container
google.golang.org/grpc/status
github.com/docker/docker/api/types/registry
github.com/containerd/containerd/errdefs
golang.org/x/net/internal/socks
github.com/containerd/containerd/platforms
golang.org/x/net/proxy
net/textproto
crypto/tls
vendor/golang_org/x/net/http/httpproxy
vendor/golang_org/x/net/http/httpguts
github.com/hashicorp/errwrap
github.com/hashicorp/go-sockaddr
github.com/hashicorp/go-multierror
github.com/hashicorp/hcl/hcl/token
github.com/hashicorp/hcl/hcl/ast
github.com/hashicorp/hcl/hcl/scanner
github.com/hashicorp/hcl/hcl/parser
github.com/docker/docker/api/types/swarm/runtime
github.com/hashicorp/hcl/json/token
github.com/docker/docker/api/types/swarm
github.com/coreos/go-systemd/activation
net/http/httptrace
github.com/docker/go-connections/tlsconfig
github.com/docker/docker/api/types
github.com/hashicorp/go-rootcerts
github.com/hashicorp/hcl/json/scanner
net/http
github.com/hashicorp/vault/helper/hclutil
github.com/hashicorp/vault/helper/compressutil
github.com/hashicorp/hcl/json/parser
github.com/hashicorp/vault/helper/jsonutil
github.com/ryanuber/go-glob
github.com/docker/docker/api/types/volume
github.com/hashicorp/vault/helper/strutil
github.com/mitchellh/mapstructure
github.com/hashicorp/hcl
golang.org/x/text/transform
golang.org/x/text/unicode/bidi
golang.org/x/text/unicode/norm
github.com/hashicorp/vault/helper/parseutil
golang.org/x/net/http2/hpack
golang.org/x/text/secure/bidirule
golang.org/x/time/rate
github.com/mitchellh/go-homedir
golang.org/x/net/idna
golang.org/x/net/http/httpguts
golang.org/x/net/context/ctxhttp
github.com/docker/go-connections/sockets
github.com/hashicorp/go-cleanhttp
net/http/httputil
github.com/hashicorp/go-retryablehttp
golang.org/x/net/http2
github.com/docker/go-plugins-helpers/sdk
github.com/docker/go-plugins-helpers/secrets
github.com/docker/docker/client
github.com/hashicorp/vault/api
Removing intermediate container 5ced50828431
 ---> 517f76a3f98a
Step 6/11 : WORKDIR /go/src/$repo
 ---> Running in 158bee5229f7
Removing intermediate container 158bee5229f7
 ---> f5fd8bc42287
Step 7/11 : COPY *.go .
 ---> b673b54b4b10
Step 8/11 : RUN CGO_ENABLED=0 go install -v
 ---> Running in 49ecc8483d0f
<redacted_host>/absukl/secrets-plugin
Removing intermediate container 49ecc8483d0f
 ---> 398ecab4fe46
Step 9/11 : FROM scratch
 --->
Step 10/11 : COPY --from=build "/go/bin/secrets-plugin" "/go/bin/secrets-plugin"
 ---> 6c3edb4a3ada
Step 11/11 : ENTRYPOINT ["/go/bin/secrets-plugin"]
 ---> Running in 3166d4b94c1a
Removing intermediate container 3166d4b94c1a
 ---> 0133c2ed9652
Successfully built 0133c2ed9652
Successfully tagged rootfsimage:latest
be3471f2002c958f41719ebedeff0964af816aa630d6198bda39dc49aee0ecd7
<redacted_registry>/absukl/secrets-plugin:latest
9d58c0c98e352cc391b027fb98442608a14adafd5efaf5aadfbce1f9dfe484ca
^Azvim11zrsrs9hlsqjck5wsrzb6
474a8n26iufukulr6l2ng3387
overall progress: 1 out of 1 tasks
1/1: running   [==================================================>]
verify: Service converged
<redacted_registry>/absukl/secrets-plugin:latest
tmu2v6hlekhafdys1z6md8ppi
kax8qazcwxxhqxkkcv1qrw90j
uem4hlemktb3ced6j9tucd3yb
Key              Value
---              -----
created_time     2018-08-25T09:22:39.720965228Z
deletion_time    n/a
destroyed        false
version          1
snitch.1.61i61bunxx57@lx64pc0265    | /run/secrets/VAULT_TOKEN_SORT_OF:
snitch.1.61i61bunxx57@lx64pc0265    | 011ee7c6-cb66-cb13-ca36-67b596da68c1
snitch.1.61i61bunxx57@lx64pc0265    | /run/secrets/haha:
snitch.1.61i61bunxx57@lx64pc0265    | sosecret
snitch.1.vnmdiur4z9yb@lx64pc0265    | /run/secrets/VAULT_TOKEN_SORT_OF:
snitch.1.vnmdiur4z9yb@lx64pc0265    | 2b1001c9-977f-8361-b74d-295bddccd7ab
snitch.1.vnmdiur4z9yb@lx64pc0265    | /run/secrets/haha:
snitch.1.vnmdiur4z9yb@lx64pc0265    | sosecret
```