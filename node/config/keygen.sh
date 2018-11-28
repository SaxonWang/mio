#!/usr/bin/env bash

mkdir -p ssl
openssl genrsa -out ssl/app.rsa 1024
openssl rsa -in ssl/app.rsa -pubout > ssl/app.rsa.pub