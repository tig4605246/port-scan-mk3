FROM alpine:3.20
RUN apk add --no-cache busybox-extras
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
EXPOSE 8080
CMD ["/entrypoint.sh"]
