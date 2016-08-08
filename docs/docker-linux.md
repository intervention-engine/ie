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

In addition, you must configure the multifactorriskservice to correctly point to your REDCap installation and contain your REDCap API token.  To do this, edit the values after `-redcap` and `-token` in the `command` entry of the `multifactorriskservice`:

```
  command: /go/src/github.com/intervention-engine/multifactorriskservice/multifactorriskservice -redcap https://your_redcap_server/redcap/api -token your_redcap_api_token -fhir http://ie:3001
```

Create or Configure SSL Certificates and Keys
=============================================
In order for nginx to use secure http (`https`), it requires an ssl certificate and key. These can be generated (self-signed certificate) or obtained from a Certificate Authority. To generate a certificate and key, run the following command:

```
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout nginx.key -out nginx.crt
```

You will then have `nginx.key` and `nginx.cert` files in your current directory. These must be moved into the Intervention Engine `nginx` repository in a folder named `ssl`. The `Dockerfile` in the `nginx` repository will then copy the key and certificate into the required location in the `nginx` container.

If you obtained your certificate and key from a Certificate Authority, simply rename them to `nginx.key` and `nginx.cert` and move or copy them into the `nginx/ssl` directory.

Add Users to System
===================
To add users to Intervention Engine, you must use the `htpasswd` utility to append users to the relevant file in the `nginx` repository. To install `apache2-utils` which contains the `htpasswd` tool, run the following command:

```
sudo apt-get install apache2-utils
```

Navigate to the `nginx` repository and run the following command, replacing `exampleuser` with the username you would like to add:

```
sudo htpasswd -c htpasswd exampleuser
```

The tool will prompt you for a password:

```
New password:
Re-type new password:
Adding password for user exampleuser
```

This will add an entry to the `htpasswd` file with the username and encrypted password.

Repeat this step for each user you would like to add to the system.

Run docker-compose to Build and Launch the Containers
=====================================================
Once your Docker Virtual Machine and docker-compose.yml file are set up and configured, you can use docker-compose to build and launch all of the Intervention Engine containers.

Navigate to the directory where the *ie* repository was cloned and run the following command:

```
docker-compose up
```

Docker will then begin downloading and building containers. Once the containers are built, docker-compose will report all stdout output from the running containers, prepended by the container name. Please note that docker-compose will then be running in the foreground of your terminal session, so you will need to open another terminal and initialize the docker environment variables with `$ eval $(docker-machine env default)` to interact with the docker containers.

Once all of the containers are built and running, you can connect to intervention engine over https (port 443). Your browser should prompt you for a username and password when you connect.


Updating Docker Containers
==========================

When there are updates made to the repositories that the docker containers are running, it is necessary to pull and redeploy the latest changes. I will use the `ie` container as an example.

To begin, run the following command to stop the docker containers:

```
docker-compose down
```

This should produce an output similar to the following:

```
$ docker-compose down
Stopping ie_nginx_1 ... done
Stopping ie_multifactorriskservice_1 ... done
Stopping ie_endpoint_1 ... done
Stopping ie_ie_1 ... done
Stopping ie_mongodb_1 ... done
Removing ie_nginx_1 ... done
Removing ie_multifactorriskservice_1 ... done
Removing ie_endpoint_1 ... done
Removing ie_ie_1 ... done
Removing ie_mongodb_1 ... done
```

Then, you will need to remove the images related to the container you would like to rebuild. Run the following command:

```
docker images
```

This should produce an output similar to the following:

```
$ docker images
REPOSITORY                  TAG                 IMAGE ID            CREATED             SIZE
ie_nginx                    latest              0a20c8c8a0c9        2 hours ago         205.9 MB
ie_multifactorriskservice   latest              fc7401615cc3        2 hours ago         829.3 MB
ie_endpoint                 latest              dfa7a9f4c4cd        2 hours ago         1.075 GB
ie_ie                       latest              1b2a6606ce81        2 hours ago         905.1 MB
mongo                       latest              87bde25ffc68        9 days ago          326.7 MB
golang                      latest              f24c8478ed40        10 days ago         744.2 MB
rails                       onbuild             eeae826cecfe        6 weeks ago         803.3 MB
nginx                       latest              0d409d33b27e        9 weeks ago         182.8 MB
```

You will then need to find the `IMAGE ID` of the container you would like to remove. In this case, we are removing `ie_ie`, so the `IMAGE ID` we need is `1b2a6606ce81`. Now run the following command:

```
docker rmi 1b2a6606ce81
```

Replacing `1b2a6606ce81` with the `IMAGE ID` that you noted.

This should produce output similar to the following:

```
$ docker rmi 1b2
Untagged: ie_ie:latest
Deleted: sha256:1b2a6606ce81f685aad0dac8eb92557b28d95ca8d32ccd6ba76dddfe1d41834b
Deleted: sha256:590c841652812c539ae4f7e8325f9ce45933821d12c4a37b99aad969f3766951
Deleted: sha256:27e52e52fa5463e722eb7e9c06aabfb959cb2496fac80719e355b011d037b437
Deleted: sha256:d5c94797e1fbc2fc73d706856761be34308dba2776fe6aa5cf06fbeca7bc84fe
Deleted: sha256:cb58792bebf694818a260cc5139d478857c43733d8b1fc3d386c62608c0913df
Deleted: sha256:068589e7f690d45c2c726c194ab17f0e40d290bbcd1931822cdc6b264b1e5853
Deleted: sha256:f5ccb380427a42fda34fdc084700614af139ff2e18a4722872bcedb72563ebb5
Deleted: sha256:3e42e6645b788b331904bc2e5e13247f835a899c00e1bdac6e79bc403634e76b
Deleted: sha256:ae63d4daa79004efba2d1c48b27553b6fa2fbaec8627144398f04f964cdefa60
Deleted: sha256:9ec97e6035245b3a8e9380c0c72d2b42146b06af56c336bb472d0f703ff6cdfd
Deleted: sha256:39533736e39f700c1224dbdfc69255b1e2356a9df693a868a119c898e2130b2e
```

This is docker reporting the deleted dependency images to the `ie_ie` image.

Now you will make whatever modifications you need to the `ie` repository. This will usually involve navigation to the `ie` directory and running `git pull`. Once your local repository is updated, you can start the service again by simply running:

```
docker-compose up
```
