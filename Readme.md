
# description

login to switch, execute ping command and return result

# flowchart

```mermaid
flowchart LR
    switch --> ip_device

```

# build for linux
set GOOS=linux

set GOARCH=amd64

go build sshping_exporter

# params
* config.file
* target.file
* web.listen-address

# test
curl http://localhost:9966/sshping?target=10.111.222.11:22

# config

target ip on query param match DSW IP on device yaml file

