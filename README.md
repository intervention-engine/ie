Intervention Engine [![Build Status](https://travis-ci.org/intervention-engine/ie.svg?branch=master)](https://travis-ci.org/intervention-engine/ie)
===================================================================================================================================================

The Intervention Engine project provides a web-application for *data-driven team huddles*. Many care teams use team huddles to improve patient outcomes via efficient team communications and a holistic view of patients (due to the interdisciplinary nature of team huddles). Intervention Engine leverages electronic clinical records and clinical risk assessments to assist care teams in selecting patients for their huddles and providing the tools necessary to promote effective discussions and interventions.

Intervention Engine is a work in progress. Current Intervention Engine features:

-	custom population filters based on age, gender, conditions, and encounter types
-	clinical risk assessment integration via an open API
	-	prototype stroke risk calculation service (based on CHA2DS2-VASc)
	-	prototype "negative outcomes" risk calculation service (condition count + medication count)
-	patient views w/ summary data, risk trends, and risk component visualization
-	FHIR-based REST server
-	C-CDA import

Still to come:

-	Near term: huddle management (scheduling, viewing, progressing)
-	Near term: automated patient selection for huddles
-	Longer term: intervention planning & tracking
-	Longer term: population views and visualizations

The ie Repository
-----------------

The *ie* repository contains the source code for the backend Intervention Engine server. The *ie* server provides RESTful services needed by other components of the Intervention Engine stack. In addition to custom Intervention Engine features (such as authentication, notifications, and insta-count), it also doubles as a FHIR server (by integrating code from the [fhir](https://github.com/intervention-engine/fhir) repository).

Building and Running ie Locally
-------------------------------

Intervention Engine is a stack of tools and technologies. For information on installing and running the full stack, please see [Building and Running the Intervention Engine Stack in a Development Environment](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md).

For information related specifically to building and running the code in this repository (*ie*), please refer to the following sections in the above guide:

-	(Prerequisite) [Install Git](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-git)
-	(Prerequisite) [Install Go](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-go)
-	(Prerequisite) [Install MongoDB](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-mongodb)
-	(Prerequisite) [Run MongoDB](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#run-mongodb)
-	[Clone ie Repository](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#clone-ie-repository)
-	[Build and Run Intervention Engine Server](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#build-and-run-intervention-engine-server)
-	(Optional) [Create Intervention Engine User](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#create-intervention-engine-user)
-	(Optional) [Generate and Upload Synthetic Patient Data](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#generate-and-upload-synthetic-patient-data)

License
-------

Copyright 2016 The MITRE Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
