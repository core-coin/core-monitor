FROM golang:1.16-alpine as base
RUN apk add --no-cache libstdc++ gcc g++ make git ca-certificates linux-headers
RUN git config --global url."https://error2215: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@github.com".insteadOf "https://github.com"
WORKDIR /go/src/github.com/core-coin/xcbexporter
ADD . .
RUN go get && go install

FROM alpine:latest
RUN apk add --no-cache jq ca-certificates linux-headers
COPY --from=base /go/bin/xcbexporter /usr/local/bin/xcbexporter

EXPOSE 9091
ENTRYPOINT xcbexporter
