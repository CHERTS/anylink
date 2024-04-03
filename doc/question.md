## common problem

### anyconnect client issue

> Please use the version of the group shared file for the client. Other versions have not been tested and are not guaranteed to work properly.
>
> Add Telegram: @cherts

### OTP dynamic code

> Please use your mobile phone to install freeotp, and then scan the otp QR code. The generated number is the dynamic code.

### User policy issues

> As long as there is a user policy, the group policy will not take effect, which is equivalent to overwriting the configuration of the group policy.

### Remote Desktop connection

> This software already supports anyconnect connection in remote desktop.

### Private certificate issue

> anylink does not support private certificates by default
>
> For other problems using private certificates, please solve them by yourself

### Client connection name

> The client connection name needs to be modified in the [profile.xml](../server/conf/profile.xml) file

```xml

<HostEntry>
     <HostName>VPN</HostName>
     <HostAddress>localhost</HostAddress>
</HostEntry>
```

### dpd timeout setting problem

```yaml
#Client failure detection time (seconds) dpd > keepalive
cstp_keepalive=4
cstp_dpd = 9
mobile_keepalive = 7
mobile_dpd = 15
```

> The above dpd parameter is the client’s timeout detection time. If there is no data transmission within a period of time, the firewall will actively close the connection.
>
> If timeout error messages appear frequently, the dpd value should be appropriately reduced according to the current firewall settings.

### Reverse proxy problem

> anylink only supports four-layer reverse proxy and does not support seven-layer reverse proxy.
>
> If Nginx, please use the stream module

```conf
stream {
     upstream anylink_server {
         server 127.0.0.1:8443;
     }
     server {
         listen 443 tcp;
         proxy_timeout 30s;
         proxy_pass anylink_server;
     }
}
```

> nginx implementation example of sharing port 443

```conf
stream {
     map $ssl_preread_server_name $name {
         vpn.xx.com myvpn;
         default defaultpage;
     }
    
     # upstream pool
     upstream myvpn {
         server 127.0.0.1:8443;
     }
     upstream defaultpage {
         server 127.0.0.1:8080;
     }
    
     server {
         listen 443 so_keepalive=on;
         ssl_preread on;
         #The receiving end also needs to set proxy_protocol
         #proxy_protocol on;
         proxy_pass $name;
     }
}

```

### Performance issues

```
Intranet environment test data
Virtual server: centos7 4C8G
anylink: tun mode tcp transmission
Client file download speed: 240Mb/s
Client network card download speed: 270Mb/s
Server network card upload speed: 280Mb/s
```

> Client TLS encryption protocol and tunnel header will occupy a certain amount of bandwidth
