Building and Running the Intervention Engine Stack in a Development Environment
===============================================================================

Intervention Engine is a collection of tools and technologies. Intervention Engine developers will often want to build and run the entire stack from source code (excepting, of course, 3rd party dependencies). This document details the steps needed to accomplish that goal.

In short, the steps are as follows:

1.	[Install Prerequisite Tools and Servers](#install-prerequisite-tools-and-servers)
2.	[Clone Intervention Engine GitHub Repositories Locally](#clone-intervention-engine-github-repositories-locally)
3.	[Build and Run Intervention Engine Servers](#build-and-run-intervention-engine-servers)
4.	[Populate Intervention Engine Data](#populate-intervention-engine-data)
5.	[Test](#test)

These instructions are written for the Mac OS X operating system. Some steps may vary for other operating systems.

Install Prerequisite Tools and Servers
======================================

Building and running the Intervention Engine backend requires the following 3rd party tools and servers:

-	Go 1.6+
-	MongoDB 3.2+
-	Git

Building and running the Intervention Engine frontend (web UI server) additionally requires the following 3rd party tools:

-	Node.js 0.12+ / 5.0+
-	Bower
-	PhantomJS (testing only)

Install Go
----------

Intervention Engine's backend services and FHIR tools are written in Go. The Go tools are needed to install Intervention Engine's dependencies and compile Intervention Engine's code into binaries. Intervention Engine requires Go 1.6 or above. At the time this documentation was written, Go 1.6.2 was the latest available release.

To install Go, follow the instructions found in the [Go Programming Language Getting Started guide](http://golang.org/doc/install).

As an alternative to manual installation, many Mac OS X developers use [Homebrew](http://brew.sh/) to install common development tools. If you prefer to install the latest Go release using Homebrew, execute the following commands:

```
$ brew update
$ brew install go
```

Be sure to follow the advice in the [Go Programming Language Getting Started guide](http://golang.org/doc/install) regarding setting up environment variables (e.g., $GOROOT, $GOPATH) and your path.

Install MongoDB
---------------

Intervention Engine and its FHIR server store all data as BSON documents in MongoDB. Intervention Engine requires MongoDB 3.2 or above. At the time this documentation was written, MongoDB 3.2.6 was the latest available release.

To install the MongoDB community edition, follow the instructions found in the [MongoDB installation guide](https://docs.mongodb.org/manual/tutorial/install-mongodb-on-os-x/).

If you prefer to install the latest MongoDB release using Homebrew, execute the following commands:

```
$ brew update
$ brew install mongodb
```

Install Git
-----------

Intervention Engine source code is hosted on GitHub. The Git toolchain is needed to clone Intervention Engine source code repositories locally. At the time this documentation was written, Git 2.8.2 was the latest available release.

To install Git, follow the instructions found in the [Git Book - Installing Git chapter](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

If you prefer to install the latest Git release using Homebrew, execute the following commands:

```
$ brew update
$ brew install git
```

Install Node.js
---------------

The Intervention Engine frontend server uses Node.js tools for building, testing, and running in development. Due to the unification of the Node.js and io.js projects, Node.js version numbering is a little funky. Intervention Engine requires Node 0.12.z or 5.x. At the time this documentation was written, Node.js 0.12.13 was the latest available pre-unification release and Node.js 5.11.0 was the latest available post-unification 5.x release.  Intervention Engine has not been tested with Node.js 6.0 or higher, but is expected to work.

To install Node.js, download and execute the installer [here](https://nodejs.org/en/download/stable/) or install via a package manager, as described [here](https://nodejs.org/en/download/package-manager/#osx).

If you prefer to install the latest Node.js release using Homebrew, execute the following commands:

```
$ brew update
$ brew install node
```

Install Bower
-------------

The Intervention Engine frontend server uses Bower to manage its dependencies. At the time this documentation was written, Bower 1.7.9 was the latest available release.

To install Bower, use `npm` (which is installed with Node.js):

```
$ npm install -g bower
```

Install PhantomJS
-----------------

The Intervention Engine frontend server uses PhantomJS to simulate a browser for automated testing. At the time this documentation was written, PhantomJS 2.1.1 was the latest available release.

To install PhantomJS, follow the instructions found on the [PhantomJS Download page](http://phantomjs.org/download.html).

If you prefer to install the latest PhantomJS release using Homebrew, execute the following commands:

```
$ brew update
$ brew install phantomjs
```

Clone Intervention Engine GitHub Repositories Locally
=====================================================

Intervention Engine source code is hosted on Github. The following repositories need to be cloned to test and run a full instance of Intervention Engine:

-	ie: [https://github.com/intervention-engine/ie](https://github.com/intervention-engine/ie)
-	multifactorriskservice: [https://github.com/intervention-engine/multifactorriskservice](https://github.com/intervention-engine/multifactorriskservice)
-	tools: [https://github.com/intervention-engine/tools](https://github.com/intervention-engine/tools)
-	frontend: [https://github.com/intervention-engine/frontend](https://github.com/intervention-engine/frontend)

In addition to the above, the following repositories are also used by different aspects of Intervention Engine. Although they do not need to be cloned locally to run Intervention Engine, they should be cloned if you want to do any further development on Intervention Engine -- as they are important underlying components of Intervention Engine:

-	fhir: [https://github.com/intervention-engine/fhir](https://github.com/intervention-engine/fhir)
-	ptgen: [https://github.com/intervention-engine/ptgen](https://github.com/intervention-engine/ptgen)
-	hdsfhir: [https://github.com/intervention-engine/hdsfhir](https://github.com/intervention-engine/hdsfhir)
-	ember-fhir-adapter: [https://github.com/intervention-engine/ember-fhir-adapter](https://github.com/intervention-engine/ember-fhir-adapter)
-	fhir-golang-generator: [https://github.com/intervention-engine/fhir-golang-generator](https://github.com/intervention-engine/fhir-golang-generator)

Clone ie Repository
-------------------

The *ie* repository contains the source code for the backend Intervention Engine server. The *ie* server provides RESTful services needed by other components of the Intervention Engine stack. In addition to custom Intervention Engine features (such as notifications and insta-count), it also doubles as a FHIR server (by integrating code from the [fhir](https://github.com/intervention-engine/fhir) repository).

Following standard Go practices, you should clone the *ie* repository under your `$GOPATH` src folder, using a package-based sub-path:

```
$ mkdir -p $GOPATH/src/github.com/intervention-engine
$ cd $GOPATH/src/github.com/intervention-engine
$ git clone https://github.com/intervention-engine/ie.git
```

Clone multifactorriskservice Repository
---------------------------------------

The *multifactorriskservice* repository contains the source code for the prototype multi-factor risk service server.  The *multifactorriskservice* server interfaces with a [REDCap](http://projectredcap.org/) database to import recorded risk scores for patients, based on a multi-factor risk model.  The *multifactorriskservice* server also provides risk component data in a format that allows the Intervention Engine [frontend](https://github.com/intervention-engine/frontend) to properly draw the "risk pies".

The integration with REDCap supports our current use case, but users outside our organization don't likely have access to a REDCap server or the specific database referenced by the multi-factor risk service.  For this reason, the *multifactorriskservice* provides a *mock* implementation for generating synthetic risk scores to allow testing and development without a REDCap server.  *The mock implementation must ONLY be used for development with synthetic patients.  It should never be used with production (real) data!*

Following standard Go practices, you should clone the *riskservice* repository under your `$GOPATH` src folder, using a package-based sub-path:

```
$ cd $GOPATH/src/github.com/intervention-engine
$ git clone https://github.com/intervention-engine/multifactorriskservice.git
```

Clone tools Repository
----------------------

The *tools* repository contains command-line tools for generating and uploading synthetic patient data, uploading FHIR bundles, and converting and uploading Health Data Standards (HDS) records.

Following standard Go practices, you should clone the *tools* repository under your `$GOPATH` src folder, using a package-based sub-path:

```
$ cd $GOPATH/src/github.com/intervention-engine
$ git clone https://github.com/intervention-engine/tools.git
```

Clone frontend Repository
-------------------------

The *frontend* repository contains the source code for the Ember web application. This application communicates with the *ie* server and *multifactorriskservice* server to provide the Intervention Engine browser-based user interface.

Since this is not a Go project, it should not be cloned under the $GOPATH. Instead, we recommend you create an *intervention-engine* folder within your favorite development location and clone the *frontend* repository there. For this documentation, we'll assume that "your favorite development location" is `~/development`.

```
$ mkdir -p ~/development/intervention-engine
$ cd ~/development/intervention-engine
$ git clone https://github.com/intervention-engine/frontend.git
```

Clone fhir Repository
---------------------

The *fhir* repository contains the source code for the FHIR DSTU2 server. This server can be run standalone without the other *ie* services (if you want only a FHIR DSTU2 server). If you are only concerned with running Intervention Engine, you do not need to clone this repository (a version of it is already a vendored dependency of the *ie* project). If you wish to modify components of the FHIR server (for the standalone use case *or* the Intervention Engine use case), however, you should clone the *fhir* repository.

*NOTE: Most of the fhir source code is generated by the [fhir-golang-generator](https://github.com/intervention-engine/fhir-golang-generator). In most cases, updates to source code in the fhir repository need to be accompanied by corresponding updates in the fhir-golang-generator.*

Following standard Go practices, you should clone the *fhir* repository under your `$GOPATH` src folder, using a package-based sub-path:

```
$ cd $GOPATH/src/github.com/intervention-engine
$ git clone https://github.com/intervention-engine/fhir.git
```

Clone ptgen Repository
----------------------

The *ptgen* repository contains the source code for the synthetic patient generation library. If you are only concerned with generating patients for Intervention Engine, you do not need to clone this repository (it is already a vendored dependency of the *tools* project). If you wish to modify synthetic patient generation logic, however, you should clone the *ptgen* repository.

*NOTE: Due to Intervention Engine's prominent use case, all synthetic records are tuned to a geriatric population.*

Following standard Go practices, you should clone the *ptgen* repository under your `$GOPATH` src folder, using a package-based sub-path:

```
$ cd $GOPATH/src/github.com/intervention-engine
$ git clone https://github.com/intervention-engine/ptgen.git
```

Clone hdsfhir Repository
------------------------

The *hdsfhir* repository contains the source code for converting Health Data Standards (HDS) records to FHIR resources. If you are only concerned with running the conversion and uploading it to a FHIR server (or Intervention Engine), you do not need to clone this repository (it is already a vendored dependency of the *tools* project). If you wish to modify HDS-to-FHIR conversion logic, however, you should clone the *hdsfhir* repository.

*NOTE: The HDS-to-FHIR conversion focuses only on those data elements that are needed by Intervention Engine. It is not a complete and robust conversion.*

Following standard Go practices, you should clone the *hdsfhir* repository under your `$GOPATH` src folder, using a package-based sub-path:

```
$ cd $GOPATH/src/github.com/intervention-engine
$ git clone https://github.com/intervention-engine/hdsfhir.git
```

Clone ember-fhir-adapter Repository
-----------------------------------

The *ember-fhir-adapter* repository contains the source code for the Ember Data FHIR DSTU2 adapter. If you are only concerned with *running* Intervention Engine, you do not need to clone this repository (it will automatically be downloaded by `npm install` / `bower install` when you build the *frontend*). If you wish to modify the adapter logic, however, you should clone the *ember-fhir-adapter* repository.

*NOTE: Most of the ember-fhir-adapter source code is generated by the [fhir-golang-generator](https://github.com/intervention-engine/fhir-golang-generator). In most cases, updates to source code in the ember-fhir-adapter repository need to be accompanied by corresponding updates in the fhir-golang-generator.*

Since this is not a Go project, it should not be cloned under the $GOPATH. Instead, we recommend you create an *intervention-engine* folder within your favorite development location and clone the *ember-fhir-adapter* repository there. For this documentation, we'll assume that "your favorite development location" is `~/development`.

```
$ cd ~/development/intervention-engine
$ git clone https://github.com/intervention-engine/ember-fhir-adapter.git
```

Clone fhir-golang-generator Repository
--------------------------------------

The *fhir-golang-generator* repository is a fork of the HL7 FHIR DSTU2 source code, with additions and modifications to support the generation of FHIR code for Go and Ember. This repository is only needed if you intend to make changes to the code generation logic. In that case, the re-generated code should also be committed in the corresponding *fhir* or *ember-fhir-adapter* repository.

Since this is not a Go project, it should not be cloned under the $GOPATH. Instead, we recommend you create an *intervention-engine* folder within your favorite development location and clone the *fhir-golang-generator* repository there. For this documentation, we'll assume that "your favorite development location" is `~/development`.

```
$ cd ~/development/intervention-engine
$ git clone https://github.com/intervention-engine/fhir-golang-generator.git
```

Build and Run Intervention Engine Servers
=========================================

A fully running Intervention Engine stack consists of the following processes:

-	MongoDB database server (mongod)
-	Intervention Engine server (ie)
-	Multi-Factor Risk Service server (multifactorriskservice)
-	Frontend Ember server (ember)

Run MongoDB
-----------

In most cases, running MongoDB is as simple as executing the `mongod` command:

```
$ mongod
```

If you wish to fork the process (so it does not hang onto the shell), pass the `--fork` option:

```
$ mongod --fork
```

If you wish to specify configuration parameters, you can use a [configuration file](https://docs.mongodb.org/manual/reference/configuration-options/):

```
$ mongod --config /usr/local/etc/mongod.conf
```

Build and Run Intervention Engine Server
========================================

Before you can run the Intervention Engine server, you must build the `ie` executable:

```
$ cd $GOPATH/src/github.com/intervention-engine/ie
$ go build
```

The above commands do not need to be run again unless you make (or download) changes to the *ie* or *fhir* source code.

To support automatic huddle scheduling, you must pass the `ie` executable a `-huddle` argument to indicate the path to the huddle configuration file.  For more information of the huddle configuration file, see the [annotated huddle configuration file](https://github.com/intervention-engine/ie/blob/master/docs/huddle_config.md).

In addition, the first time you run the `ie` executable, you should also pass the `-loadCodes` option to load the ICD-9 and ICD-10 codes that are needed for the ICD-9/ICD-10 auto-complete feature:

```
$ ./ie -huddle ./config/multifactor_huddle_config.json -loadCodes
```

Automatic huddle scheduling will happen at the times indicated by the cron expression in the huddle configuration file.  You can also force huddles to be rescheduled by performing an HTTP GET on [http://localhost:3001/ScheduleHuddles](http://localhost:3001/ScheduleHuddles).

Subsequent runs of *ie* do not need to load the codes again:

```
$ ./ie -huddle ./config/multifactor_huddle_config.json
```

If you are concurrently modifying the *ie* source code, sometimes it is easier to combine the build and run steps into a single command (forcing a recompile on every run):

```
$ go run server.go -huddle ./config/multifactor_huddle_config.json
```

The *ie* server accepts connections on port 3001 by default.

Build and Run Multi-Factor Risk Service Server
==============================================

*NOTE: If you wish to run the MOCK Multi-Factor Risk Service server instead, please see the [multifactorriskservice README](https://github.com/intervention-engine/multifactorriskservice/blob/master/README.md).*

Before you can run the Multi-Factor Risk Service server, you must build the `multifactorriskservice` executable:

```
$ cd $GOPATH/src/github.com/intervention-engine/multifactorriskservice
$ go build
```

The above commands do not need to be run again unless you make (or download) changes to the *multifactorriskservice* source code.

The `multifactorriskservice` executable requires several arguments to indicate the URL to the REDCap API server (`-redcap`), the REDCap API token to use (`-token`), the URL to the FHIR API server (`-fhir`), and (optionally) a cron expression for when the data import should occur (`-cron`):

```
$ ./multifactorriskservice -redcap http://example.org/redcap/api -token abcdefg -fhir http://localhost:3001 -cron `0 0 22 * * *`
```

If no cron expression is passed in, it defaults to `0 0 22 * * *` (daily at 10:00pm).  For more information on supported cron expressions, see the [cron package documentation](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format).

If you are concurrently modifying the *multifactorriskservice* source code, sometimes it is easier to combine the build and run steps into a single command (forcing a recompile on every run):

```
$ go run main.go -redcap http://example.org/redcap/api -token abcdefg -fhir http://localhost:3001 -cron `0 0 22 * * *`
```

The *multifactorriskservice* server accepts connections on port 9000 by default.

Build and Run Frontend Ember Server
-----------------------------------

Before you can run the frontend Ember server, you must install and configure its dependencies. The following commands assume that the *frontend* repository is located at `~/development/intervention-engine/frontend`.

```
$ cd ~/development/intervention-engine
$ npm install
$ bower install
```

To run the frontend server, use the Ember CLI client (which was automatically installed as part of `npm install` above) and pass along the `--proxy` flag to indicate the URL of the *ie* server:

```
$ node_modules/.bin/ember s --proxy http://localhost:3001
```

Frequent npm users often define an `npm-exec` alias to allow them to more easily execute npm-installed local executables:

```
alias npm-exec='PATH=$(npm bin):$PATH'
```

With the `npm-exec` alias defined, you can run the frontend using the following command:

```
npm-exec ember s --proxy http://localhost:3001
```

The *frontend* server accepts connections on port 4200 by default.

Populate Intervention Engine Data (OPTIONAL)
============================================

Once the Intervention Engine servers are running, you'll likely want to populate the server with synthetic data in order to test it.

Generate and Upload Synthetic Patient Data (OPTIONAL)
-----------------------------------------------------

Generating synthetic patient data requires the *generate* command-line tool in the *tools* repository. Before you can run the *generate* tool, you must build the `generate` executable:

```
$ cd $GOPATH/src/github.com/intervention-engine/tools/cmd/generate
$ go build
```

The *generate* tool takes a `-fhirURL` flag to indicate the FHIR server to upload the patients to, as well as a `-n` flag to indicate the number of patients to generate (with the default being 100).

```
$ ./generate -fhirURL http://localhost:3001 -n 20
```

When you generate patients, you should see logging statements in the *ie* console indicating the posting of patient records.

Refresh Multifactor Risk Assessments (OPTIONAL)
-----------------------------------------------

With patient data now in the system, you may want to trigger the multifactor risk service to refresh its risk assessments.  If you're using the _mock_ multifactor risk service, this will generate fake risk assessments for every patient in the database.  If you are using the normal multifactor risk service, it will communicate with REDCap to update patient risk assessments.  To trigger a refresh (or generation) of the mock assessments, issue an HTTP POST to [http://localhost:9000/refresh](http://localhost:9000/refresh).

```
$ curl -X POST http://localhost:9000/refresh
```

Schedule Huddles (OPTIONAL)
---------------------------

With patient data and risk scores now in the system, you may want to trigger a rescheduling of the huddles.  To trigger huddles scheduling issue an HTTP GET to [http://localhost:3001/ScheduleHuddles](http://localhost:3001/ScheduleHuddles).

```
$ curl http://localhost:3001/ScheduleHuddles
```

Troubleshoot Slow Server Communications
=======================================

The Intervention Engine servers all interconnect via network protocols. In environments that use network proxies, sometimes better results are achieved when local client proxy handling is turned off in each shell that executes a server process (e.g., `unset http_proxy`). Keep in mind, however, that the `-loadCodes` option may need the proxy to reach Internet servers.

Test
====

Now that the Intervention Engine servers are running and data has been populated, it's time to try it out! Simply browse to the following URL:

http://localhost:4200
