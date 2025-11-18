### Command used for building and running containers

#### Build image (move into root of app before launching the cmd)

`docker build -t <image_name> .`

#### Run container using local network, only use locally

`docker run -d --network=host <image_name>`

#### List running containers

`docker ps`
