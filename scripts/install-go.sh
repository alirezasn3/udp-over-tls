#!/bin/bash

cd
curl -OL https://golang.org/dl/go1.20.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xvf go1.20.3.linux-amd64.tar.gz
rm go1.20.3.linux-amd64.tar.gz
sleep 1
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile
sleep 1
sudo source ~/.profile
go version