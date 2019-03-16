                                     Buzzer

Buzzer is a microblogging service written in Go on which users socialize by
posting messages known as "buzzes". Registered users can subscribe to another
users posts which appear on their "buzz-feed" along with any message in which
they were mentioned (@username). Additionally, anyone can search for messages
by tags (#topic).

Buzzer uses Go's channels and goroutines to coordinate the asynchronous
activity and exposes the service via a WebSockets-based API for real-time,
bidirectional communication with a minimal web client.

Buzzer will be written solely by Taeber Rapczak <taeber@ufl.edu> for his
COP5618 Spring 2019 class project to demonstrate the effectiveness of Go's
implementation of ideas from Sir Tony Hoare's 1978 CSP paper.

Hoare, C. A. R. (1978). "Communicating sequential processes". Communications of the ACM. 21 (8): 666–677. doi:10.1145/359576.359585.
