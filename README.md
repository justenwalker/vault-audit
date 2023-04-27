# vault-audit

Sidecar for providing an audit device for HashiCorp Vault

It ingests vault audit log events and emits them to a file it manages and rotates, or to stdout.

The vault audit device can listen on:
- TCP (`-addr tcp://0.0.0.0:5555`) 
- UDP (`-addr udp://0.0.0.0:5555`)
- Unix Domain Socket (`-addr unix:///path/to/socket`)

The resulting logs can be directed to:
- File (`-out file:///var/log/vault/audit.log`)
- Stdout (`-out -`)

## Testing

Prerequisites:
- Go 1.20
- Docker with Docker Compose

Run `start-vault.sh` to start a persistent vault server. It will set up vault, an agent, and vault-audit.
After it comes up, you can tail the output of the `vault-audit` container to see all the vault audit logs.

To tear everything down, run `destroy-vault.sh`

## Use Cases

### Redirect to rotated file

Vault does not rotate its audit.log file. Without some rotation scheme, it will
eventually fill up the disk. It's possible to set up logrotate to rotate these logs
and send SIGHUP to vault; but its cumbersome to set up something like that inside a Kubernetes Pod.

```sh
# Run audit device on unix socket redirecting to /var/log/vault/audit.log
vault-audit -addr unix:///var/log/vault/audit.sock -out file:///var/log/vault/audit.log
# Enable vault audit on the socket
vault audit enable socket address=/var/log/vault/audit.sock socket_type=unix
```

Now all of the log events that go to the socket will come out `file:///var/log/vault/audit.log`
They will also be rotated automatically as they stream out, without share process namespace to SIGHUP vault or 
configure logrotate.

### Redirect to stdout

Redirecting the vault audit socket to stdout is useful if you are running inside of a kubernetes pod and 
you want your audit logs to be collected directly from the container logs, without mixing them in with
the vault server logs.

Instead, you can run the `vaut-audit` device as a side-car and enable vault audit socket device.
This way you can take advantage of kubernetes built-in stdout/stderr log ingestion and the various
observability tools that will ship these logs automatically.

```sh
# Run audit device on unix socket redirecting to stdout
vault-audit -addr unix:///var/log/vault/audit.sock -out -
# Enable vault audit on the socket
vault audit enable socket address=/var/log/vault/audit.sock socket_type=unix
```

## Docker Containers

Pre-built docker containers are available:

- `justenwalker/vault-audit:0.0.1` : multi-arch image for `linux/amd64v1` and `linux/arm64v8`
- `justenwalker/vault-audit:0.0.1-amd64` : image for `linux/amd64v1`
- `justenwalker/vault-audit:0.0.1-arm64` : image for `linux/arm64v8`
