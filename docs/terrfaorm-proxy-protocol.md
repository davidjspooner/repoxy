## What Is the Provider Network Mirror Protocol?

### Terraform

* This protocol provides an **alternative installation source** for Terraform providers, independent of origin registries.
* It allows you to configure Terraform CLI to download provider binaries from a **single internal network mirror**, even if those providers originate from different hostnames or registries. ([HashiCorp Developer][1], [opentofu.org][2]).
* It’s **not** meant for implementing an origin registry; for that you’d use the provider *registry* protocol instead ([HashiCorp Developer][1]).

### OpenTofu

* Works the same way: it's an optional mechanism to serve providers from a custom mirror, which can host providers from any registry hostname. ([opentofu.org][3]).
* Also distinct from provider registry protocol, which is used when implementing origin registries. ([opentofu.org][3]).

---

## How It Works

1. **Configuration in CLI**

   In your CLI configuration file (`.terraformrc`, `terraform.rc`, or `tofu.rc`), you enable a network mirror like so:

   #### Terraform example:

   ```hcl
   provider_installation {
     network_mirror {
       url = "https://terraform.example.com"
     }
   }
   ```

   The CLI first calls `https://terraform.example.com/.well-known/terraform.json` and reads the `providers.v1` key to discover the actual `/v1/providers/` base path. ([opentofu.org][2], [HashiCorp Developer][1])

   #### OpenTofu example:

   ```hcl
   provider_installation {
     network_mirror {
       url = "https://tofu.example.com"
     }
   }
   ```

   Works in exactly the same way. ([opentofu.org][3])

2. **URL Constructs**

   After well-known discovery, Terraform/OpenTofu builds requests relative to the advertised base URL. A request for a provider might look like:

   ```
   https://mirror.example.com/:hostname/:namespace/:type/index.json
   ```

   Here, `:hostname` refers to the provider’s original registry host, not the mirror’s host. ([opentofu.org][3], [HashiCorp Developer][1])

3. **Fetching Providers**

   * The client retrieves an `index.json` listing available versions, then fetches the specific provider binary (`.zip`) for the appropriate OS/architecture.
   * A mirror doesn't need to host *all* platforms—just what your environment requires. ([HashiCorp Developer][1])

4. **Mirroring Tools**

   * **Terraform CLI** includes the `terraform providers mirror` command to generate a filesystem-based mirror that includes the proper `.zip` files and JSON indices. This directory can be deployed as a static mirror (e.g., HTTP server) ([HashiCorp Developer][4]).
   * **OpenTofu** has a similar command: `tofu providers mirror`, supporting the same logic and optional platform targeting ([opentofu.org][5]).

5. **Security / Signing Considerations**

   * When using a mirror, Terraform/OpenTofu trusts that mirror implicitly. It does not verify upstream provider signatures via the mirror. You only get TLS-based integrity checks on the mirror itself. ([HashiCorp Discuss][6]).
   * If provider binary signature verification is important, you need to run a **private provider registry** implementing the provider registry protocol, rather than using a mirror.

---

## Official Documentation Links

* **Terraform — Provider Network Mirror Protocol (Reference)**
  Official protocol spec, URL structure, and usage.
  ([HashiCorp Developer][1], [opentofu.org][3])

* **Terraform CLI — `terraform providers mirror` Command**
  How to use CLI to generate filesystem mirrors.
  ([HashiCorp Developer][4], [opentofu.org][2])

* **OpenTofu — Provider Network Mirror Protocol (Internals)**
  OpenTofu’s version of the spec with matching behavior.
  ([opentofu.org][7])

* **OpenTofu CLI — `tofu providers mirror` Command**
  Equivalent mirror generation command in OpenTofu.
  ([opentofu.org][5])

---

## Repoxy Implementation Notes

- The `/v1/providers/...` endpoints follow the upstream registry schema and are advertised via `/.well-known/terraform.json` so Terraform/OpenTofu only need the mirror hostname in `.terraformrc`/`.tofurc`.
- Provider metadata (`versions.json`, `manifest.json`, and per-platform download metadata) is cached in Repoxy’s `refs/` storage, so repeat inits do not hit upstream registries unless the cache is cleared.
- Provider archives (`.zip`) are saved once under `packages/` and streamed locally on subsequent installs, matching the package caching requirement in `requirements/framework/storage-heirachy.md`.
- Cached download metadata is rewritten on-the-fly so the client always receives a mirror-local `download_url`, even when the JSON body was fetched earlier.

### Summary (Concise)

| Aspect             | Terraform                                          | OpenTofu                |
| ------------------ | -------------------------------------------------- | ----------------------- |
| Purpose            | Mirror providers via network                       | Same                    |
| Configuration      | `network_mirror { url = ... }`                     | Same                    |
| CLI Mirror Command | `terraform providers mirror`                       | `tofu providers mirror` |
| Signing Integrity  | Mirror implicitly trusted; no signature validation | Same                    |
| Use Case           | Secure internal distribution without internet      | Same                    |

---

Let me know if you want help setting up a mirror server, integrating with Artifactory/Nexus, or anything else.

[1]: https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol?utm_source=chatgpt.com "Provider network mirror protocol reference | Terraform"
[2]: https://opentofu.org/docs/cli/config/config-file/?utm_source=chatgpt.com "CLI Configuration File ( .tofurc or tofu.rc )"
[3]: https://opentofu.org/docs/internals/provider-network-mirror-protocol/?utm_source=chatgpt.com "Provider Network Mirror Protocol"
[4]: https://developer.hashicorp.com/terraform/cli/commands/providers/mirror?utm_source=chatgpt.com "terraform providers mirror command reference"
[5]: https://opentofu.org/docs/cli/commands/providers/mirror/?utm_source=chatgpt.com "Command: providers mirror"
[6]: https://discuss.hashicorp.com/t/signature-verification-with-provider-network-mirrors/47478?utm_source=chatgpt.com "Signature verification with provider network mirrors"
[7]: https://opentofu.org/docs/internals/provider-meta/?utm_source=chatgpt.com "Provider Metadata"
