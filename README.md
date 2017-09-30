One time download
=================

Simple golang server for one time download.

Place files you want to serve into `files`

For example you can add `files/test.html`.
Example run:
```sh
BIND_ADDRESS=localhost:8080 go run main.go
Starting server on localhost:8080...
```

```sh
$ curl -i localhost:8080/test.html
HTTP/1.1 200 OK
Accept-Ranges: bytes
Content-Length: 14
Content-Type: text/html; charset=utf-8
Last-Modified: Thu, 28 Sep 2017 05:32:48 GMT
X-Times-Remaining: 0
Date: Thu, 28 Sep 2017 05:33:14 GMT

<h1>Hello</h1>
$ curl -i localhost:8080/test.html
HTTP/1.1 410 Gone
Content-Type: text/html
Date: Thu, 28 Sep 2017 05:33:20 GMT
Content-Length: 66

<center><h1>File is no longer available for download</h1></center>
```
