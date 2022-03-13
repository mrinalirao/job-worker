# job-worker
The Job worker service provides an API to run arbitrary Linux processes. 

## Overview

### The Library

The Library supports the following features:
- **Start Job**: A job is a linux command which is represented internally by a JobID. A random UUID is assigned to the underlying job
- **Stop Job**: Stops a job with the given JobID. For this POC, we send a SIGKILL signal to the running process.
- **Get Job Status**: Gets the status of a job with the given JobID. Jobs can have 3 statuses:
        
    - RUNNING: Job process started
    - STOPPED: Job is force stopped
    - FINISHED: Job finished successfully or exited with error
- **Stream Output**: When a user streams the output of a job with the given JobID, the worker adds the userID as a subscriber of the job.
  The output is then published to all the active listeners until no more data is left to stream or a job is stopped forcefully, whichever happens first.
  Both Stderr and Stdout output will be combined for the sake of simplicity.

The library also adds resource control using **cgroups V2**. The limits for CPU, memory and Disk IO will be hardcoded in the codebase itself.
In a production system, these would be passed in via a config file.

### The API

Exposes gRPC API's to access the underlying library's functionality over network. The API layer is responsible for authentication and authorization.
The service and message definitions can be found [here](./proto/workerservice.proto)

### Security

#### Authentication

Authentification is based on x.509 certificates. Both the provider and the consumer require to produce their own certificates to the other party. These certificates are validated by both parties with their respective CAs.
For the purpose of this exercise, all certificates and keys are kept within the source code for clients and server to use.

#### Authorization

Authorization is based on role-based access control. For the purpose of this exercise, we support only 2 roles:
- **admin**: The admin user has access to all RPCs and additionally can access jobs of any user in the system
- **user**: The user role also has access to all RPCs but is restricted to have access to their own jobs and cannot access other jobs in the system

The AuthService is responsible for issuing JWT's to our users. The service and message definitions can be found [here](./proto/authservice.proto)

The JWT claims will contain information about the logged-in user's role.
For the purpose of this exercise we will pre-seed two users into the system, one with `admin` and the other with `user` role.
We also set the expiry time to 30 days.
In a production system we would expire the JWT every 15 minutes or so and the AuthService would expose an RPC to refresh the token. However, this is outside the scope of this exercise.

### The CLI

The CLI can be used to communicate with server over the network.
The CLI has a few base parameters that will need to be met for all subcommands

Some examples are provided below.

**Login Request**
This issues a JWT which will be needed to access all the underlying gRPC API's
```
./client login -u <username> -p <password>  
```

**Start Job**
Returns the Job ID of the job that is started
```
./client start -t <JWT_Token> -c  <command> -args <arg1> <arg2>
```

**StopJob**
Stops the job with the given ID
```
./client stop -t <JWT_Token> -j <JobID>
```

**StreamOutput**
Streams the output of the job with the given ID
```
./client stream -t <JWT_Token> -j <JobID>
```

### Trade-Offs
- All data is stored in memory, for a production grade service we would need persistent storage to store all the user as well as job information.
- CA and certificates will be generated manually using openssl. 
- Users/Roles will be pre-seeded on the server side.
- Configuration will be harded in the app itself
- The scope of this project would only deal with a single linux worker server interfacing with multiple clients


