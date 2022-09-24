# Dockerfile.distroless
FROM golang:1.19-bullseye as base

WORKDIR $GOPATH/src/booking-server/

COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /booking-server cmd/booking-server/main.go

FROM gcr.io/distroless/static-debian11

COPY --from=base /booking-server .

ENV SERVER_ADDRESS 0.0.0.0:5000

CMD ["./booking-server"]
