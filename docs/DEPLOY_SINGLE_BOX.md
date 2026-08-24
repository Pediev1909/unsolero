# Single-box production deployment

This is the deployment path for launching with no infrastructure budget: the
whole stack on one small VPS, roughly €5 per month plus a domain.

It is a different shape from [PHASE_13_INFRASTRUCTURE_PLAN.md](../PHASE_13_INFRASTRUCTURE_PLAN.md),
which assumes managed databases with standbys, separate operator roles, and a
budget in the low hundreds per month. That plan is not wrong; it is written for
a funded team. Use this one to get real users, and move to that one when
revenue justifies it.

## What you give up

Be clear about this before you start, because it cannot be papered over:

- **No standby and no automatic failover.** If the host dies, the site is down
  until you restore onto a new one.
- **No point-in-time recovery.** You can recover to your last backup, not to
  an arbitrary moment.
- **Phase 13 gates G15 to G19 stay uncovered.** Those cover replica loss,
  dependency failover, PITR and media recovery drills. A single box cannot
  satisfy them.
- **Recovery is manual.** Budget an hour, and rehearse it before you need it.

What you do **not** give up: every production configuration check stays on.
PostgreSQL and Valkey run with real TLS inside the Docker network rather than
having their checks disabled. If a step below fails, fix the configuration —
do not relax a check to make startup succeed.

## What it costs

| Item | Monthly |
| --- | --- |
| VPS, 4 GB RAM minimum | about €5 |
| Cloudflare DNS, CDN and edge TLS | €0 |
| Cloudflare R2 for media and backups | €0 under 10 GB, no egress charge |
| SMTP free tier (Resend or Brevo) | €0 |
| Discord or Slack webhook for alerts | €0 |
| Domain | about €12 per year |

**Size the host for ClamAV.** Malware scanning is required in production and
its signature database is the largest single consumer of memory on the box.
The allocations in `compose.production.yaml` total roughly 3.9 GB, so 4 GB is
the floor and 8 GB is comfortable. If memory is tight, buy a bigger host rather
than turning scanning off.

## Before you touch the server

1. **Register the domain.** Certificates are issued over ACME against a real
   DNS record, so nothing works until this exists.
2. **Create the accounts:** VPS provider, Cloudflare (DNS and R2), an SMTP
   provider, and a Discord or Slack webhook.
3. **Create two R2 buckets:** one for media, one for backups. Keep them
   separate so a mistaken deletion cannot take both.
4. **Point DNS at the server** and let it propagate. Start with the Cloudflare
   proxy disabled (grey cloud) so the first certificate issuance is direct;
   turn it on afterwards and set SSL/TLS mode to **Full (strict)**.

## Deploy

```bash
# On the server, as a non-root user in the docker group
git clone <your-repo> /opt/unsolero
cd /opt/unsolero

cp .env.production.example .env
# Fill in every value. Generate the three 32-byte secrets with:
#   openssl rand -base64 32 | tr -d '='
# The padding must be stripped: these decode as raw base64, so an unmodified
# `openssl rand -base64 32` value is rejected at startup.
$EDITOR .env

# Issue the TLS material for the in-network database and rate-limit store.
./scripts/generate-internal-tls.sh

docker compose -f compose.yaml -f compose.production.yaml \
  --profile production up -d --build
```

Migrations run automatically as a dependency of the API; the API does not start
until they complete.

The first start is slow because ClamAV downloads its signature database. Until
that finishes its health check fails and the API waits, which is intended — the
API verifies the scanner answers before it accepts traffic.

### Load the catalog

The database starts empty. There is no active recommendation policy until a
catalog exists, and until then recommendations correctly return nothing.

To load the fictional development fixture, useful for confirming the stack
works end to end:

```bash
docker compose -f compose.yaml -f compose.production.yaml run --rm seed
```

Everything it loads is explicitly fictional and must be removed before real
traffic. Real products go in through the admin interface, where the evidence
workflow requires a source for each fact before a product can be published.

