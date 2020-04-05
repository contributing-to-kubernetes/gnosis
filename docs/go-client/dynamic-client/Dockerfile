From golang:1.12 AS base

ENV GO111MODULE=on
WORKDIR /go/src/app
ADD go.mod go.sum /go/src/app/
RUN go mod download
ADD main.go /go/src/app/
# CGO_ENABLED=0 disables cgo - no cross-compiled dependencies.
# -a forces a rebuild.
# -ldflags -w disables debug.
# -extldflags "-static" 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go \
    build -a -ldflags='-w -extldflags "-static"' \
    -o /go/bin/app


FROM scratch
COPY --from=base /go/bin/app /go/bin/app
ENTRYPOINT ["/go/bin/app"]
