# Alethic Instruction-Based State Machine (ISM) Core SDK for Golang

The Alethic ISM Core SDK is the foundational layer of the Alethic ISM project group. It provides the core processor and state management functionalities required to build specialized processors—including language-based processors—and to handle the majority of state input/output processing.

## Key Concepts
 * State Information: Handles state data for specific processing configurations. 
 * State Management: Maintains coherence in state columns and rows. 
 * Routing Management: Manages the NATS routing table and facilitates data subscription and transmission. 
 * ISM Data Model: Implements a subset of the ISM data model with some components still under development.

## Usage

To use this SDK in your project:

```go
import "github.com/quantumwake/alethic-ism-core-go"
```

## Development

### Prerequisites
- Go 1.24+

### Testing
Run the test suite:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

### Versioning
This project follows [Semantic Versioning](https://semver.org/). 

To create a new release:
```bash
# Bump patch version (v0.1.0 -> v0.1.1)
make version

# Bump minor version (v0.1.0 -> v0.2.0)
make version-minor

# Bump major version (v0.1.0 -> v1.0.0)
make version-major
```

These commands will automatically:
1. Create a new git tag with the appropriate version
2. Push the tag to GitHub, triggering a GitHub release

## Status 
This project is actively under development and remains in an experimental/prototype phase. Contributions and feedback are welcome as the project evolves.

## License
Alethic ISM is under a DUAL licensing model, please refer to [LICENSE.md](LICENSE.md).

**AGPL v3**  
Intended for academic, research, and nonprofit institutional use. As long as all derivative works are also open-sourced under the same license, you are free to use, modify, and distribute the software.

**Commercial License**
Intended for commercial use, including production deployments and proprietary applications. This license allows for closed-source derivative works and commercial distribution. Please contact us for more information.