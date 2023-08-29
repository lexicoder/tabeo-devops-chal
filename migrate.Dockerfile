FROM golang:1.21-bullseye as base

WORKDIR $GOPATH/src/migrate/

RUN go install github.com/jackc/tern@v1.13.0
RUN cp $GOPATH/bin/tern /tern

FROM gcr.io/distroless/base-debian11

COPY --from=base /tern .
COPY /migrations /migrations

CMD ["./tern", "migrate", "--migrations", "/migrations"]
