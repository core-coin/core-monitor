FROM golang:1.16-alpine as base
RUN apk add --no-cache libstdc++ gcc g++ make git ca-certificates linux-headers
RUN git config --global url."https://{login}:{token}@github.com".insteadOf "https://github.com"
WORKDIR /go/src/github.com/core-coin/core-monitor
ADD . .
RUN go get && go install

FROM alpine:latest
RUN apk add --no-cache jq ca-certificates linux-headers
COPY --from=base /go/bin/core-monitor /usr/local/bin/core-monitor

EXPOSE 9091
ENTRYPOINT core-monitor
