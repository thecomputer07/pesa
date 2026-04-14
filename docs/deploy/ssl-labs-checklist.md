# SSL Labs A+ checklist

Goal: `https://www.ssllabs.com/ssltest/analyze.html?d=pesadia.com` returns A+
and Mozilla Observatory returns A+.

| Control | Status |
|---------|--------|
| TLS 1.3 only (TLS 1.2 fallback off unless required) | enforced in `runTLS()` (`tls.VersionTLS13`) |
| Cipher suites: AES-GCM + ChaCha20-Poly1305 only | enforced (3 suites whitelisted) |
| Forward secrecy (ECDHE only) | TLS 1.3 mandates this |
| OCSP stapling | autocert handles |
| HSTS `max-age=63072000; includeSubDomains; preload` | `SecurityHeaders` middleware |
| HSTS preload submitted at https://hstspreload.org | manual after first deploy |
| Certificate chain complete | autocert |
| DNS CAA record `0 issue "letsencrypt.org"` | manual |
| HTTP -> HTTPS 301 redirect on `:80` | autocert HTTPHandler |
| HTTP/2 enabled | `net/http` defaults |
| `Content-Security-Policy` strict | `SecurityHeaders` |
| `X-Frame-Options: DENY` | `SecurityHeaders` |
| `Referrer-Policy: strict-origin-when-cross-origin` | `SecurityHeaders` |
| `Permissions-Policy` deny camera/mic/geo/payment/usb | `SecurityHeaders` |
| `Cross-Origin-Opener-Policy: same-origin` | `SecurityHeaders` |
| Cookies `Secure`, `HttpOnly` (session), `SameSite=Lax` | scs config + handlers |

## Verification commands

```bash
# Headers
curl -sI https://pesadia.com/ | grep -iE 'strict-transport|content-security|x-frame|permissions-policy'

# TLS info
nmap --script ssl-enum-ciphers -p 443 pesadia.com

# Online scans (bookmark)
# https://www.ssllabs.com/ssltest/analyze.html?d=pesadia.com
# https://observatory.mozilla.org/analyze/pesadia.com
```
