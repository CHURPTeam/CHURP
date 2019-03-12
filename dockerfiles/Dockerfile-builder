FROM golang:1.12.0-alpine3.9

# install Pairing Based Cryptography library (PBC)
RUN apk add --no-cache gmp-dev build-base flex bison git bash

RUN wget https://crypto.stanford.edu/pbc/files/pbc-0.5.14.tar.gz && \
    tar xvzf pbc-0.5.14.tar.gz && \
    cd pbc-0.5.14 && \
    ./configure && \
    make && \
    make install && \
    make clean && \
    cd .. && \
    rm pbc-0.5.14.tar.gz && \
    rm -rf pbc-0.5.14