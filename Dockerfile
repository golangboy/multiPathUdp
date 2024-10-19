FROM golang as builder
WORKDIR /app
COPY . .
WORKDIR /app/cmd/client
RUN export CGO_ENABLED=0 && go build -o app .
WORKDIR /app/cmd/server
RUN export CGO_ENABLED=0 && go build -o app .
WORKDIR /app
RUN chmod +x docker-entrypoint.sh

FROM alpine
COPY --from=builder /app /app
EXPOSE 8888/udp 9999/udp
CMD ["/app/docker-entrypoint.sh"]