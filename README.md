# Koni

Koni handles [Mozilla Autoconfig](https://wiki.mozilla.org/Thunderbird:Autoconfiguration), [Microsoft Autodiscover](https://learn.microsoft.com/en-us/exchange/client-developer/exchange-web-services/autodiscover-for-exchange) and Apple mobileconfig profiles for all your domains, in one place. With automatic Let's Encrypt certificates for extra convenience.

## Installation

These are the basic steps needed to install koni:

1. Download a release from the GitHub release page
2. Extract koni-$version.tar.gz to your server
3. Edit koni.conf to match your environment
4. Optional: Customize and install the koni.service systemd unit
5. Make koni reachable on ports 80 and 443

   On Linux there are various methods to do this. See e.g. https://superuser.com/questions/710253/allow-non-root-process-to-bind-to-port-80-and-443

   Alternatively, set up a TCP proxy like [haproxy](https://www.haproxy.org/) in front of koni.

6. Run koni (through systemd or directly)

## DNS Setup

For each domain you want to handle, add the following CNAME entries to DNS:

```
CNAME autoconfig.userdomain.com    -> koniserver.mydomain.com
CNAME autodiscover.userdomain.com  -> koniserver.mydomain.com
```

Additionally, you can set up a SRV record for clients that only use the SRV record during Autodiscover:

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

## Endpoints

| Path | Purpose |
|---|---|
| `/mail/config-v1.1.xml`, `/.well-known/autoconfig/mail/config-v1.1.xml` | Mozilla Autoconfig (`?emailaddress=`) |
| `/autodiscover/autodiscover.xml`, `/Autodiscover/Autodiscover.xml` | Microsoft Autodiscover (POX, POST) |
| `/autodiscover/autodiscover.json` | Microsoft Autodiscover JSON, redirects to the XML endpoint |
| `/mobileconfig.xml` | Apple configuration profile (`?emailaddress=`) |

Plain HTTP requests are redirected to HTTPS, except for ACME http-01 challenges.

## Configuration

See comments in `koni.conf`. Unknown settings are rejected at startup.

Koni logs to stderr in `key=value` format without timestamps (systemd/journald adds those).
Set `debug = true` to additionally log full request dumps.

Each access log line has a `category`: `autoconfig`, `autodiscover`, `mobileconfig`, `acme` (Let's Encrypt
challenges) or `noise` (everything else, mostly bots scanning for vulnerabilities). For example:

```
journalctl -u koni | grep msg=request | grep -v category=noise           # real traffic only
journalctl -u koni | grep -o 'category=[a-z]*' | sort | uniq -c         # requests per category
```

Note that bots also probe the real endpoints, especially `autodiscover`. Requests from real clients usually
succeed (`status=200`), bot probes typically get `status=400`.

### Templates

The response templates in `templates/` (Go [text/template](https://pkg.go.dev/text/template) syntax) are built into
the binary. To change them, edit the files and rebuild koni. All values are XML-escaped before rendering.

## Upgrading from 0.5.x

* Templates are now built into the binary; the `templates/` directory is no longer needed at runtime.
* The access log is now structured (`level=INFO msg=request client=... status=...`) instead of Apache format.
* `debug` is now a TOML boolean (`debug = true`); the old `"yes"`/`"no"` values are rejected.
* `letsencrypt.email` is optional.

## Contributing / Building

Requires Go 1.26 or newer (`go.mod` selects the toolchain automatically).

1. Clone the repo
2. Hack on the code
3. Run `make check` (vet, gofmt, staticcheck, tests, govulncheck)
4. Run `git tag -a v<NEW VERSION>`
5. Run `make release` to build a release package
6. Run `git push --tags` to push changes to GitHub
7. Upload the release to GitHub