In production the `seed` service refuses to run: the fixture publishes invented
products at the same status the public catalog serves, so loading it would put
invented prices in front of real visitors. Removing the fixture is always
allowed, and has its own service so the flag cannot be mistyped:

```bash
docker compose -f compose.yaml -f compose.production.yaml \
  --profile production --profile tools run --rm seed-purge
```

Both profiles are required, not just `tools`. `tls-init` is a production-profile
service and `postgres` depends on it, so a command that omits `--profile
production` fails with "depends on undefined service" before it runs anything.

It reports how many products it unpublished, how many it deleted, and how many
an active recommendation policy is still holding. A held product stays in the
database but leaves the public catalog, which is the part that matters.

## Verify

```bash
# All services healthy
docker compose -f compose.yaml -f compose.production.yaml --profile production ps

# The public origin answers over TLS
curl -sS -o /dev/null -w '%{http_code} %{ssl_verify_result}\n' https://your-domain.example

# The API is NOT reachable from outside; this must fail to connect
curl -sS --max-time 5 http://your-server-ip:8080/ && echo "PROBLEM: API is exposed"

# Metrics are not public either
curl -sS -o /dev/null -w '%{http_code}\n' https://your-domain.example/metrics   # expect 404
```

Confirm the alert path actually delivers rather than assuming it does — an
alerting channel nobody has ever seen fire is not an alerting channel.

## Backups

Add to cron once the stack is running:

```cron
0 3 * * * cd /opt/unsolero && ./scripts/backup-to-r2.sh >> /var/log/unsolero-backup.log 2>&1
```

Set `BACKUP_AGE_RECIPIENT` in `.env` first. The dump contains every user
record; an unencrypted copy in third-party storage is a real exposure. The
script warns and continues without it, so the warning is easy to miss.

**Rehearse a restore.** Schedule it, do it against a scratch database, and
confirm the data is what you expect. An untested backup is a guess. This
matters more here than on managed infrastructure, because restore-from-backup
is your only recovery path.

## Updating

```bash
cd /opt/unsolero
git pull
# APP_VERSION must identify the build; startup rejects a placeholder.
sed -i "s/^APP_VERSION=.*/APP_VERSION=$(git rev-parse --short HEAD)/" .env
docker compose -f compose.yaml -f compose.production.yaml \
  --profile production up -d --build
# nginx inside web resolves the api hostname once, at startup, and caches the
# address. A backend-only change rebuilds api and leaves web untouched, so web
# keeps proxying to an address nothing answers on and every request becomes a
# 502 that Caddy reports as "no upstreams available". Recreating web after the
# fact costs seconds and prevents that outage.
docker compose -f compose.yaml -f compose.production.yaml \
  --profile production up -d --force-recreate web
```

Use `--profile production` alone here. Adding `--profile tools` to an `up`
starts the seed, purge, backup and reconciliation containers as well, because
`up` starts everything in every active profile. Tools belong on `run`, which
starts only the service named.

Migrations are forward-only. There is no automatic rollback, so take a backup
before deploying a release that changes the schema.

## When something will not start

The application fails closed by design, and the error names the setting.

| Message contains | Cause |
| --- | --- |
| `must enable PostgreSQL TLS` | `sslmode=require` missing from `DATABASE_URL` |
| `must use TLS` for the Redis URL | scheme is `redis://`, must be `rediss://` |
| `media scanner is not ready` | ClamAV is still downloading signatures, or unreachable |
| `must be raw-base64 for exactly 32 bytes` | trailing `=` padding not stripped from the secret |
| `APP_VERSION must identify the deployed release` | placeholder left in `.env` |
| `certificate signed by unknown authority` | `generate-internal-tls.sh` not run, or run after the containers started |
| `alert webhook configuration is incomplete or unsafe` | `ALERT_WEBHOOK_TOKEN` shorter than 32 characters, or the URL carries a query string |
| `PUBLIC_SITE_URL must use a public DNS hostname` | a reserved suffix such as `.local`, `.test`, `.example` or a bare IP |

If Valkey rejects the connection after regenerating certificates, restart the
API and worker: the trust bundle is read at startup.
