FROM --platform=$BUILDPLATFORM alpine:3.21 AS certs
RUN apk add --no-cache ca-certificates && update-ca-certificates

FROM scratch
EXPOSE 8080
VOLUME /data
WORKDIR /data
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY dreaming-bard /
ENTRYPOINT [ "/dreaming-bard" ]
CMD [ "server" ]