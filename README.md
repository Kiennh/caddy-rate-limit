## caddy-rate-limit


### New

 - Allow rate limit for localIP address (Optional)

 - Allow rate limit for specific http method (GET, POST, ..), * for rate limit allmethod

 - Caddyfile example:

```

...

  ratelimit  2 3 minute  {
     allowLocalIPs 192.168.0.0/6
     resources GET / ^/profiles ^/static
     resources POST /profiles
  }


...

```

>a `rate limit` plugin for [caddy](https://caddyserver.com/)

### Syntax

**Excessive requests will be terminated with an error 429 (Too Many Requests)! And `X-RateLimit-RetryAfter` header will be returned.**

**[Private IPs](https://en.wikipedia.org/wiki/Private_network) will be ignored.**

For single resource:

```
ratelimit path rate burst unit
```

- path is the file or directory to apply `rate limit`

- rate is the limited request in every time unit (r/s, r/m, r/h) (e.g. 1)

- burst is the maximum burst size client can exceed; burst >= rate (e.g. 2)
 
- unit is the time interval (currently support: `second`, `minute`, `hour`)

For multiple resources:

```
ratelimit rate burst unit {
    resources
}
```

- resources is a list of files/directories to apply `rate limit`, one per line

**Note:** If you don't want to apply `rate limit` on some special resources, add `^` in front of the path.


### Examples

Limit clients to 2 requests per second (bursts of 3) to any resources in /r:

```
ratelimit /r 2 3 second
```

For the listed paths, limit clients to 2 requests per minute (bursts of 2) and always ignore `/dir/app.js`:

```
ratelimit 2 2 minute {
    /foo.html
    /dir
    ^/dir/app.js
}
```

### Test

```bash
docker pull xuqingfeng/caddy-rate-limit
docker run -d -p 2016:2016 --name ratelimit xuqingfeng/caddy-rate-limit
```

---

**Inspired By**

[http://nginx.org/en/docs/http/ngx_http_limit_req_module.html](http://nginx.org/en/docs/http/ngx_http_limit_req_module.html)

[https://github.com/didip/tollbooth](https://github.com/didip/tollbooth)
