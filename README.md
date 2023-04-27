# udp over tls tunnel

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
  "role": "server",
  "connect": "1.2.3.4:443",
  "listen": "0.0.0.0:1194",
  "certificateLocation": "certs/client.pem",
  "keyLocation": "certs/client.key"
}
```