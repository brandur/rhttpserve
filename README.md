# rhttpserve [![Build Status](https://travis-ci.org/brandur/rhttpserve.svg?branch=master)](https://travis-ci.org/brandur/rhttpserve)

An HTTP server that can serve files out of an rclone
remote. Includes a command line utility to generate time
expiring signed URLs (in the style of S3) that will be
verified by the server.

This has the effect of allowing files to be shared simply
(and temporarily) without having to walk through web-based
sharing prompts and the like.

## Build

    $ go install

## Setup

Generate a public/private key pair:

    $ rhttpserve generate

This will produce `RHTTPSERVE_PUBLIC_KEY` and
`RHTTPSERVE_PRIVATE_KEY`, which you will need to set up the
server and client respectively.

### Server

The server needs to be configured with a public key so that
it can verify requests signed by client private keys:

    $ export RHTTPSERVE_PUBLIC_KEY=

Any rclone remotes you plan on serving files from should
also be configured in the environment:

    $ export RCLONE_CONFIG_MYREMOTE_TYPE="amazon cloud drive"
    $ export RCLONE_CONFIG_MYREMOTE_CLIENT_ID=
    $ export RCLONE_CONFIG_MYREMOTE_CLIENT_SECRET=
    $ export RCLONE_CONFIG_MYREMOTE_TOKEN=

The remote above is called `MYREMOTE` and can be reference
below with `myremote:`. Naming conventions by rclone's
normal standard.

The server can then be started with:

    $ rhttpserve serve

It defaults to listening on port 8090, but tries to read a
value out of `PORT` if one is configured.

### Client

The client needs a private key and the host that the server
is listening on:

    $ export RHTTPSERVE_PRIVATE_KEY=
    $ export RHTTPSERVE_HOST=localhost:8090

We use a local host value, but it could just as easily be
something like `serve.example.com`, just as long as
rhttpserve is listening on that server.

Because you'll likely be running the client locally, it
might be useful to store these values in your `.zshrc` or
equivalent.

## Usage

With both server and client set up, it's now possible to
have rhttpserve generate a URL for a file in your remote:

    $ rhttpserve sign myremote:papers/raft.pdf
    https://serve.example.com/myremote/papers/raft.pdf?expires_at=1484239044&signature=QH816bQ_OlGDIIOHfhFYYTlSvVqtlNyboRgQDLJLp1R6wEU4tivChyPXIOOKETH_kvWN-UEakhNgVFU00jdIAA==

Or change the output to be a cURL command:

    $ rhttpserve sign --curl myremote:papers/raft.pdf
    curl -o 'raft.pdf' 'https://serve.example.com/myremote/papers/raft.pdf?expires_at=1484239058&signature=x7u1d6D3TXyieXEQ88wTcrheQWm6NI9wBGFbJbqjliq6YiRO38OSeB777xFUZ46tNlnnTCaYpoxNWRYNVIl1BA=='

Compose with `xargs` to sign all files in a directory:

    $ rclone ls -q myremote:papers/ | awk '{print "myremote:papers/" $2}' | xargs rhttpserve sign --curl --skip-check

## Development

## Run Tests

    $ go test $(go list ./... | egrep -v '/vendor/')

## Vendoring Dependencies

Dependencies are managed with govendor. New ones can be vendored using these
commands:

    $ go get -u github.com/kardianos/govendor
    $ govendor add +external
