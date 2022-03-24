# job-worker
The Job worker service provides an API to run arbitrary Linux processes. 

## Overview

### The Library

The Library supports the following features:
- **Start Job**: A job is a linux command which is represented internally by a JobID. A random UUID is assigned to the underlying job
- **Stop Job**: Stops a job with the given JobID. For this POC, we send a SIGKILL signal to the running process.
- **Get Job Status**: Gets the status of a job with the given JobID and the exit code of the process. Jobs can have 3 statuses:

    - RUNNING: Job process started
    - STOPPED: Job is force stopped
    - FINISHED: Job finished successfully or exited with error
  
  For this exercise, the Library will keep the job status in memory (in a map),if the library goes down this data will be lost. 
  
- **Stream Output**: When a user streams the output of a job with the given JobID, the worker adds the userID as a subscriber of the job.
  The output is then published to all the active listeners until no more data is left to stream or a job is stopped forcefully, whichever happens first.
  Both Stderr and Stdout output will be combined for the sake of simplicity.
  
  Multiple clients should be able to read the output of the job from the beginning. When a job is started, the worker will add the output of the job to a temporary file and monitor for changes to the file event.
  However, the drawback of such a system, is the files could consume a lot of disk space and can potentially crash the app.
  In a production system this would probably be stored in distributed file system instead.

The library also adds resource control using **cgroups V2**. The limits for CPU, memory and Disk IO will be hardcoded in the codebase itself.
In a production system, these would be passed in via a config file.

### The API

Exposes gRPC API's to access the underlying library's functionality over network. The API layer is responsible for authentication and authorization.
The service and message definitions can be found [here](./proto/workerservice.proto)

### Security

Transport security is based on TLS 1.3.

#### Authentication

Authentification is based on x.509 certificates. Both the provider and the consumer require to produce their own certificates to the other party. These certificates are validated by both parties with their respective CAs.
For the purpose of this exercise, all certificates and keys are kept within the source code for clients and server to use.

#### Authorization

Authorization is based on role-based access control. For the purpose of this exercise, we support only 2 roles:
- **admin**: The admin user has access to all RPCs and additionally can access jobs of any user in the system
- **user**: The user role also has access to all RPCs but is restricted to have access to their own jobs and cannot access other jobs in the system

The user role will be added into the client certificate as an extention. We will use roleOid 1.2.840.10070.8.1 = ASN1:UTF8String for the client certificate.
And have the server read and verify the roles to authorize the client.

### The CLI

The CLI can be used to communicate with server over the network.
The CLI has a few base parameters that will need to be met for all subcommands

Some examples are provided below.


**Start Job**
Returns the Job ID of the job that is started
```
./client start -c  <command> -args <arg1> <arg2>
```

**StopJob**
Stops the job with the given ID
```
./client stop -j <JobID>
```

**GetStatus**
Returns the job status of the job with the given ID
```
./client status -j <JobID>
```

**StreamOutput**
Streams the output of the job with the given ID
```
./client stream -j <JobID>
```

### Trade-Offs
- All data is stored in memory, for a production grade service we would need persistent storage to store all the user as well as job information.
- CA and certificates will be generated manually using openssl. 
- Users/Roles will be pre-seeded on the server side.
- Configuration will be harded in the app itself
- The scope of this project would only deal with a single linux worker server interfacing with multiple clients
- Most of the time, the users want to see the full log content to check if the job performs as expected. The Worker writes the process output (stderr/stdout) on the disk as a log file. On the other hand, the old log files consume disk space and can potentially crash the system when no more space is left. Also a malicious or misconfigured program could potentially truncate the output file.

### How to run this

## Generate proto files
```sh
$ make protofile
```

## Test

```sh
$ make test
```
## Build and run API

```sh
$ make api
go build -o ./job-worker cmd/main.go
```

```sh
$ ./job-worker
```

## Build and run client

```sh
$ make client
go build -o ./client cli/client/userclient.go
```

Examples:
Note: You must provide the role of the user while running client commands, the role can be one of [user, admin]
```sh
$ ./client start -r user -c bash -args "-c" "while true;do date;sleep 1;done"
started JobID: aa0319d4-e37d-420b-ac43-7316f6b032e5
```

```sh
$ ./client status -r user -j aa0319d4-e37d-420b-ac43-7316f6b032e5
jobID: aa0319d4-e37d-420b-ac43-7316f6b032e5, status: RUNNING, exitCode: 0
```

```sh
$ ./client stream -r user -j aa0319d4-e37d-420b-ac43-7316f6b032e5
Wed Mar 23 15:05:03 AEDT 2022
Wed Mar 23 15:05:04 AEDT 2022
Wed Mar 23 15:05:05 AEDT 2022
Wed Mar 23 15:05:06 AEDT 2022
Wed Mar 23 15:05:07 AEDT 2022
Wed Mar 23 15:05:08 AEDT 2022
Wed Mar 23 15:05:10 AEDT 2022
```

```sh
 ./client stop -r user -j aa0319d4-e37d-420b-ac43-7316f6b032e5
stopped Job: aa0319d4-e37d-420b-ac43-7316f6b032e5
```
