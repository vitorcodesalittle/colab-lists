# Colablists
This app was built as a learning project. So it has some rough edges.
It's designed to be a collaboratively list builder, where multiple people
can update the list simultenously. 

It's using the following stack:
- Go lang
- HTMX
- Alpine.js
- Sqlite3

## Running

Simply run:

```bash
go run main.go
```

Configuration allows to change the no-reply email, SMTP credentials, point to TLS certificates, define session timeout, listening address and others. Full list is defined below (from `--help`):

```
Usage of /tmp/go-build3802330467/b001/exe/main:
  -app-url string
    	the URL of the app (default "https://lists.vilmasoftware.com.br")
  -certificate string
    	Path to file with certificate
  -database-url string
    	Database URL (default "./data/colablist.db")
  -hot-reload
    	If passed, will serve a websocket endpoint that identifies this run, allowing the client to restart
  -listen string
    	Listen (default ":8080")
  -private-key string
    	Path to file with private key
  -session-timeout duration
    	Session timeout (default 4h0m0s)
  -smtp-host string
    	SMTP Host
  -smtp-noreply string
    	SMTP Password (default "something.something.noreply@domain.com")
  -smtp-password string
    	SMTP Password
  -smtp-port int
    	SMTP Port
  -smtp-username string
    	SMTP Username
  -tls
    	Listen
```


## Future roadmap:

- [x] Real-time update of lists.
    - [ ] and marking items as gathered
- [ ] Explore delivery automation
- [ ] Explore diet planning
- [ ] Production ready check-list
    - [ ] Recaptcha in strategic forms
    - [ ] IP request rate limiting
    - [ ] Compressing responses (e.g. html is currently sent with lots of spaces)
