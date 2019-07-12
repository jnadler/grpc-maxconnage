# grpc-maxconnage
This test uses a streaming client that sends data for a long time... longer than the configured
keepalive.MaxConnectionAge on the server.

That initially works fine; if you run this test you can see that it reconnects and data continues
to flow:

```
GOT POINT: 3500000
INFO: 2019/07/12 15:02:44 pickfirstBalancer: HandleSubConnStateChange: 0xc0000a4180, TRANSIENT_FAILURE
INFO: 2019/07/12 15:02:44 pickfirstBalancer: HandleSubConnStateChange: 0xc0000a4180, CONNECTING
INFO: 2019/07/12 15:02:44 pickfirstBalancer: HandleSubConnStateChange: 0xc0000a4180, READY
GOT POINT: 4000000
```

When we reach keepalive.MaxConnectionAge + keepalive.MaxConnectionAgeGrace, the client gets an `EOF` 
error and fails.   
