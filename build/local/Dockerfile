# global arg
ARG WORKDIR=/app
ARG APP_ENV=local

FROM golang:1.21-alpine as build

# install tools
RUN apk update && apk upgrade && apk add git openssh

# arg
ARG WORKDIR

# env
ENV GOPRIVATE=github.com/SeanZhenggg CGO_ENABLED=0 GO111MODULE=on GONOPROXY=github.com/SeanZhenggg GONOSUMDB=github.com/SeanZhenggg

WORKDIR $WORKDIR

COPY go.mod go.sum ./

RUN --mount=type=secret,id=pat_token \
    export TOKEN=$(cat /run/secrets/pat_token) && \
    git config --global url."https://${TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"

RUN go mod download

COPY . .

RUN go build -buildvcs=false -o ./main ./cmd/web

# minimize the docker image size by using multistage build
FROM alpine:3.14 as final

# ARG
ARG WORKDIR
ARG APP_ENV

# env
ENV APP_ENV=$APP_ENV

COPY --from=build $WORKDIR/main /

COPY --from=build $WORKDIR/configs/$APP_ENV $WORKDIR/configs/$APP_ENV

ENTRYPOINT ["./main"]

