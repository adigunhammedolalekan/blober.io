## Blober.io
A simple file storage http API built on-top Minio

### About
Blober is a simple http server that simplify application file management. Your apps can make http
 calls to upload and download files. Think of it as a simpler Filestack.io. It uses Minio object storage under the hood.
The aim is to have a simple, easy to use file management service that can handle file of any type or size.

### Motivation
The easiest way to deploy Web apps is to use PaaS like **Heroku**, storing files on these PaaS platform is not possible because
of the ephemeral nature of their file systems. I wanted a simple http service that i can easily call to store files in a secure, most simplest way.
Other services offering these features are focused on image/video transformations which sometimes does not fit my need(I may just want to upload a video/PDF/Doc file and have a simple URL to download them later)

## Installation

### Step 1: Download
`git clone  https://github.com/adigunhammedolalekan/blober.io.git`

### Step 2: Run locally with docker

`cd blober.io && docker-compose up`

### Step 3: API Usage

#### Accounts and Credential
To get started with blober, you first need to create an account to have access to a set of credentials. These credentials are named
private and public key and they both have their usefulness. PrivateKey can be used to authorized critical actions like 
* creating an app
* fetch list of apps created
* download a private file


PublicKey are useful for uploading files directly from clients(Android, iOS, JS etc), this prevent
you from unnecessarily leaking your account's PrivateKey. PublicKey can only be used to upload files.


#### Apps
A single account can create many apps. Each apps can upload unlimited number of files. This is useful for separation
of concern.
You can deploy a single blober.io instance and create a new app for every one of your projects

## API Documentation

### Create an account
`curl -d '{"email": "mail@mail.com", "password":"passcode", "first_name" : "First Name"}' -H "Content-Type: application/json" -X POST http://localhost:9008/account/new`


* Response
```JSON
{
   "data" : {
      "email" : "user@mail.com",
      "first_name" : "First Name",
      "ID" : 3,
      "password" : "",
      "cred" : {
         "expires_in" : "2019-04-19T20:02:16.196840002Z",
         "private_access_key" : "privd320edfdc84af7c8fc673ac375074f0fc0fc3d8331ff3a9a2e45c566b9e5a81fbc8b88dc10aad16e",
         "public_access_key" : "publd320edfdc84af7c88dc925d3515f1b32d2d3f1bf4d62a9468c2be71e1e5a339787cb5e6318ed49c2"
      },
      "DeletedAt" : null,
      "CreatedAt" : "2019-04-14T15:02:16.29810483Z",
      "UpdatedAt" : "2019-04-14T15:02:16.29810483Z",
      "last_name" : ""
   },
   "error" : false,
   "message" : "account created"
}
```

### Authenticate an account (After credential has expired or no longer valid)
`curl -d '{"email": "user@mail.com", "password":"passcode"}' -H "Content-Type: application/json" -X POST http://localhost:9008/account/authenticate`


* Response
```JSON 
{
   "error" : false,
   "data" : {
      "password" : "",
      "email" : "user@mail.com",
      "UpdatedAt" : "2019-04-14T15:02:16.298105Z",
      "cred" : {
         "private_access_key" : "privd201bbe102bfc98f1002a86ea6294827e9da579b2ee0a60118d703992d33adafa537bf3d8c0f92c5",
         "public_access_key" : "publd201bbe102bfc98fcbde23329d648316a7c645cdd496306c3d7d3b4556d5e5ed77779b8e62dc10af",
         "expires_in" : "2019-04-19T20:04:34.313606676Z"
      },
      "last_name" : "",
      "DeletedAt" : null,
      "ID" : 3,
      "first_name" : "First Name",
      "CreatedAt" : "2019-04-14T15:02:16.298105Z"
   },
   "message" : "account authenticated"
}
```

### Create a new app

`curl -d '{"name": "FileShareApp"}' -H "Content-Type: application/json" -H "X-Blober-ID: privfc9d59ff8fba0d5388031758f0610f0e1bdd93661374c867aa34baecbb8216105f9eb5fa15c0d35a" -X POST http://localhost:9008/app/new | json_pp`

* create a new app response
```JSON
{
   "error" : false,
   "data" : {
      "name" : "FileShareApp",
      "CreatedAt" : "2019-04-14T15:15:36.730845911Z",
      "account_id" : 3,
      "DeletedAt" : null,
      "ID" : 14,
      "account" : {
         "last_name" : "",
         "cred" : null,
         "DeletedAt" : null,
         "first_name" : "First Name",
         "UpdatedAt" : "2019-04-14T15:02:16.298105Z",
         "password" : "$2a$10$XHRxCiPtsY4m8SBgnfS1pOMMOFSmZtuzzVbOWiI9CqmGv8kDv7p/u",
         "ID" : 3,
         "email" : "user@mail.com",
         "CreatedAt" : "2019-04-14T15:02:16.298105Z"
      },
      "UpdatedAt" : "2019-04-14T15:15:36.730845911Z"
   },
   "message" : "app created"
}

```

* Upload a file
`curl -F 'file_data=@/home/adigun/Downloads/food.jpeg' -H 'X-Blober-ID:privfc9d59ff8fba0d5388031758f0610f0e1bdd93661374c867aa34baecbb8216105f9eb5fa15c0d35a' -X POST http://localhost:9008/FileShareApp/upload | json_pp`

* Upload a file response
```JSON
{
   "message" : "success",
   "error" : false,
   "data" : {
      "filename" : "food.jpeg",
      "content_type" : "image/jpeg",
      "DeletedAt" : null,
      "hash" : "1907a5418dddc44bffa8be6cacd7cc50",
      "CreatedAt" : "2019-04-24T15:37:32.17130731Z",
      "app_id" : 15,
      "app_name" : "ChatApp",
      "UpdatedAt" : "2019-04-24T15:37:32.17130731Z",
      "is_private" : false,
      "download_url" : "http://blober.io/res/ChatApp/1907a5418dddc44bffa8be6cacd7cc50",
      "ID" : 10,
      "size" : 7421
   }
}

```

* Download a file
You can download a file using it `hash` and the `application name'
Downloading the file we upload previously, we can use the name of the app and the hash returned as part of the upload response

`http://localhost:9008/res/ChatApp/1907a5418dddc44bffa8be6cacd7cc50`

You can also access a straightforward URL using `response.download_url`