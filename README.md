
  

# flyvo-api

  

  

**Introduction**

  

Flyvo-api is the RPC server receiving and sending requests to/from the RPC client and exposing an API to retrieve various data.

Data received from FlyVo through the RPC client is pushed to google-calendar.

  

**How to build**

  

  

1. Build as a runable file

  

  

We use make to build our projects. You can define what system to build for by configuring the GOOS environment variable.

  

  

  

>\> GOOS=windows make clean build

  

  

  

>\> GOOS=linux make clean build

  

  

  

These commands will build either a runable linux or windows file in the /bin/amd64 folder

  

  

  

2. Build a docker container

  

  

First you have to define the docker registry you are going to use in "envfile". Replace the REGISTRY variable with your selection.

  

  

Run the following command

  

  

  

>\> GOOS=linux make clean build container push

  

  

  

This will build the container and push it to your docker registry.

  

  

  

**How to run**

  

  

1. Executable

  

  

If you want to run it as a executable (Windows service etc) you will need to configure the correct environment variable. When you start the application set the CONFIG environment to **file::\<location\>** for linux or run it as a argument for windows

  

  

  

Windows example: **set "CONFIG=file::Z://folder//cfg.yml" & flyvo-api.exe**

  

Linux example: **CONFIG=file::../folder/cfg.yml ./flyvo-rpc-client**

  

  

  

2. Docker

  

  

You have to first mount the cfg file into the docker container, and then set the config variable to point to that location before running the service/container

  

**Configuration file**

  

**debug:** Set to true/false to enable debugging of application

  

**qrUrl:** URL that is contacted to generate a QR code. The code is implemented to run against https://github.com/samwierema/go-qr-generator that you have to run as an external service.
You can implement your own QR API, or do it directly in the code by changing generateQrCode function in /internal/api/qrGenerator.go
 

**gcalUrl:** URL to the google calendar integration. This variable is only used to retrieve calendar data.

**absentCron:** Since FlyVo has no way to register participants but only absentees we have to run a daily cron job that reverses the participation list to see who has been absent from the course. We have decided to run this 02:00 each night.

**rpc.port:** What port the RPC server should expose. This port will be used by the RPC client to connect to the server.

**rpc.gcalUrl:** Url to google calendar integration. This configuration variable is used to push events to. Because google has limitations on how many events/invites we can create you might want to use the google calendar queue URL here https://github.com/tktip/flyvo-calendar-queue but if you dont have any issues by the limit you can just use https://github.com/tktip/google-calendar. In that case it would be the same URL as the other gcalUrl.

**rpc.cert:** Contains the filepath of the public certificate if you want to run with encryption. If you do not need any encryption between the server and the client leave this blank.

**rpc.key:** Contains the filepath of the private certificate if you want to run with encryption. If you do not need any encryption between the server and the client leave this blank.

**redis.url:** We use redis to store generated participation URLs and to register participations. This should point to the redis instance.

**redis.db:** Redis supports out of the box 16 logical databases. Each database is separated from eachother. This value should be between 0 and 15.

**redis.password:** Redis password

**redis.ttl:** How long should we cache the participation codes, participations etc. The participations will be removed once the daily absentees sync runs. If its set to 24h the teacher can generate the participation code 24h before the course starts. After this it will be invalid and he has to generate another one if anyone still needs to register participation.

**trovo.creds:** We use google groups to figure out if a logged in used is a teacher. Creds should point to a google credential file used to talk to the google APIs.

**trovo.adminUser:** What user should we impersonate while using the google APIs

**trovo.teacherGroup:** The group the user needs to be part of to be concidered a teacher.