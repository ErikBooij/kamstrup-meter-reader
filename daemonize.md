# How to daemonize the server on the remote machine

Take the following content and store it in /etc/systemd/system/meter-reader.service (root owned), then
run `sudo systemctl enable meter-reader`.

```
[Unit]
Description=Meter Reader Server
After=network-online.target
Wants=network-online.target systemd-networkd-wait-online.service

StartLimitIntervalSec=500
StartLimitBurst=5

[Service]
Restart=on-failure
RestartSec=5s

ExecStart=sudo /home/erikbooij/meter-reader-server --port /dev/ttyUSB0

[Install]
WantedBy=multi-user.target
```