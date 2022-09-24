FROM golang:1.19-bullseye as base

WORKDIR $GOPATH/src/write-hello/

COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /write-hello cmd/write-hello/main.go

FROM gcr.io/distroless/static-debian11

COPY --from=base /write-hello .

CMD ["./write-hello"]
