#!/bin/bash
cd $HOME
FileName='go1.13.linux-armv6l.tar.gz'
wget https://dl.google.com/go/$FileName
sudo tar -C /usr/local -xvf $FileName
rm $FileName
cat >> ~/.bashrc << EOF
export GOPATH=$HOME/go
export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin
EOF
source ~/.bashrc