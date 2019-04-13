## Blober.io
A simple file storage http server built on-top Minio 

### About
Blober is a simple http server that simplify application file management. Apps can make simple http
API calls to upload and download files. Think of it as a simpler Filestack.io. It uses Minio object storage under the hood,
the aim is to have a simple file management server deployed as a service in a system

### Benefits
* Make development easier and faster by not having to worry about file downloads and storage
* Fits in perfectly for microservices architecture
* Simple HTTP API

## Installation

* Download/Clone project
`go get github.com/adigunhammedolalekan/blober.io`

### Docker

`cd bloberio && docker-compose up`

### Kubernetes

`kubectl deploy -f stack.yml`

## Usage

### Concept
The first thing you need to do is to create an account
`curl localhost:9008/api/account/new -B {"email" : "adigunhammed@gmail.com", "password" : "yourPassword"}`

