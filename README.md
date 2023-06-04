# udp over tls tunnel

This program listens on a UDP port on client side and forwards the traffic to the server on over a TLS connection. A new TLS connection to the server is created for each device that connects to the listening port of the client. The server will forward the received traffic over TLS to the specified port over UDP.

This diagram shows udp-over-tls and wireguard as an example:

`wireguardClient` <-> `udp-over-tls-client` <-> `udp-over-tls-server` <-> `wireguardServer`

The connection is initiated by `wireguardClient` in this example.

## How to use

* Create a `config.json` file and geneate certs using `gen-certs.sh` script. Use samples below to populate config file.

* Optionaly, you can enable bbr using the `enable-bbr.sh` script.

* Then build the project using `go build .` command.

* Add service file to systemctl using `add-to-systemd.sh` script. You can set path to config and certificate files by changing first argument passed to `ExecStart` field in the service file.
The default location for program and the config files is `/root/udp-over-tls/`.

## sample server config file
```json
{
  "role": "server",
  "connect": "0.0.0.0:1194",
  "listen": "0.0.0.0:443",
  "certificateLocation": "certs/server.pem",
  "keyLocation": "certs/server.key"
}
```

## sample client config file
```json
{
  "role": "client",
  "connect": "1.2.3.4:443",
  "listen": "0.0.0.0:1194",
  "certificateLocation": "certs/client.pem",
  "keyLocation": "certs/client.key"
}
```