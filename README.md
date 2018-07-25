# quic-loadtest
Simple bidirectional stresstest using the QUIC-protocol. This utilizes
the go implementation of the proprietary QUIC protocol
by
[https://github.com/lucas-clemente/quic-go](https://github.com/lucas-clemente/quic-go) and
heavily borrows from the echo server in the example section of this
project. The tool can either start a server or a client connecting to
this server and transmit data from the client to the server as fast as
connection and protocol allows. Data is echoed back to the
client. Purpose is to put load onto a communication link other than
plain old TCP, e.g. for testing the stability of your DSL link. 

## Installation

Install the utility by calling `go get
github.com/stmichaelis/quic-loadtest`. You should have the
`quic-loadtest` binary in your GOPATH/bin-folder.

## Usage example

Sample call for starting the server:
``
quic-loadtest -q -s <publicfacingip>:4242
``

As always: Running a server on a public facing IP typically is a bad
idea, so keep the time as short as possible. The server waits for
exactly one incoming connection and terminates after a timeout without
any further data, i.e. the client stopped sending.

Sample call for starting the client and connecting to the server:
``
quic-loadtest -q -c <serverip>:4242 -d 60 -p 1300
``
Runs the test for 60 seconds, using packets of 1,300 Bytes. The quiet
flag is recommended for both sides, as otherwise each send/received
packet is confirmed with a character (r/s/.) on the commandline, which
may reduce your throughput. 


## Flags

    -s Server mode. Set address and port to listen on, e.g. localhost:4242
	-c Client mode. Specify address and port to connect to
	-d Duration in seconds, default is 10
	-p Payload buffer size in bytes, default is 1,000
	-q Quiet mode. Suppress output of s/r/. packet indicators
	
