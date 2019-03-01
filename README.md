# ChuRP: Dynamic-Committee Proactive Secret Sharing

[![CircleCI](https://circleci.com/gh/bl4ck5un/ChuRP.svg?style=svg&circle-token=34c3da94eba4225de1da5c4eaabd37466cd50a8a)](https://circleci.com/gh/bl4ck5un/ChuRP)

## Getting Started

These intructions will get you through the organization of protocol codes and prepare you to run the protocol on you local machine for development and testing purposes. Automatic scripts for AWS deployment will come along in the near future. The codes are currently implemented using Go Wrappers of [GNU Multi Precision library](https://github.com/ncw/gmp), [Pairing Based Cryptography library](https://github.com/Nik-U/pbc) and [Google Protobuffer](https://github.com/golang/protobuf). You need to install all dependencies listed below. The recommended way to do this is to use the docker image we provided.

### Run with Docker

#### Install Docker

We strongly recommend you to run the codes with the docker image
we provide which already contains all the dependencies. To run wi
th docker, you first need to install docker on your machine. Refe
r to the [docker document](https://docs.docker.com/install/#supported-platforms) for installation instructions.

#### Download Docker Image

After you successfully install docker on your machine, run the following command to download our docker image.

`docker pull churp/mpss:latest`

#### Start an Interactive Container

To start a container with the image, run the following command.

`docker run -ti churp/mpss:latest`

#### Run on Local Machine

To test the codes on your local machine, run the following command.

~~~
cd src/localtest
./reset.sh
./simple.sh 3 1
~~~

## Authors
Lun Wang, Fan Zhang, Andrew Low

## Contacts
Lun Wang, wanglun@berkeley.edu
