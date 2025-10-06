FROM golang:1.25 AS builder
WORKDIR /workspace
COPY . .
RUN go build ./...

FROM gcr.io/distroless/base-debian12
WORKDIR /workspace
COPY --from=builder /workspace /workspace
CMD ["/bin/sh", "-c", "echo gostratum/core library image"]
