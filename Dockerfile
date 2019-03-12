# docker build -t churp/churp .

FROM churp/builder as builder

COPY src /src
WORKDIR /src

RUN make

FROM churp/runtime

WORKDIR /root/

RUN apk add --no-cache bash
COPY localtest/* /root/
COPY --from=builder /src/*.exe /root/