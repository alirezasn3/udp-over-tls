#!/bin/bash

sudo cp ../udp-over-tls.service /etc/systemd/system/udp-over-tls.service
sudo chmod 664 /etc/systemd/system/udp-over-tls.service
sudo systemctl daemon-reload
sudo systemctl start udp-over-tls
sudo systemctl enable udp-over-tls
sudo systemctl status udp-over-tls