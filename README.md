Buzzer
======

Buzzer is a microblogging service written in Go on which users socialize by
posting messages known as "buzzes". Registered users can subscribe to another
users posts which appear on their "buzz-feed" along with any message in which
they were mentioned (@username). Additionally, anyone can search for messages
by tags (#topic).

Buzzer uses Go's channels and goroutines to coordinate the asynchronous
activity and exposes the service via a WebSockets-based API for real-time,
bidirectional communication with a minimal web client.

Buzzer was written solely by Taeber Rapczak using the Git ID
Taeber Rapczak <taeber@rapczak.com>.

Quick Start
-----------

Install Go from https://golang.org/dl/.

    $ make
    $ ./buzzer src/client
    $ open http://localhost:8080/

Alternatively, you can connect directly to the WebSocket server, using a
WebSocket client such as https://github.com/hashrocket/ws.git, but the
protocol would have to be inferred from the accept() function in ws.go.

    $ ws ws://localhost:8080
    > register user pass
    > login user pass
    > post Hello there!
    > logout


Design
------

### Server

The main components of the server are:

 * kernel
 * channelServer
 * Message
 * User
 * wsClient

`kernel` is a implementation of Server that can only be used serially.

`channelServer` implements the Server interface and essentially puts a layer of
channels in front of the actual kernel to provide safe, concurrent access.

`Message` is a message posted by a user.

`User` is a person or bot that uses the service.

`wsClient` represents a client connected to the WebSocket server.


### Client

The client is a React.js application located in `src/client/`.


Testing
-------

There are some unit tests which can be run using:

    $ make test

There is also some benchmarking tests:

    $ make benchmark
