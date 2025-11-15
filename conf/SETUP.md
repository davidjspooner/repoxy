# Repoxy Client Setup (Linux)

This guide explains how to point a Linux workstation or build agent at a remote Repoxy instance that is configured with the sample `conf/repoxy.yaml` (Docker Hub, GHCR, Terraform, and OpenTofu mirrors). Replace the hostnames and paths with your environment as needed.

---

## 1. Prerequisites

- Repoxy is running and reachable at `https://repoxy.example.com` (the sample config uses `https://wsl.home.dolbyn.com:8443` with a real TLS certificate).
- You can reach the server over HTTPS (firewall opened, DNS resolves, certificate chain trusted on the client).
- You have `sudo` on the Linux host to edit Docker configuration files.

Before changing package managers, test basic connectivity:

```bash
curl -I https://repoxy.example.com/v2/
curl -I https://repoxy.example.com/.well-known/terraform.json || true  # not implemented yet but should return 404/501, proving TLS works
```

If the TLS certificate is signed by your own CA, install that CA into `/usr/local/share/ca-certificates/repoxy.crt` and run `sudo update-ca-certificates`.

---

## 2. Docker Engine

Repoxy exposes two Docker repositories from the sample config:

- `dockerhub` — proxies Docker Hub (`https://registry-1.docker.io`) for every namespace (`*/*`).
- `github` — proxies GHCR (`https://ghcr.io`) but only for images under `davidjspooner/*`.

### 2.1 Mirror Docker Hub pulls transparently

1. Create or edit `/etc/docker/daemon.json`:

   ```json
   {
     "registry-mirrors": ["https://repoxy.example.com"]
   }
   ```

   Docker treats `registry-mirrors` as substitutes for Docker Hub only, so other registries remain untouched.

2. Restart Docker:

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl restart docker
   ```

3. Pull an image to confirm traffic goes through Repoxy:

   ```bash
   docker pull library/alpine:latest
   journalctl -u docker --no-pager | grep repoxy.example.com
   ```

### 2.2 Pull GHCR images via Repoxy

Docker does not support transparent mirrors for arbitrary registries, so reference the Repoxy hostname explicitly:

```bash
docker pull repoxy.example.com/davidjspooner/my-image:tag
```

- The path after the hostname (`davidjspooner/my-image`) matches the `mappings` entry `davidjspooner/*`, so Repoxy proxies the request to `ghcr.io/davidjspooner/my-image`.
- Tag images accordingly in CI/CD (`repoxy.example.com/davidjspooner/app:sha`).

If GHCR requires authentication, run `docker login repoxy.example.com` (Repoxy will forward credentials upstream once authentication middleware is implemented).

---

## 3. Terraform CLI (HashiCorp)

The `terraform-hashicorp` repo mirrors `https://registry.terraform.io` under `/v1/providers/hashicorp/...`. Configure the Terraform CLI to fetch providers from Repoxy:

1. Edit `~/.terraformrc` (or `~/.config/terraformrc`):

   ```hcl
   provider_installation {
     network_mirror {
       url = "https://repoxy.example.com"
     }
     direct {
       exclude = ["registry.terraform.io/hashicorp/*"]
     }
   }
   ```

   - Terraform automatically requests `https://repoxy.example.com/.well-known/terraform.json` and follows the `providers.v1` URL (Repoxy responds with `/v1/providers/`), so no manual path rewriting is required.
   - The mirror URL must be HTTPS. Repoxy redirects clients to `/v1/providers/*` internally; you only need to point Terraform at the origin hostname.
   - The `direct` block excludes the namespaces handled by the mirror, ensuring Terraform does not fall back to the public registry for those providers.

2. Test with `terraform init` inside any project that depends on HashiCorp providers.

3. Inspect `/home/<user>/.terraform.d/plugin-cache` or run `tfenv`/`terraform` with `TF_LOG=DEBUG` to confirm downloads hit Repoxy.

---

## 4. OpenTofu CLI

The `opentofu-registry` repo proxies `https://registry.opentofu.org` for the namespace `opentofu/*`. Configure OpenTofu’s mirror (`~/.tofurc` or `~/.config/tofurc`):

```hcl
provider_installation {
  network_mirror {
    url = "https://repoxy.example.com"
  }
  direct {
    exclude = ["registry.opentofu.org/opentofu/*"]
  }
}
```

Run `tofu init` to verify provider downloads use the mirror. OpenTofu performs the same `.well-known/terraform.json` discovery flow, so the configuration is nearly identical.

---

## 5. Troubleshooting Tips

- **Confirm routing:** `curl -H 'Host: repoxy.example.com' https://repoxy.example.com/v2/` should show `Docker-Distribution-API-Version`.
- **Check Repoxy logs:** `journalctl -u repoxy` (or wherever you run it) to confirm incoming traffic.
- **HTTP status 502:** usually indicates Repoxy cannot reach the upstream registry—verify outbound internet access from the proxy host.
- **Provider cache misses:** Delete the mirror cache directory (`/var/lib/repoxy/type/tf/proxies/<name>/refs`) if stale manifests cause issues, then retry.

Following these steps routes Docker Hub pulls, GHCR images (under `davidjspooner/*`), and Terraform/OpenTofu provider downloads through your Repoxy deployment for consistent auditing and caching.
