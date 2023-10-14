# dumbdl

## Why

Impersonate TLS fingerprint and akamai fingerprint. I need a reliable tool as the backend of my crawler.

```powershell
go build
./dumbdl.exe 2023-10-12-1-5-8.log.json -o out/15d
```

## TODO

- [ ] an HTTP/WebSocket interface 
- [x] parallel download
- [ ] retry after failure and logging failure to a file

## See also

- [curl-impersonate](https://github.com/lwthiker/curl-impersonate)
- [curl_cffi](https://github.com/yifeikong/curl_cffi)
- [How to Bypass Cloudflare](https://www.zenrows.com/blog/bypass-cloudflare)
