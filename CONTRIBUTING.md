# Contributing to Go Licenser

Thank you for considering contributing to Go Licenser! We welcome contributions from the community.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in the [issues](https://github.com/dredfort42/go_licenser/issues)
2. If not, create a new issue with:
    - Clear description of the bug
    - Steps to reproduce
    - Expected vs actual behavior
    - Go version and OS information

### Suggesting Features

1. Check existing [issues](https://github.com/dredfort42/go_licenser/issues) for similar suggestions
2. Create a new issue with:
    - Clear description of the feature
    - Use cases and motivation
    - Potential implementation approach

### Submitting Pull Requests

1. **Fork the repository**
2. **Create a feature branch**
    ```bash
    git checkout -b feature/your-feature-name
    ```
3. **Make your changes**
    - Write clear, documented code
    - Add tests for new functionality
    - Ensure existing tests pass
4. **Test your changes**
    ```bash
    go test -v ./...
    go test -bench=. ./...
    go vet ./...
    ```
5. **Commit your changes**
    ```bash
    git commit -m "Add feature: your feature description"
    ```
6. **Push to your fork**
    ```bash
    git push origin feature/your-feature-name
    ```
7. **Create a Pull Request**

## Development Guidelines

### Code Style

-   Follow standard Go conventions
-   Run `gofmt` on your code
-   Use meaningful variable and function names
-   Add comments for exported functions and types

### Testing

-   Write tests for all new functionality
-   Maintain or improve code coverage
-   Include both positive and negative test cases
-   Add benchmarks for performance-critical code

### Documentation

-   Update README.md if needed
-   Add or update code comments
-   Update CHANGELOG.md for notable changes

## Code of Conduct

Please be respectful and constructive in all interactions. We aim to create a welcoming environment for all contributors.

## Questions?

If you have questions about contributing, feel free to:

-   Open an issue for discussion
-   Contact the maintainer: [dredfort.42@gmail.com](mailto:dredfort.42@gmail.com)

Thank you for contributing!
