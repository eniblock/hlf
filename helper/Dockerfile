# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.17-alpine

WORKDIR /go/src/github.com/hyperledgendary/helper

COPY . .

RUN go install -v ./cmd/...
