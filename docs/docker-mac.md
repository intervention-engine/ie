Building and Running the Intervention Engine Stack in a Containerized Environment with Docker (Mac OS X)
========================================================================================================

This document details the steps needed to build and deploy the Intervention Engine stack in a containerized environment using [Docker](https://www.docker.com/). To deploy the Intervention Engine stack in a local development environment, please refer to [Building and Running the Intervention Engine Stack in a Development Environment](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md).

These instructions are written for deployment on the Mac OS X operating system. Some steps may vary for other operating systems. For instructions to deploy Intervention Engine with Docker on linux, please refer to the instructions found [here](https://github.com/intervention-engine/ie/blob/master/docs/docker-linux.md)

Install Prerequisite Tools
==========================

Building and running Intervention Engine with Docker requires the following 3rd party tools:

- [Docker Toolbox](https://www.docker.com/products/docker-toolbox)
- Git

Install Docker Toolbox
----------------------
Docker Toolbox is a set of tools including `docker`, `docker-compose`, and `docker-machine`, required for the creation and deployment of the Intervention Engine component containers.

To install the Docker Toolbox, follow the instructions found in the [Install Docker Toolbox on Mac OS X guide](https://docs.docker.com/mac/step_one/).

As an alternative to manual installation, many Mac OS X developers use [Homebrew](http://brew.sh/) to install common development tools. If you prefer to install the Docker Toolbox using Homebrew, execute the following commands:

```
$ brew update
$ brew cask install dockertoolbox
```

Install Git
-----------
Intervention Engine source code is hosted on GitHub. The Git toolchain is needed to clone Intervention Engine source code repositories locally. At the time this documentation was written, Git 2.7.0 was the latest available release.

To install Git, follow the instructions found in the [Git Book - Installing Git chapter](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

If you prefer to install the latest Git release using Homebrew, execute the following commands:

```
$ brew update
$ brew install git
```

Clone Intervention Engine GitHub Repositories Locally
=====================================================
To clone the Intervention Engine Repositories Locally, follow the instructions found [in our local development build and install instructions](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#clone-intervention-engine-github-repositories-locally).

The repositories that are required to deploy Intervention Engine in Docker are as follows:

-	ie: https://github.com/intervention-engine/ie
-	riskservice: https://github.com/intervention-engine/riskservice
-	nginx: https://github.com/intervention-engine/nginx
- ie-ccda-endpoint: https://github.com/intervention-engine/ie-ccda-endpoint

Cloning the ie-ccda-endpoint repository is not covered in our local setup instructions.

Clone ie-ccda-endpoint Repository
---------------------------------
The *ie-ccda-endpoint* repository contains the source code for a Ruby on Rails web endpoint that accepts POST requests containing a single Consolidated Clinical Document Architecture (CCDA) XML document, converts the CCDA document to a FHIR document, and uploads it to the Intervention Engine FHIR server.

Since this is not a Go project, it should not be cloned under the $GOPATH. Instead, we recommend you create an *intervention-engine* folder within your favorite development location and clone the *ie-ccda-endpoint* repository there. For this documentation, we'll assume that "your favorite development location" is `~/development`.

```
$ cd ~/development/intervention-engine
$ git clone https://github.com/intervention-engine/ie-ccda-endpoint.git
```


Set Up Docker Virtual Machine
================================
If you followed the official Docker Mac OS X setup guide completely, you will have already run the Docker Quickstart Terminal program. If you installed the Docker Toolbox using Homebrew, please continue with the instructions in this section.

Using Mac OS X Launchpad, find the Docker Quickstart Terminal program. It should prompt you to choose your favorite terminal to run in. This tool will provision a Virtual Machine and configure Docker to use that machine to host it's containers. You will know this operation was complete and successful when you see this in the terminal:

```
Docker is up and running!
To see how to connect your Docker Client to the Docker Engine running on this virtual machine, run: /usr/local/bin/docker-machine env default


                        ##         .
                  ## ## ##        ==
               ## ## ## ## ##    ===
           /"""""""""""""""""\___/ ===
      ~~~ {~~ ~~~~ ~~~ ~~~~ ~~~ ~ /  ===- ~~~
           \______ o           __/
             \    \         __/
              \____\_______/


docker is configured to use the default machine with IP 192.168.99.100
For help getting started, check out the docs at https://docs.docker.com
```

If you run the command that the success message provides (`/usr/local/bin/docker-machine env default`) you should see something like this:

```
$ /usr/local/bin/docker-machine env default
export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://192.168.99.100:2376"
export DOCKER_CERT_PATH="/Users/ahubley/.docker/machine/machines/default"
export DOCKER_MACHINE_NAME="default"
# Run this command to configure your shell:
# eval $(docker-machine env default)
```

You may notice that if you try to run any `docker` commands in a terminal session other than the one opened by Docker Quickstart Terminal, the shell will not be able to communicate with the Docker host. In order to configure your shell, run the command provided in the above message:

```
$ eval $(docker-machine env default)
```

You should then be able to run commands like `docker ps` to list running containers.

Configure docker-compose.yml
============================
Docker-compose is part of the Docker Toolbox which allows the orchestration of multiple containers, allowing the linking of containers and definition of which ports each container exposes and which ports those are mapped to on the Docker host. Docker-compose is configured in the appropriately named docker-compose.yml file. This file currently exists in the *ie* repository.

The top level of the YAML in the docker-compose.yml defines each container to be built. The second level tags of the YAML define specific configuration options for specific containers. The only field that needs to be configured before deployment is the `build` field. This field specifies the local directory in which to find the `Dockerfile` with which to build its parent container.

Since our instructions for cloning the Intervention Engine repositories specify to clone some repositories into the $GOPATH and some repositories to your preferred development location, you must set the `build` fields to point to each repository's local directory. For example, if you cloned the *ie-ccda-endpoint* repository to `~/development/intervention-engine/ie-ccda-endpoint`, the `endpoint` section of your docker-compose.yml should look as follows:

```
endpoint:
  build: ~/development/intervention-engine/ie-ccda-endpoint
  ports:
    - "3000:3000"
  links:
    - ie

```

You must configure *all* of the `build` fields to point to each repositories local directory.

Create or Configure SSL Certificates and Keys
=============================================
In order for nginx to use secure http (`https`), it requires an ssl certificate and key. These can be generated (self-signed certificate) or obtained from a Certificate Authority. To generate a certificate and key, run the following command:

```
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout nginx.key -out nginx.crt
```

You will then have `nginx.key` and `nginx.cert` files in your current directory. These must be moved into the Intervention Engine `nginx` repository in a folder named `ssl`. The `Dockerfile` in the `nginx` repository will then copy the key and certificate into the required location in the `nginx` container.

If you obtained your certificate and key from a Certificate Authority, simply rename them to `nginx.key` and `nginx.cert` and move or copy them into the `nginx/ssl` directory.

Run docker-compose to Build and Launch the Containers
=====================================================
Once your Docker Virtual Machine and docker-compose.yml file are set up and configured, you can use docker-compose to build and launch all of the Intervention Engine containers.

Navigate to the directory where the *ie* repository was cloned and run the following command:

```
docker-compose up
```

Docker will then begin downloading and building containers. Once the containers are built, docker-compose will report all stdout output from the running containers, prepended by the container name.

Once the containers are all up and running, you can then connect to the Intervention Engine Frontend, FHIR server, MongoDB database, and CCDA endpoint on the following ports:

- Intervention Engine Frontend - port 443 (https)
- FHIR server - port 3001
- MongoDB - port 27017
- CCDA endpoint - port 3000

Remember that these ports are exposed on the Docker Virtual Machine, NOT on your host machine. To navigate to the frontend in a browser, enter `http://{your docker virtual machine IP}:443`
