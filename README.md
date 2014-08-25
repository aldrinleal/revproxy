# revproxy

Generic Reverse Proxy in Go. Useful to nest multiple services under supervisor and friends

## Installation

go get bitbucket.org/aldrinleal/revproxy/revproxy

## Practical Example

Suppose you've want to nest three different services, so you want to reverse-proxy:

  * /app to localhost:8001
  * /api to localhost:8002
  * /static to localhost:8003

A sample line would be:

```
$ revproxy [-port <port>] /app:8001 /api:8002 /static:8003
```

It was created with AWS Elastic Beanstalk in mind, since it only allows a single EXPOSE port, instead of dealing with the need to install nginx/manage.

## Things that revproxy (or its author) doesn't want to do:

  * Log your requests (however, we'd like to add some tracing and instrumentation into it)
  * Validate data in general
  
Its more like a kludge. However, further versions are likely to include support for etcd as well as WebSockets.

## Kudos

A Big thank you goes to [azer/boxcars](https://github.com/azer/boxcars), which gave me the impetus to try my own stab at it