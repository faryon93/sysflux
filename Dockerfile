# ----------------------------------------------------------------------------------------
# Image: Builder
# ----------------------------------------------------------------------------------------
FROM golang:alpine as builder

# setup the environment
ENV TZ=Europe/Berlin

# install dependencies
RUN apk --update --no-cache add git tzdata
WORKDIR /go/src/github.com/faryon93/sysflux
ADD ./ ./

# build the go binary
RUN go build -ldflags \
        '-X "main.BuildTime='$(date -Iminutes)'" \
         -X "main.GitCommit='$(git rev-parse --short HEAD)'" \
         -X "main.GitBranch='$(git rev-parse --abbrev-ref HEAD)'" \
         -X "main.BuildNumber='$CI_BUILDNR'" \
         -s -w' \
         -v -o /tmp/sysflux .

# ----------------------------------------------------------------------------------------
# Image: Deployment
# ----------------------------------------------------------------------------------------
FROM alpine:latest
MAINTAINER Maximilian Pachl <m@ximilian.info>

# setup the environment
ENV TZ=Europe/Berlin

RUN apk --update --no-cache add ca-certificates tzdata su-exec

# add relevant files to container
COPY --from=builder /tmp/sysflux /usr/sbin/sysflux

# make binary executable
RUN chmod +x /usr/sbin/sysflux

CMD /usr/sbin/sysflux
