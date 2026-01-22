# external-dns-webhook-opnsense

A webhook provider for [external-dns](https://github.com/kubernetes-sigs/external-dns) that integrates with OPNsense's Unbound DNS service.

> [!IMPORTANT]
> CURRENTLY AT DOG FOODING STAGE. USE AT YOUR OWN RISK.

## How It Works

This webhook provider acts as a bridge between `external-dns` and OPNsense's Unbound DNS service.

```
┌─────────────────┐         ┌──────────────────────┐         ┌──────────────┐
│   External-DNS  │ ◄─────► │  Webhook Provider    │ ◄─────► │   OPNsense   │
│  (Kubernetes)   │  HTTP   │  (sidecar)           │   API   │  Unbound DNS │
└─────────────────┘         └──────────────────────┘         └──────────────┘
```

## Features

- Supports `A`, `AAAA`, `TXT` record types.
- Supports `TXT` registry for ownership management.
- Supports domain filtering capabilities of `external-dns`.
- Supports multiple targets for DNS records.
- Can run multiple instances for different domains and clusters as well as manual management. This is the feature goal ignited this implementation since [crutonjohn/external-dns-opnsense-webhook](https://github.com/crutonjohn/external-dns-opnsense-webhook) could do everything else, but this.
- Does not support `CNAME` records since OPNSense Unbound uses a different mechanism of having aliases for that purpose, however that still relies on original record to exist.
- Can support `MX` records however I have not implemented it yet, since I do not need it personally. Contributions are welcome.

## Installation

The recommended way to install this webhook is using the [external-dns Helm chart](https://github.com/kubernetes-sigs/external-dns/tree/master/charts/external-dns) with the webhook provider configured as a sidecar.

### Create the User in OPNSense

- Under `System/Access/Users`, create a new user with the username of your choice.
  - The user needs `Services: Unbound (MVC)`, `Services: Unbound DNS: Edit Host and Domain Override`, `Status: Services` permissions as minimum.
- For the given user create a API key pair and note it accordingly.

## Create the Secrets in Kubernetes

Create your secret with desired method.

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: external-dns-opnsense
stringData:
  url: "https://opnsense.example.com/api"
  insecure: "true" # set to "false" if you have valid SSL certs
  key: "your_api_key"
  secret: "your_api_secret"
```

## Deploy `external-dns` as Normal with the Webhook Provider

Deploy the `external-dns` Helm chart with the webhook provider configured.

```yaml
provider:
  name: webhook
  webhook:
    image:
      repository: docker.io/cenk1cenk2/external-dns-webhook-opnsense
      tag: latest # use the specific tag in production
    livenessProbe:
      httpGet:
        path: /healthz
    readinessProbe:
      httpGet:
        path: /readyz
      initialDelaySeconds: 3
      periodSeconds: 300
    env:
      - name: LOG_LEVEL
        value: warn
      - name: OPNSENSE_URL
        valueFrom:
          secretKeyRef:
            name: external-dns-opnsense
            key: url
      - name: OPNSENSE_ALLOW_INSECURE
        valueFrom:
          secretKeyRef:
            name: external-dns-opnsense
            key: insecure
      - name: OPNSENSE_API_KEY
        valueFrom:
          secretKeyRef:
            name: external-dns-opnsense
            key: key
      - name: OPNSENSE_API_SECRET
        valueFrom:
          secretKeyRef:
            name: external-dns-opnsense
            key: secret
      # as you desire
      - name: DOMAIN_FILTER
        value: |-
          example.com
```

```bash
helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/
helm repo update
# install with the given values or with your preferred deployment method
helm install external-dns external-dns/external-dns \
  --namespace external-dns \
  --create-namespace \
  -f values.yaml
```

## CLI

<!--- clidocs -->

### Application Settings

| Flag / Environment               | Description                                                                                           | Type                                                 | Required | Default |
| -------------------------------- | ----------------------------------------------------------------------------------------------------- | ---------------------------------------------------- | -------- | ------- |
| `--log-level` / `$LOG_LEVEL`     | Log level for the application.                                                                        | `enum("debug", "info", "warning", "error", "fatal")` | `false`  | `info`  |
| `--log-encoder` / `$LOG_ENCODER` | Log encoder format.                                                                                   | `enum("console", "json")`                            | `false`  | `json`  |
| `--port` / `$PORT`               | Port on which the webhook server will listen.                                                         | `uint16`                                             | `false`  | `8888`  |
| `--health-port` / `$HEALTH_PORT` | Port on which the health check server will listen.                                                    | `uint16`                                             | `false`  | `8080`  |
| `--dry-run` / `$DRY_RUN`         | The application will not make any changes to the OPNsense DNS records, only log the intended actions. | `bool`                                               | `false`  | `false` |

### OPNsense Connection

| Flag / Environment                                       | Description                                              | Type     | Required | Default |
| -------------------------------------------------------- | -------------------------------------------------------- | -------- | -------- | ------- |
| `--opnsense-url` / `$OPNSENSE_URL`                       | The base URI of the OPNsense API endpoint.               | `string` | `true`   | -       |
| `--opnsense-api-key` / `$OPNSENSE_API_KEY`               | The API key for authenticating with the OPNsense API.    | `string` | `true`   | -       |
| `--opnsense-api-secret` / `$OPNSENSE_API_SECRET`         | The API secret for authenticating with the OPNsense API. | `string` | `true`   | -       |
| `--opnsense-allow-insecure` / `$OPNSENSE_ALLOW_INSECURE` | Allow insecure TLS connections to the OPNsense API.      | `bool`   | `false`  | `false` |

### OPNsense Retry Configuration

| Flag / Environment                                 | Description                                                         | Type    | Required | Default |
| -------------------------------------------------- | ------------------------------------------------------------------- | ------- | -------- | ------- |
| `--opnsense-min-backoff` / `$OPNSENSE_MIN_BACKOFF` | Minimum backoff time in seconds for retrying OPNsense API requests. | `int64` | `false`  | `120`   |
| `--opnsense-max-backoff` / `$OPNSENSE_MAX_BACKOFF` | Maximum backoff time in seconds for retrying OPNsense API requests. | `int64` | `false`  | `120`   |
| `--opnsense-max-retries` / `$OPNSENSE_MAX_RETRIES` | Maximum retries for OPNsense API requests.                          | `int64` | `false`  | `120`   |

### Domain Filtering

These flags match the upstream [external-dns domain filtering configuration](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/flags.md).

| Flag / Environment                                     | Description                          | Type       | Required | Default |
| ------------------------------------------------------ | ------------------------------------ | ---------- | -------- | ------- |
| `--domain-filter` / `$DOMAIN_FILTER`                   | List of domain include filters.      | `string[]` | `false`  | -       |
| `--exclude-domains` / `$EXCLUDE_DOMAINS`               | List of domain exclude filters.      | `string[]` | `false`  | -       |
| `--regex-domain-filter` / `$REGEX_DOMAIN_FILTER`       | Domain include filter in regex form. | `string`   | `false`  | -       |
| `--regex-domain-exclusion` / `$REGEX_DOMAIN_EXCLUSION` | Domain exclude filter in regex form. | `string`   | `false`  | -       |

<!--- clidocsstop -->

## Related Projects

- [external-dns](https://github.com/kubernetes-sigs/external-dns) - The core library that enables this.
- [opnsense-go](https://github.com/browningluke/opnsense-go) - ~OPNsense API client library that is used in this webhook provider.~ I have initially used this however for each CRUD it was reconfiguring services, so I had to rewrite a similar implementation.
- [crutonjohn/external-dns-opnsense-webhook](https://github.com/crutonjohn/external-dns-opnsense-webhook) - Initial implementation that inspered this project.
