# Contributing to IBN Network

First off, thanks for taking the time to contribute! 🎉

The following is a set of guidelines for contributing to IBN Network. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Styleguides](#styleguides)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

This section guides you through submitting a bug report for IBN Network. Following these guidelines helps maintainers and the community understand your report, reproduce the behavior, and find related reports.

**Before Submitting a Bug Report:**

- Check if the bug has already been reported in [Issues](https://github.com/your-repo/issues)
- Check the [SETUP_GUIDE.md](SETUP_GUIDE.md) for common issues and troubleshooting

**When Submitting a Bug Report:**

- **Use a clear and descriptive title** for the issue to identify the problem
- **Describe the exact steps which reproduce the problem** in as many details as possible
- **Provide specific examples** to demonstrate the steps
- **Include environment details:**
  - OS and version
  - Docker version
  - Go version (if backend related)
  - Node.js version (if frontend/chaincode related)
- **Include relevant logs** (sanitize sensitive data like API keys, passwords, certificates)
- **Describe the expected behavior** and what actually happened
- **Include screenshots** if applicable

**Example Bug Report:**

```
Title: Chaincode deployment fails with "broken pipe" error

Environment:
- OS: Ubuntu 22.04
- Docker: 24.0.7
- Script: scripts/setup.sh (option 5)

Steps to Reproduce:
1. Run Fresh Setup (option 1)
2. Wait for network to start
3. Deploy chaincode (option 5)
4. Error occurs at Step 2/7: Installing chaincode

Expected: Chaincode installs successfully
Actual: Error: "write unix @->/run/docker.sock: write: broken pipe"

Logs:
[Include relevant log snippets]
```

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for IBN Network, including completely new features and minor improvements to existing functionality.

- **Use a clear and descriptive title** for the issue to identify the suggestion
- **Provide a step-by-step description of the suggested enhancement** in as many details as possible
- **Explain why this enhancement would be useful** to most IBN Network users
- **Describe the expected behavior** after the enhancement
- **Consider implementation complexity** and potential breaking changes

### Pull Requests

- Fill in the required template
- Do not include issue numbers in the PR title
- Include screenshots and animated GIFs in your pull request whenever possible
- Follow the style guides
- Ensure all tests pass
- Update documentation if needed

## Development Setup

### Prerequisites

- **Docker** 20.10+ and Docker Compose 2.0+
- **Go** 1.24.0+ (for backend development)
- **Node.js** 20.x+ (for frontend/chaincode development)
- **Git** 2.30+

### Initial Setup

1. Clone the repository:
```bash
git clone https://github.com/your-repo/ibn-network.git
cd ibn-network
```

2. Run the setup script:
```bash
./scripts/setup.sh
# Choose option 1: Fresh Setup
```

3. Verify the setup:
```bash
docker compose ps
# All services should be running
```

### Project Structure

```
ibn-network/
├── backend/              # Go backend API
│   ├── internal/
│   │   ├── handlers/     # HTTP handlers
│   │   ├── services/     # Business logic
│   │   └── infrastructure/ # Database, cache, gateway
│   └── cmd/server/       # Application entry point
├── frontend/             # React + TypeScript UI
├── api-gateway/          # Nginx + Fabric Gateway SDK
├── teaTraceCC/           # Hyperledger Fabric chaincode
├── core/                 # Fabric network configuration
│   ├── organizations/    # Crypto material
│   ├── configtx/         # Channel configuration
│   └── system-genesis-block/ # Genesis block
├── scripts/              # Automation scripts
└── docs/                 # Documentation
```

## Styleguides

### Git Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line
- Use conventional commit format when possible:
  - `feat:` for new features
  - `fix:` for bug fixes
  - `docs:` for documentation changes
  - `refactor:` for code refactoring
  - `test:` for test additions/changes
  - `chore:` for maintenance tasks

**Examples:**
```
feat: Add Admin MSP support for lifecycle queries

fix: Resolve TLS certificate mismatch in genesis block

docs: Update SETUP_GUIDE.md with troubleshooting steps
```

### Go Styleguide

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `go fmt` to format your code
- Use `golangci-lint` for linting (if configured)
- Follow the project's architecture patterns:
  - **Layered Architecture:** Handler → Service → Repository → Infrastructure
  - **Domain-Driven Design:** Organize services by domain
  - **Dependency Injection:** Use interfaces for dependencies

**Code Organization:**
- Handlers: HTTP request/response handling only
- Services: Business logic and orchestration
- Repositories: Data access layer
- Infrastructure: External dependencies (DB, cache, gateway)

**Naming Conventions:**
- Exported functions/types: `PascalCase`
- Private functions/types: `camelCase`
- Interfaces: End with `er` or descriptive name (e.g., `AuthService`)
- Constants: `UPPER_SNAKE_CASE`

**Example:**
```go
// Handler
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // Parse request
    // Validate input
    // Call service
    // Return response
}

// Service
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    // Business logic
    // Call repository
    // Return result
}
```

### TypeScript/React Styleguide

- Use functional components with hooks
- Use `const` over `let` and `var`
- Use strict type checking
- Follow React best practices:
  - Extract reusable components
  - Use custom hooks for shared logic
  - Optimize re-renders with `React.memo` when needed
  - Use TypeScript interfaces for props and state

**Component Structure:**
```typescript
interface ComponentProps {
  title: string;
  onAction: () => void;
}

export const Component: React.FC<ComponentProps> = ({ title, onAction }) => {
  // Hooks
  const [state, setState] = useState<string>('');
  
  // Effects
  useEffect(() => {
    // Side effects
  }, []);
  
  // Handlers
  const handleClick = () => {
    onAction();
  };
  
  // Render
  return (
    <div>
      <h1>{title}</h1>
      <button onClick={handleClick}>Action</button>
    </div>
  );
};
```

### Shell Script Styleguide

- Use `#!/bin/bash` shebang
- Use meaningful variable names
- Add comments for complex logic
- Use functions for reusable code
- Follow the project's script patterns (see `scripts/setup.sh`)

## Testing Guidelines

### Backend Testing

- Write unit tests for services and repositories
- Use table-driven tests for multiple scenarios
- Mock external dependencies (database, cache, gateway)
- Target >80% code coverage

**Example:**
```go
func TestAuthService_Login(t *testing.T) {
    tests := []struct {
        name    string
        req     *LoginRequest
        wantErr bool
    }{
        {
            name: "valid credentials",
            req:  &LoginRequest{Email: "test@example.com", Password: "password"},
            wantErr: false,
        },
        {
            name: "invalid credentials",
            req:  &LoginRequest{Email: "test@example.com", Password: "wrong"},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Frontend Testing

- Write unit tests for components and hooks
- Use React Testing Library for component testing
- Test user interactions and edge cases

### Integration Testing

- Test API endpoints with real database (use testcontainers)
- Test chaincode deployment and invocation
- Test end-to-end workflows

## Pull Request Process

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes:**
   - Write code following style guides
   - Add tests for new features
   - Update documentation if needed

3. **Commit your changes:**
   ```bash
   git commit -m "feat: Add your feature description"
   ```

4. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create a Pull Request:**
   - Fill in the PR template
   - Link related issues
   - Request review from maintainers

6. **Address review feedback:**
   - Make requested changes
   - Update PR description if needed
   - Re-request review when ready

## Additional Notes

### Issue and Pull Request Labels

This section lists the labels we use to help us track and manage issues and pull requests.

* `bug` - Issues that are bugs
* `enhancement` - Issues that are feature requests
* `documentation` - Issues or PRs related to documentation
* `good first issue` - Good for newcomers
* `blockchain` - Related to Hyperledger Fabric/chaincode
* `backend` - Related to Go backend API
* `frontend` - Related to React frontend
* `infrastructure` - Related to Docker, deployment, or infrastructure
* `security` - Security-related issues
* `performance` - Performance improvements
* `breaking-change` - Changes that break backward compatibility

### Security Considerations

- **Never commit sensitive data:**
  - API keys
  - Passwords
  - Private keys
  - Certificates (except test certificates)
- **Sanitize logs** before sharing
- **Use environment variables** for configuration
- **Follow security best practices** for blockchain operations

### Blockchain-Specific Guidelines

- **Chaincode changes:** Must be tested thoroughly before deployment
- **Network configuration:** Changes to `configtx.yaml` require network restart
- **Crypto material:** Never commit real certificates to repository
- **Genesis block:** Must be regenerated if crypto material changes

---

Thank you for contributing to IBN Network! 🚀
