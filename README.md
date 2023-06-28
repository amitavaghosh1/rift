## Requirements:
- golang


## Usage

This cli tool expects you to pass a proto file to parse log messages. Each log backend can have its own proto to be parsed.

#### Cloudwatch
A sample proto for a sample cloudwatch log stream has been provided in `proto/cloudwatch/log.proto`.

There are few fields that are required for a cloudwatch log.
- time         : timestamp of the request
- request_id   : a trace id for each request
- msg          : the actual message
- level        : log level of the message


#### SQS

Presently it displays in whatever is there in request body


### TODO:
- [ ] Check possibility of using proto. And how to parse non-JSON body
- [ ] Improve code
- [ ] Support more backends
- [ ] Suppor only matching flag
