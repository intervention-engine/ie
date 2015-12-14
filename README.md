Intervention Engine
===================
[![Build Status](https://travis-ci.org/intervention-engine/ie.svg?branch=master)](https://travis-ci.org/intervention-engine/ie)
The goal of this project is to allow providers to easily interact with their own clinical data to enable the creation of “homegrown” clinical quality measures. After providers have created their own measures, they will be able to use Intervention Engine to track their performance over time and make adjustments to their measures.

This repository contains the intervention-engine-specific functionality that makes use of the [generic FHIR server](http://github.com/intervention-engine/fhir). Specifically, it contains middleware handlers for fact creation and management to support query execution in MongoDB.

Environment
-----------

This project currently uses Go 1.3.3 and is built using the Go toolchain.

To install Go, follow the instructions found at the [Go Website](http://golang.org/doc/install).

Following standard Go practices, you should clone this project to:

    $GOPATH/src/github.com/intervention-engine/ie

To get all of the dependencies for this project, run:

    go get

In this directory.

This project also requires MongoDB 2.6.* or higher. To install MongoDB, refer to the [MongoDB installation guide](http://docs.mongodb.org/manual/installation/).

To start the server, simply run server.go:

    go run server.go

Creating a User
---------------

To use the web application, you must register a user account using the `ie-user` tool.  For more info see the [intervention-engine/tools](https://github.com/intervention-engine/tools) repo.

License
-------

Copyright 2014 The MITRE Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
