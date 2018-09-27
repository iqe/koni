# Koni

Koni handles [Mozilla Autoconfig](https://developer.mozilla.org/en-US/docs/Mozilla/Thunderbird/Autoconfiguration) and [Microsoft Autodiscover](https://docs.microsoft.com/en-us/exchange/client-developer/exchange-web-services/autodiscover-for-exchange) for all your domains, in one place. With automatic Let's Encrypt certificates for extra convenience.

## Installation

These are the basic steps needed to install koni:

1. Download a release from the GitHub release page
2. Extract koni-$version.tar.gz to your server
3. Edit koni.conf to match your environment
4. Optional: Customize and install koni.service systemd unit
5. Make koni reachable on port 443

   On Linux there are various methods to do this. See e.g. https://superuser.com/questions/710253/allow-non-root-process-to-bind-to-port-80-and-443

   You could also setup a tcp proxy like[haproxy](https://www.haproxy.org/)

6. Run koni (through systemd or directly)

## DNS Setup

For each domain you want to handle, add the following CNAME entries to DNS:

```
CNAME autoconfig.userdomain.com    -> koniserver.mydomain.com
CNAME autodiscover.userdomain.com  -> koniserver.mydomain.com
```

Additionally, you can set up a SRV record for clients that only used the SRV record during Autodiscover:

```
SRV _autodiscover._tcp.userdomain.com -> koniserver.mydomain.com:443
```

Koni listens for HTTPS requests on `koniserver.mydomain.com` and responds to any clients that request an URL from a `autoconfig.*` or `autodiscover.*` host or directly from `koniserver.mydomain.com`.

If a user configures their email client, the following happens:

1. User starts mail configuration on the client, enters email address `user@userdomain.com`.
2. Mail client looks up `autoconfig.userdomain.com` (Mozilla and others) or `autodiscover.userdomain.com` (Microsoft and others)
3. Mail client sends HTTP(s) request to the domain
4. Koni looks for a certificate of `autoconfig.userdomain.com` / `autodiscover.userdomain.com` in its certs cache dir. If there is no cert or the cert is expired, koni requests a certificate from Let's Encrypt for the requested domain
5. Koni sends HTTP response to client, with valid TLS cert.
6. Mail client proceeds with auto config of the user's email account

## Configuration

See comments in `koni.conf`.

## Contributing / Building

1. Clone the repo
2. Install [dep](https://golang.github.io/dep/) for dependency management
3. Run `make deps` to install/update the vendored dependencies
4. Hack on the code
5. Run `git tag -a v<NEW VERSION>`
6. Run `make release` to build a release package
7. Run `git push --tags` to push changes to GitHub
8. Upload the release to GitHub
