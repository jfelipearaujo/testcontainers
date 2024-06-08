#!/bin/sh

echo "Initializing SNS topics..."

awslocal sns create-topic \
    --name TestTopic