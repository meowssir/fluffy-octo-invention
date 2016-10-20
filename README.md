# fluffy-octo-invention
Asynchronous oplog tailing using the WebSocket Protocol

# Abstract

Abstract This design is intended to supersede the existing bidirectional communication technologies that use HTTP as a transport layer. In brief, the WebSocket Protocol provides a simple solution by using a single TCP connection for traffic in both directions reducing the need to abuse HTTP by polling the server for updates while sending upstream notifications as distinct HTTP calls. The low-latency nature of the WebSocket Protocol is ideal for interacting with the *oplog*, a continuously updated capped collection which stores an ordered history of logical writes to a MongoDB database.
