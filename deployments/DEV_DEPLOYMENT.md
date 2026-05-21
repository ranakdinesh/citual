# Citual Development Deployment

Citual dev deployment is designed for a GitHub self-hosted runner installed on the development server.

## Flow

```text
push to citual/main
  -> GitHub Actions
  -> self-hosted runner on dev server
  -> update /opt/citual-dev workspace
  -> docker compose build
  -> docker compose up -d
  -> health checks
```

The dev server workspace is expected to be monorepo-shaped because the current Docker build uses the workspace root as build context:

```text
/opt/citual-dev
  go.work
  go.work.sum
  citual
  citual-web
  spur-identity
  spur-engage
  spur-messaging
  spur-storage
  spur-template
```

## First-Time Server Setup

1. Install Docker and Docker Compose.
2. Install a GitHub self-hosted runner for the `ranakdinesh/citual` repository on the dev server.
3. Create the workspace:

```bash
mkdir -p /opt/citual-dev
git clone https://github.com/ranakdinesh/citual.git /opt/citual-dev/citual
cd /opt/citual-dev/citual
cp deployments/.env.example deployments/.env
```

4. Edit `/opt/citual-dev/citual/deployments/.env` on the server. Do not commit it.
5. Run the first deployment manually:

```bash
cd /opt/citual-dev/citual
deployments/scripts/update-workspace.sh
deployments/scripts/deploy-dev.sh
```

After this, pushes to `main` run `.github/workflows/deploy-dev.yml`.

## Scripts

- `deployments/scripts/update-workspace.sh`
  Updates or clones the workspace repositories.

- `deployments/scripts/deploy-dev.sh`
  Validates compose config, builds containers, starts the stack, records deployment state, and runs health checks.

- `deployments/scripts/healthcheck.sh`
  Checks backend and web health endpoints.

- `deployments/scripts/rollback-dev.sh`
  Resets repositories to the previous successful deployment commits and redeploys.

## Important Notes

- `deployments/.env` stays only on the server.
- Module repositories should be tagged independently.
- `citual` pins module versions and remains the deployment source of truth.
- The workflow uses `runs-on: self-hosted`; add stricter runner labels later if multiple self-hosted runners exist.
