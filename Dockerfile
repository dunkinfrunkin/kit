FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/kitd ./cmd/kitd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/kitd /bin/kitd
EXPOSE 80
ENV PORT=80
ENTRYPOINT ["/bin/kitd"]
