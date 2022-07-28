### tiny remote login

This is a tiny remote login program. It has no authentication and it's totally
insecure, it's just for learning.

### how to use it

To start the server:

```
$ go run server.go bash
```

To connect as a client:

```
stty raw -echo && nc localhost 7777  && stty sane
```

or

```
go run client.go localhost 7777
```
