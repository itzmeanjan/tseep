FROM golang:alpine as builder
RUN apk add --no-cache gcc musl-dev linux-headers git
WORKDIR /app
COPY . .
RUN cd client; go build -o ../client_app; cd ..

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/client_app /usr/bin
ENV ADDR="0.0.0.0"
ENV PORT=7000
ENV CLIENT_COUNT=8192
ENV DELAY=1s
ENV CONCURRENCY=256
ENTRYPOINT [ "client_app" ]
