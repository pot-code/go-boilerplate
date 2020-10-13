FROM golang:1.14 as builder

WORKDIR /go/src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
# the last 'main' is the package name
ARG SKAFFOLD_GO_GCFLAGS
RUN GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -gcflags="${SKAFFOLD_GO_GCFLAGS}" -o app ./cmd

FROM alpine:3.8
ENV GOTRACEBACK=single

WORKDIR /root
COPY --from=builder /go/src/app .
RUN chmod 0755 app

EXPOSE 8081

CMD [ "./app" ]
