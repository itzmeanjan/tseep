FROM golang:alpine as builder
RUN apk add --no-cache gcc musl-dev linux-headers git
WORKDIR /app
COPY . .
RUN cd v1/server; go build -o ../../server; cd ../..

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server /usr/bin
ENV ADDR="0.0.0.0"
ENV PORT=7000
EXPOSE ${PORT}
ENTRYPOINT [ "server" ]
