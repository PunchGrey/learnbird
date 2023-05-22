# builder image
#FROM golang:1.19-alpine as builder
FROM golang:1.19-alpine as builder

ENV CGO_ENABLED 0
ENV GO111MODULE on
RUN apk --no-cache add git
WORKDIR /go/src/learnbird
COPY . .
ENV GOARCH amd64
RUN go build -o /bin/learnbird -v

# final image
FROM alpine:3.14.6
MAINTAINER PunchGrey

ENV LB_MONGO_URL "set mongo url"
ENV LB_MONGO_PASSWORD "set mongo password"
ENV LB_MONGO_USER "set mongo user"
ENV LB_DEPTH  "set depth"
ENV LB_TELEGRAM_APITOKEN "set telegram api token"
RUN apk --no-cache add ca-certificates dumb-init tzdata
COPY --from=builder /bin/learnbird /bin/learnbird

USER 65534
ENTRYPOINT ["dumb-init", "--", "/bin/learnbird"]
