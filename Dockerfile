FROM golang:1.23.6 AS builder

MAINTAINER MultiversX

WORKDIR /multiversx
COPY . .

WORKDIR /multiversx/cmd/multi-factor-auth

RUN go build -o tcs

# ===== SECOND STAGE ======
FROM ubuntu:20.04
COPY --from=builder /multiversx/cmd/multi-factor-auth /multiversx

EXPOSE 8080

WORKDIR /multiversx

ENTRYPOINT ["./tcs"]
CMD ["--log-level", "*:DEBUG", "--start-swagger-ui"]
