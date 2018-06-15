# sysflux
Sysflux transforms RFC3164 syslog messages sent via UDP into influxdb data points.
Any log message can be parsed by defining a custom regular expression.

## Regular Expression Syntax
All named capture groups of the regex are ether parsed to a datapoint value or a datapoint tag.
To parse a tag, add the "tag_" prefix to the name of the capture group.
Values are parsed as floats if the "val_" prefix is configured in the name of the capture group.

## Example: NGINX Upstream Timing
Configure custom access log format in NGINX and send to a remote syslog server.

**/etc/nginx/nginx.conf:**

    ...
    log_format  metric  '$host $uri $upstream_status $upstream_connect_time $upstream_header_time $upstream_response_time $request_time';
    access_log syslog:server=<sysfluxip>:5014,facility=local7,tag=nginx,severity=info metric;
    ...

**/etc/sysflux/sysflux.yml:**

    influx:
      addr: http://127.0.0.1:8086
      user: test
      password: test
      database: test

    syslog:
      - measurement: http_proxy
        listen: 0.0.0.0:5014
        regex: "(?P<tag_host>.+)\\s+(?P<tag_url>.+)\\s+(?P<tag_status>.+)\\s+(?P<val_upstream_connect>.+)\\s+(?P<val_upstream_header>.+)\\s+(?P<val_upstream_response>.+)\\s+(?P<val_request_time>.+)"

## Disclaimer
This software has not yet been tested in production. Please be carefull!
