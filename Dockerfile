FROM alpine
RUN apk add libc6-compat curl
COPY ./shitcoin /bin/