FROM alpine:3.9

# install Pairing Based Cryptography library (PBC)
# we MUST use a compound command (i.e., multiple commands chained with &&)
# because we can only remove packages in the same layer.
# See: https://github.com/gliderlabs/docker-alpine/issues/45
RUN apk add --no-cache gmp-dev build-base flex bison && \
    wget https://crypto.stanford.edu/pbc/files/pbc-0.5.14.tar.gz && \
        tar xvzf pbc-0.5.14.tar.gz && \
        cd pbc-0.5.14 && \
        ./configure && \
        make && \
        make install && \
        make clean && \
        cd .. && \
        rm pbc-0.5.14.tar.gz && \
        rm -rf pbc-0.5.14 && \
    apk del build-base flex bison