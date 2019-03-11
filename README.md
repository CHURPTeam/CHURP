# CHURP: Dynamic-Committee Proactive Secret Sharing

[![CircleCI](https://circleci.com/gh/bl4ck5un/ChuRP.svg?style=svg&circle-token=34c3da94eba4225de1da5c4eaabd37466cd50a8a)](https://circleci.com/gh/bl4ck5un/ChuRP)

## Getting Started

These instructions will get you through the organization of the CHURP code and prepare you to run CHURP on your local machine for development and testing purposes. Automatic scripts for AWS deployment will come along in the near future. The code is currently implemented using Go Wrappers of [GNU Multi Precision library](https://github.com/ncw/gmp), [Pairing Based Cryptography library](https://github.com/Nik-U/pbc) and [Google Protobuffer](https://github.com/golang/protobuf). You need to install all dependencies listed below. The recommended way to do this is to use the docker image we provide.

### Run with Docker

#### Install Docker

We strongly recommend you to run the codes with the docker image we provide which already contains all the dependencies. To run with docker, you first need to install docker on your machine. Refer to the [docker document](https://docs.docker.com/install/#supported-platforms) for installation instructions.

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

## API

At a high level, CHURP provides the following API:

* `initialize(t, [nodeList], ...)`: Set the required parameters for CHURP: `t` stands for the threshold and `nodeList` represents the set of nodes that form a committee. Some other parameters that need to be set are the epoch duration and commitment scheme parameters.

* (Optional) `storeSecret(SK)`: Distribute the secret `SK` using [(t, n)-sharing](https://en.wikipedia.org/wiki/Shamir%27s_Secret_Sharing) `(n=|nodeList|)` such that each node in `nodeList` stores a share of the secret. (Note that this function is optional. For some applications, the secret might be generated randomly using [Distributed Key Generation](https://en.wikipedia.org/wiki/Distributed_key_generation) protocols.)

* `changeCommittee([newNodeList])`: Execute CHURP to handoff the secret `SK` from the old committee, `nodeList`, to the new committee, `newNodeList`.

* (Optional) `retrieveSecret() -> SK`: Reconstruct the secret from shares retrieved from nodes in the `nodeList`. (Note that this function is optional, i.e., CHURP works without any need to explicitly reconstruct the secret.)


## Authors
Lun Wang, Fan Zhang, Andrew Low

## Contacts
Lun Wang, wanglun@berkeley.edu
