# Security Policy

## Reporting

Please report suspected vulnerabilities privately through GitHub Security Advisories. Do not include secrets, private keys, real payment credentials, or personal data in a public issue.

## Trust boundaries

- A structurally valid receipt is not proof that a transaction finalized on-chain.
- Configure `ALLOWED_RESOURCE_HOSTS` in production; the empty default is intended for local evaluation.
- Place the service behind authenticated infrastructure if audit history contains sensitive resource URLs.
- Back up and protect the ledger volume. It is part of the replay-defense boundary.
- Terminate TLS at a trusted reverse proxy and restrict direct access to the application port.

## Supported versions

Security fixes are applied to the latest commit on `main`.
