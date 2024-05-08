FROM golang:latest AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/bashrun/bin/main ./cmd/bashrun/

FROM alpine:latest
WORKDIR /bashrun
RUN mkdir /bashrun/logs
COPY --from=build /build/cmd/bashrun/bin/main .
CMD ["/bashrun/main"]