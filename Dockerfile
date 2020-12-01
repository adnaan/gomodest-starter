FROM golang:1.15.5-buster as build-go
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o app .

FROM node:14.3.0-stretch as build-node
WORKDIR /usr/src/app
COPY . .
RUN cd web && npm install
RUN cd web && npm run build

FROM alpine:latest
RUN apk add ca-certificates curl
WORKDIR /opt
COPY --from=build-go /go/src/app/ /bin
#COPY .env.production /opt
RUN chmod +x /bin/app
RUN mkdir -p web
COPY --from=build-node /usr/src/app/web/dist /opt/web/dist
COPY --from=build-node /usr/src/app/web/html /opt/web/html
#CMD ["app","-config", ".env.production"]
CMD ["/bin/app"]