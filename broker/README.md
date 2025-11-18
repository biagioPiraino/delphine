### Setup local kafka broker

#### 1. Setup docker engine / desktop

**gnome terminal**

`sudo apt install gnome-terminal`

**docker apt repo**

```# Add Docker's official GPG key:
sudo apt update
sudo apt install ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
sudo tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc
EOF

sudo apt update
```

**docker packages**

`sudo apt install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin`

**verify**

`sudo docker run hello-world`

**downlaod deb package**

[link to package](https://desktop.docker.com/linux/main/amd64/docker-desktop-amd64.deb?utm_source=docker&utm_medium=webreferral&utm_campaign=docs-driven-download-linux-amd64)

**install the package**

```
sudo apt-get update
sudo apt-get install ./docker-desktop-amd64.deb
```

#### 2. Setup confluent cli

```
sudo apt update
sudo apt install confluent-cli -y
```

#### 3. Start the service and register topics

```
# launch local broker and list it
sudo confluent local kafka start

# launch to specific port to avoid random port allocation (useful with fixed port for containers)
sudo confluent local kafka start --plaintext-ports 9094 --brokers 1

# listing brokers
sudo confluent local kafka broker list

# create a topic and list it
sudo confluent local kafka topic create <topic_name>
sudo confluent local kafka topic list

# start reading messages arrived into broker
sudo confluent local kafka topic consume <topic_name> --from-beginning

# stop broker
sudo confluent local kafka stop
```
