# Juno Inc., Test Task: In-memory cache

Simple implementation of Redis-like in-memory cache

Desired features:
- Key-value storage with string, lists, dict support
- Per-key TTL
- Operations:
  - Get
  - Set
  - Remove
- Golang API client
- Provide some tests, API spec, deployment docs without full coverage, just a few cases and some examples of telnet/http calls to the server.

Optional features:
- Telnet-like/HTTP-like API protocol
- performance tests
- Operations:
  - Keys
