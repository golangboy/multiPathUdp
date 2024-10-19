# client
```bash
cd client
export MIDDLE_SERVERS=middle1:6666 middle2:6666
docker-compose up -d
```
> socks5 will listen at port 17170,then the traffic will forward to server via middle1 and middle2


# server
```bash
cd server
docker-compose up -d
```
> server will listen at port 18443

# middle
```bash
cd middle
export TARGET_HOST=server:18443
docker-compose up -d
```
> udp traffic from udp/port 6666 will forward to server:18443


# summary
Therefore, TCP traffic is forwarded to the server through one or more middle servers. Even if one or more middle servers go down, it does not affect the TCP connection established from the client. This is because the TCP traffic has already been converted to UDP traffic through the KCP protocol and reaches the server via multiple middle paths.