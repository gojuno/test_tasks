# Juno Inc., Test Task: In-memory cache

Simple implementation of Redis-like in-memory cache

Desired features:
- Key-value storage with string, lists, dict support
- Per-key TTL
- Operations:
  - Get
  - Set
  - Update
  - Remove
  - Keys
- Custom operations(Get i element on list, get value by key from dict, etc)
- Golang API client
- Telnet-like/HTTP-like API protocol

Provide some tests, API spec, deployment docs without full coverage, just a few cases and some examples of telnet/http calls to the server. 

Optional features:
- persistence to disk/db
- scaling(on server-side or on client-side, up to you)
- auth
- perfomance tests
