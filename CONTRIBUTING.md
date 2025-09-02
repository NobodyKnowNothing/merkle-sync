# Contributing to Universal MerkleSync

Thank you for your interest in contributing to Universal MerkleSync! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Architecture Overview](#architecture-overview)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your feature or bugfix
4. Make your changes
5. Test your changes thoroughly
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- PostgreSQL 15+ (for testing)
- MongoDB 7+ (for testing)

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/EdgeMerkleDB/universal-merkle-sync.git
   cd universal-merkle-sync
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Start the development environment:
   ```bash
   docker-compose up -d
   ```

4. Run tests:
   ```bash
   go test ./...
   ```

### Building from Source

```bash
# Build all components
go build ./cmd/server
go build ./cmd/postgresql-connector
go build ./cmd/mongodb-connector
go build ./cmd/edge-client
```

## Contributing Guidelines

### Code Style

- Follow Go conventions and best practices
- Use `gofmt` to format your code
- Use `golint` and `go vet` to check for issues
- Write comprehensive tests for new functionality
- Document public APIs with Go doc comments

### Commit Messages

Use clear, descriptive commit messages:

```
feat: add support for Redis connector
fix: resolve memory leak in Merkle tree construction
docs: update API documentation
test: add integration tests for PostgreSQL connector
```

### Testing

- Write unit tests for all new functionality
- Add integration tests for database connectors
- Ensure all tests pass before submitting a PR
- Aim for high test coverage

### Documentation

- Update relevant documentation for new features
- Add examples for new APIs
- Update README if needed
- Document any breaking changes

## Pull Request Process

1. **Create a Feature Branch**: Create a descriptive branch name from `main`
2. **Make Changes**: Implement your feature or bugfix
3. **Add Tests**: Ensure adequate test coverage
4. **Update Documentation**: Update relevant docs
5. **Run Tests**: Ensure all tests pass
6. **Submit PR**: Create a pull request with a clear description

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass
```

## Issue Reporting

When reporting issues, please include:

1. **Environment**: OS, Go version, database versions
2. **Steps to Reproduce**: Clear, numbered steps
3. **Expected Behavior**: What should happen
4. **Actual Behavior**: What actually happens
5. **Logs**: Relevant error messages or logs
6. **Additional Context**: Any other relevant information

## Architecture Overview

### Core Components

- **Core Library** (`core/`): Merkle tree construction and proof verification
- **gRPC Server** (`server/`): API server for MerkleSync operations
- **Connectors** (`connectors/`): Database-specific change capture
- **Edge Client** (`edge-client/`): Offline-first client library

### Adding New Connectors

To add support for a new database:

1. Create a new package in `connectors/`
2. Implement the connector interface
3. Add change data capture logic
4. Create integration tests
5. Update documentation

### Adding New Features

1. Design the feature following the existing architecture
2. Implement in the appropriate component
3. Add comprehensive tests
4. Update API documentation
5. Consider backward compatibility

## Community

- **Discussions**: Use GitHub Discussions for questions and ideas
- **Issues**: Use GitHub Issues for bug reports and feature requests
- **Pull Requests**: Use GitHub Pull Requests for code contributions

## License

By contributing to Universal MerkleSync, you agree that your contributions will be licensed under the Apache License 2.0.

## Questions?

If you have questions about contributing, please:

1. Check existing issues and discussions
2. Create a new discussion
3. Contact the maintainers

Thank you for contributing to Universal MerkleSync!
