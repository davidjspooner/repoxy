## Version 0.3 – Control Plane Surface
1. **API requirements:** define management API contracts (read-only in MVP) for listing repos, cache entries, and metrics snapshots.
2. **REST implementation:** build the Go handlers and authentication model for the management API, following the configuration guardrails.  
3. **UI foundation:** specify UI requirements and ship a lightweight React frontend (Material UI) that consumes the REST API for basic dashboards (status only, no writes yet).

## Version 0.4 – Alpha testing and first round rework
1. **Internal alpha program:** exercise the proxy in-house across Docker/Terraform workloads, capturing logs, metrics, and qualitative feedback.
2. **Bug triage:** categorize alpha findings (P0/P1/etc.), feed them into the backlog, and document temporary workarounds in `requirements/`.
3. **Stabilization fixes:** address alpha blockers (crashers, data-corruption risks, major UX gaps) and add regression tests covering each fix.

## Version 0.5 – Beta testing and second round rework
1. **Limited external beta:** onboard a small set of partner teams, provide upgrade notes, and monitor their telemetry for edge cases not seen in alpha.
2. **Performance tuning:** run load/perf tests, profile storage hot paths, and optimize cache access/eviction policies where required.
3. **Upgrade and rollback tooling:** script data migrations (if any), ship rollback instructions, and update `conf/SETUP.md` with beta-specific caveats.

## Version 0.6 – Packaging & Edge Features
1. **Distribution:** create build artifacts (Docker image, systemd unit sample) and document deployment steps for Linux hosts.
2. **End-to-end validation:** integration tests covering Docker + Terraform flows with upstream auth, cache hits/misses, and metrics verification.
3. **Security review:** threat model the read-only proxy, ensure TLS guidance is in `conf/SETUP.md`, and add configuration validation for misconfigured upstreams.
4. **Extensibility hooks:** document how to register new repo types and provide template packages/tests so third parties can extend Repoxy safely.

## Version 1.0 – Minimum Viable Product
1. **Docs freeze:** finalize `README.md`, `requirements/`, and `conf/SETUP.md` so operators can deploy without tribal knowledge; tag v1.0.
2. **Publish:** make the GitHub project public and cut the first release (ZIP + Docker image).

## Version 1.1 – Post-MVP Extensions
1. **Client auth:** enforce authentication/authorization for the UI and per-repo read/write access, covering both humans and automation.
2. **Writable local repositories (no upstream):** design and implement mode (b) from the requirements (local-only origin) including storage layout, locking, and auth rules.
3. **Cache governance:** add TTL/eviction policies aligned with `requirements/framework/storage-heirachy.md` plus CLI commands to refresh or purge specific refs/blobs.
4. **Signing of files:** consider if signing live inside or before repoxy for local terraform artifacts

## Version 1.2 – More Post-MVP Extensions
1. **Debian repo:** add a Debian repository type (Packages/Sources indices, .deb caching) leveraging the shared storage layout.  
2. **General files:** support a generic file/artifact mirror for workflows that do not fit OCI or Terraform protocols, reusing the same caching primitives.
