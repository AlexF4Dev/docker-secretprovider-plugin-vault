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
sha256:3e31eb2085e5150cfd9ccfe2d07baf06f1c82416885db72f01dde4ed0e6f1b09
f6c875836213db3abd5efabb56b6fbf682220b97882e9cb3d47f30803959bead
docker.io/absukl/secrets-plugin:latest
0f63f8c9acc40c37d2f7c63402d03888e7716278bf26721e5321b0eacf1ed0bf
Success! Uploaded policy: snitch
Key              Value
---              -----
created_time     2018-08-30T01:56:21.028242253Z
deletion_time    n/a
destroyed        false
version          1
Key              Value
---              -----
created_time     2018-08-30T01:56:21.138722437Z
deletion_time    n/a
destroyed        false
version          1
zjjwiauez2t48sb6ipa33hmyc
6g9wmnqjx7stqxevkczwkkunl
overall progress: 1 out of 1 tasks
ij2r01ffy6ak: running   [==================================================>]
verify: Service converged
docker.io/absukl/secrets-plugin:latest
wnxa45lpqzh7mh2ax0oqvytjo
s9os7zjabkcibzewmz88iiifc
sn313aqwy27ua9libmeqs3mon
ln04aag56gne0ojltcol5dezn
snitch.2.m4y7bxrxth9j@redacted_host    | secret:              this_was_not_wrapped
snitch.2.m4y7bxrxth9j@redacted_host    | wrapped_secret:      0a318705-45b6-3dcb-364f-cf50fca0ce02
snitch.2.m4y7bxrxth9j@redacted_host    | unwrapped_secret:    this_was_once_wrapped
snitch.2.m4y7bxrxth9j@redacted_host    | generic_vault_token: ca57c187-c2ff-c04f-146c-66af0a1336ce
snitch.1.m4qj1siyv6u1@redacted_host    | secret:              this_was_not_wrapped
snitch.1.m4qj1siyv6u1@redacted_host    | wrapped_secret:      f1a79064-6da9-185d-1451-61db703c8934
snitch.1.m4qj1siyv6u1@redacted_host    | unwrapped_secret:    this_was_once_wrapped
snitch.1.m4qj1siyv6u1@redacted_host    | generic_vault_token: f22f7fa5-038b-ad97-bc7a-7fe2ffa1731c
snitch.1.vwfhg080geu0@redacted_host    | secret:              this_was_not_wrapped
snitch.1.vwfhg080geu0@redacted_host    | wrapped_secret:      45fec9dc-cdcb-2d89-f9e0-405956687d7a
snitch.1.vwfhg080geu0@redacted_host    | unwrapped_secret:    this_was_once_wrapped
snitch.1.vwfhg080geu0@redacted_host    | generic_vault_token: 2f5508eb-c001-4f2b-a3c9-b30e37a67a3c
snitch.2.jinfngfkl15a@redacted_host    | secret:              this_was_not_wrapped
snitch.2.jinfngfkl15a@redacted_host    | wrapped_secret:      219d0378-1cfd-7767-587e-c85654dd3a3b
snitch.2.jinfngfkl15a@redacted_host    | unwrapped_secret:    this_was_once_wrapped
snitch.2.jinfngfkl15a@redacted_host    | generic_vault_token: 377b501a-93a9-770e-772d-960fbe96013b
...
```