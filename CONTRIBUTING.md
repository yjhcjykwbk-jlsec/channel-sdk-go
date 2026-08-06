# Contributing to Channel SDK for Go

Thank you for your interest in contributing to Channel SDK for Go.

## Before You Start

Before submitting a pull request:

- Search existing issues and pull requests to avoid duplicate work.
- For significant features, public API changes, or breaking changes, open an
  issue for discussion first.
- Do not include credentials, tokens, private user data, internal URLs,
  internal-only dependencies, or test assets without confirmed redistribution
  rights.

Report security vulnerabilities according to [SECURITY.md](SECURITY.md). Do
not disclose security vulnerabilities through public GitHub issues.

## Development Setup

Channel SDK for Go requires Go 1.18 or later.

```bash
git clone https://github.com/larksuite/channel-sdk-go.git
cd channel-sdk-go
go mod download
```

## Making Changes

Please follow these guidelines:

- Keep each pull request focused on one change.
- Preserve backward compatibility unless a breaking change has been discussed
  and approved in advance.
- Add or update tests for behavior changes and bug fixes.
- Update public documentation and examples when public APIs or behavior change.
- Add the following header to new Go source files:

  ```go
  // Copyright (c) 2026 Lark Technologies Pte. Ltd.
  // SPDX-License-Identifier: MIT
  ```

## Local Checks

Run the following checks before submitting a pull request:

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
go test ./examples/...
```

Real end-to-end tests require Feishu or Lark application credentials and are
not required for external contributors. After creating `.env` as described in
[the E2E testing guide](e2e/README.md), its configuration can be validated
without making network requests:

```bash
CHANNEL_E2E_DRY_RUN=1 go test -tags=e2e ./e2e -run TestChannelE2E -v
```

## Pull Requests

A pull request should:

- Clearly describe the problem and the proposed solution.
- Reference the related issue when applicable.
- Include tests for changed behavior.
- Pass all required CI checks.
- Avoid unrelated refactoring or formatting changes.

Maintainers may request changes before accepting a pull request.

## Contributor License Agreement

External contributors must complete the ByteDance Contributor License
Agreement (CLA).

After you open a pull request, the CLA check will report whether all
contributors have signed the agreement and will provide signing instructions
when required. A pull request cannot be merged until the CLA check passes.

By submitting a contribution, you confirm that you have the right to submit
the contribution and agree that it may be distributed under this project's
license.

## Code of Conduct

All contributors must follow our [Code of Conduct](CODE_OF_CONDUCT.md).
