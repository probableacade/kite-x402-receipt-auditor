# Contributing

1. Open an issue describing the behavior or policy change.
2. Add tests for new validation and replay scenarios.
3. Run `go test -race ./...` and `go vet ./...`.
4. Keep commits focused and explain security trade-offs in the pull request.

Please use synthetic addresses and transaction hashes in fixtures. Never commit credentials or real payment data.
