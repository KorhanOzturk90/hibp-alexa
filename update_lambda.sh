#!/usr/bin/env bash
set -e

GOFILE=${1:-"pwned"}
PROFILE=${2:-"default"}


GOOS=linux go build -o ${GOFILE}
zip pwned.zip ${GOFILE}
aws --profile ${PROFILE} lambda update-function-code --function-name pwned --zip-file fileb://pwned.zip --region eu-west-1