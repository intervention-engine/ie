Building and Running the Intervention Engine Stack in a Containerized Environment with Docker (Ubuntu)
========================================================================================================

This document details the steps needed to build and deploy the Intervention Engine stack in a containerized environment using [Docker](https://www.docker.com/). To deploy the Intervention Engine stack in a local development environment, please refer to [Building and Running the Intervention Engine Stack in a Development Environment](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md).

These instructions are written for deployment on the linux (Specifically, [Ubuntu Server version 14.04.4](http://www.ubuntu.com/download/server)) operating system. Some steps may vary for other operating systems. For instructions to deploy Intervention Engine with Docker on Mac OS X, please refer to the instructions found [here](https://github.com/intervention-engine/ie/blob/master/docs/docker-mac.md)

Install Prerequisite Tools
==========================

Building and running Intervention Engine with Docker requires the following 3rd party tools:

- docker-engine
- docker-compose
- Git

Install Docker Engine
---------------------
[Docker Engine](https://www.docker.com/products/docker-engine) is the lightweight runtime that builds and runs your Docker containers, required for the creation and deployment of the Intervention Engine component containers.

To install the Docker Engine, follow the instructions found in the [Install Docker](https://docs.docker.com/linux/step_one/) documentation on the Docker website.

Install Docker Compose
----------------------

[Docker Compose](https://docs.docker.com/compose/overview/) is the Docker tool that orchestrates the linking of containers to allow them to communicate with one another.

To install Docker Compose, follow the instructions found in the [Install Docker Compose](https://docs.docker.com/compose/install/) documentation on the Docker website, starting with step 3.

Install Git
-----------
Intervention Engine source code is hosted on GitHub. The Git toolchain is needed to clone Intervention Engine source code repositories locally. At the time this documentation was written, Git 2.7.0 was the latest available release.

To install Git, follow the instructions found in the [Git Book - Installing Git chapter](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

If you prefer to install the latest Git release using apt-get, execute the following commands:

```
$ sudo apt-get update
$ sudo apt-get install git
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

Docker will then begin downloading and building containers. Once the containers are built, docker-compose will report all stdout output from the running containers, prepended by the container name. Please note that docker-compose will then be running in the foreground of your terminal session, so you will need to open another terminal and initialize the docker environment variables with `$ eval $(docker-machine env default)` to complete the following step of adding users to Intervention Engine.

Adding Users to Intervention Engine
===================================
Once the containers are all up and running, you can add Users to Intervention Engine login/authentication. To do so, run the following command for each User you would like to add:

```
docker exec ie_ie_1 bash -c '/go/src/github.com/intervention-engine/ie/deploy/ieuser add <username> <password> "$IE_MONGODB_1_PORT_27017_TCP_ADDR"'
```

Replacing `<username>` and `<password>` with the desired username and password. Keep in mind the username should be an email address, and the password should be at least 8 characters long.

Once you've registered your desired User accounts, you can then connect to the Intervention Engine Frontend, FHIR server, MongoDB database, and CCDA endpoint on the following ports:

- Intervention Engine Frontend - port 443 (https)
- FHIR server - port 3001
- MongoDB - port 27017
- CCDA endpoint - port 3000

These ports are exposed on your host machine. To navigate to the frontend in a browser, enter `http://{your host IP}:443`
