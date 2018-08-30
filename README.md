# Sample Docker/Vault secrets plugin

## How it works / requirements

This plugin works by calling out to Vault whenever the value of a secret needs to be provided to a task (container). In order to let the plugin authenticate itself with Vault, it needs its own secret credentials. This plugin handles that by execing a command inside a helper service's container, which reveals a regular Swarm secret that contains the Vault token that the plugin should use.

Currently only the opaque token authentication method is supported, but support for the AppRole method would be a welcome contribution.

In order for the plugin to work, the helper service must first be created with a secret attached that gives the right permissions inside Vault.

## TODO:

- [ ] Docs
- [ ] Make TLS configurable
- [x] CI/CD
- [ ] Use docker-compose for teardown/bootstrap

## Preliminary instructions:

Run `./rebuild.sh`, you should get the following

(Note the different tokens in the two task instances of the `snitch` service).

```console
$ ./rebuild.sh
...
sha256:2698a4d10452b8de062c5506a7e8a7af1c189491c0bcecc5db4db7a6e0f6e848
6a1349700598b8d087bb0ab85afb8c8fb75fafad08a7e3a874844d073f07a5c8
sirlatrom/docker-secretprovider-plugin-vault
21c414520a21939b90013234f16cb24eeafc5fb404898b56e0cf1a3090fb082e
Success! Uploaded policy: snitch
Key              Value
---              -----
created_time     2018-08-30T02:21:27.980476389Z
deletion_time    n/a
destroyed        false
version          1
Key              Value
---              -----
created_time     2018-08-30T02:21:28.088984314Z
deletion_time    n/a
destroyed        false
version          1
fiqw1xaqjqofvinflvmnzo83t
zj2hvlev230x0s1ei9t25ft9m
overall progress: 1 out of 1 tasks
ij2r01ffy6ak: running   [==================================================>]
verify: Service converged
sirlatrom/docker-secretprovider-plugin-vault
5ioiauam5n9nms9neb6szbwxj
n5j855gu0i460bno1aaw3neq9
t99bbs7y5c1y61tyxxhd3msoj
i9sbmcaqc46v5a44u00fkfxfv
snitch.2.y42sqzj8524y@redacted_host    | secret:              this_was_not_wrapped
snitch.2.y42sqzj8524y@redacted_host    | wrapped_secret:      1afd51f9-c1a2-d4ec-8ceb-8e043b77b53a
snitch.1.gpy8rj3oxz0n@redacted_host    | secret:              this_was_not_wrapped
snitch.1.gpy8rj3oxz0n@redacted_host    | wrapped_secret:      6567b96c-338e-cd3b-e9bc-67c65597fd0f
snitch.2.y42sqzj8524y@redacted_host    | unwrapped_secret:    this_was_once_wrapped
snitch.2.y42sqzj8524y@redacted_host    | generic_vault_token: ddef57f5-a235-923c-4e7c-0a519d307f10
snitch.1.gpy8rj3oxz0n@redacted_host    | unwrapped_secret:    this_was_once_wrapped
snitch.1.gpy8rj3oxz0n@redacted_host    | generic_vault_token: b7b27691-1776-ae52-ffc3-b6a59152d12f
snitch.2.lepelzpcjscj@redacted_host    | secret:              this_was_not_wrapped
snitch.2.lepelzpcjscj@redacted_host    | wrapped_secret:      84df01da-11f9-acba-0373-89bd1f161798
snitch.2.lepelzpcjscj@redacted_host    | unwrapped_secret:    this_was_once_wrapped
snitch.2.lepelzpcjscj@redacted_host    | generic_vault_token: 9214b53f-027c-6552-a0a9-1b18783550d1
snitch.1.edwdvkmvpbke@redacted_host    | secret:              this_was_not_wrapped
snitch.1.edwdvkmvpbke@redacted_host    | wrapped_secret:      df1714e1-a1b6-07eb-10c1-7e1ba4e73022
snitch.1.edwdvkmvpbke@redacted_host    | unwrapped_secret:    this_was_once_wrapped
snitch.1.edwdvkmvpbke@redacted_host    | generic_vault_token: bedf29d5-fbdd-6085-7809-f113078c66b1
...
```