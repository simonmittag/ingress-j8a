FROM golang:1.20.4-alpine AS build

RUN apk update && apk upgrade && apk add --no-cache bash git

WORKDIR .
COPY . /proj/

RUN /bin/bash
RUN cd /proj && CGO_ENABLED=0 go build github.com/simonmittag/ingress-j8a/cmd/ingress-j8a

#multistage build uses output from previous image
FROM alpine
COPY --from=build /proj/ingress-j8a /ingress-j8a

EXPOSE 80
EXPOSE 443
ENTRYPOINT "/ingress-j8a"