# Contributing to IBN Network

Cảm ơn bạn đã quan tâm đến việc đóng góp cho IBN Network! Tài liệu này cung cấp hướng dẫn về cách bạn có thể đóng góp cho dự án.

## 📋 Mục Lục

- [Code of Conduct](#code-of-conduct)
- [Cách Đóng Góp](#cách-đóng-góp)
- [Quy Trình Phát Triển](#quy-trình-phát-triển)
- [Tiêu Chuẩn Code](#tiêu-chuẩn-code)
- [Testing](#testing)
- [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)

---

## Code of Conduct

Dự án này tuân thủ [Code of Conduct](CODE_OF_CONDUCT.md). Bằng cách tham gia, bạn đồng ý tuân thủ các quy tắc này.

---

## Cách Đóng Góp

### Báo Cáo Lỗi (Bug Reports)

Nếu bạn phát hiện lỗi, vui lòng tạo một issue với:

- **Tiêu đề rõ ràng** mô tả vấn đề
- **Mô tả chi tiết** về lỗi
- **Các bước để reproduce** lỗi
- **Expected behavior** vs **Actual behavior**
- **Environment** (OS, Go version, Node version, etc.)
- **Logs/Error messages** (nếu có)

**Template:**
```markdown
## Mô Tả
[Miêu tả ngắn gọn về lỗi]

## Các Bước Reproduce
1. ...
2. ...
3. ...

## Expected Behavior
[Miêu tả hành vi mong đợi]

## Actual Behavior
[Miêu tả hành vi thực tế]

## Environment
- OS: [e.g., Ubuntu 22.04]
- Go Version: [e.g., 1.24.0]
- Node Version: [e.g., 20.10.0]

## Logs
```
[Paste logs here]
```
```

### Đề Xuất Tính Năng (Feature Requests)

Để đề xuất tính năng mới:

- **Tiêu đề rõ ràng** mô tả tính năng
- **Mô tả chi tiết** về tính năng và use case
- **Lý do** tại sao tính năng này hữu ích
- **Ví dụ** về cách sử dụng (nếu có)

---

## Quy Trình Phát Triển

### 1. Fork Repository

```bash
# Fork repository trên GitHub
# Clone fork của bạn
git clone https://github.com/YOUR_USERNAME/luongthiOlympic.git
cd ibn
```

### 2. Tạo Branch

```bash
# Tạo branch mới từ main
git checkout -b feature/your-feature-name
# hoặc
git checkout -b fix/your-bug-fix
```

**Naming Convention:**
- `feature/` - Tính năng mới
- `fix/` - Sửa lỗi
- `docs/` - Cập nhật tài liệu
- `refactor/` - Refactor code
- `test/` - Thêm tests

### 3. Phát Triển

- Viết code theo [Tiêu Chuẩn Code](#tiêu-chuẩn-code)
- Thêm tests cho code mới
- Cập nhật documentation nếu cần
- Đảm bảo tất cả tests pass

### 4. Commit Changes

```bash
# Stage changes
git add .

# Commit với message rõ ràng
git commit -m "feat: add new feature description"
```

Xem [Commit Messages](#commit-messages) để biết format.

### 5. Push và Tạo Pull Request

```bash
# Push branch lên fork
git push origin feature/your-feature-name

# Tạo Pull Request trên GitHub
```

---

## Tiêu Chuẩn Code

### Go Code Style

- **Format:** Sử dụng `gofmt` hoặc `goimports`
- **Linting:** Tuân thủ `golangci-lint` rules
- **Naming:**
  - Exported: `PascalCase`
  - Private: `camelCase`
  - Constants: `UPPER_SNAKE_CASE`
- **Error Handling:** Luôn kiểm tra và return errors
- **Context:** Truyền `context.Context` cho async operations

**Example:**
```go
// Good
func (s *Service) GetUser(ctx context.Context, userID string) (*User, error) {
    if userID == "" {
        return nil, fmt.Errorf("userID cannot be empty")
    }
    // ...
}

// Bad
func GetUser(id string) *User {
    // Missing error handling, no context
    return user
}
```

### TypeScript/React Code Style

- **Format:** Sử dụng Prettier
- **Linting:** Tuân thủ ESLint rules
- **Naming:**
  - Components: `PascalCase`
  - Functions/Variables: `camelCase`
  - Constants: `UPPER_SNAKE_CASE`
- **Type Safety:** Sử dụng TypeScript types, tránh `any`

**Example:**
```typescript
// Good
interface UserProps {
  userId: string;
  name: string;
}

export const UserCard: React.FC<UserProps> = ({ userId, name }) => {
  // ...
};

// Bad
export const UserCard = (props: any) => {
  // Missing types
};
```

### Architecture Rules

**QUAN TRỌNG:** Tuân thủ kiến trúc layered:

```
Handler → Service → Repository → Infrastructure
```

- ❌ **KHÔNG** skip layers (Handler → Database)
- ❌ **KHÔNG** đặt business logic trong Handler hoặc Repository
- ✅ Business logic PHẢI ở Service layer
- ✅ Handler chỉ xử lý HTTP request/response

**Example:**
```go
// ✅ Good: Handler → Service → Repository
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    user, err := h.service.CreateUser(r.Context(), &req)
    // ...
}

// ❌ Bad: Handler → Repository (skip Service)
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.repository.CreateUser(r.Context(), &req)
    // ...
}
```

---

## Testing

### Go Tests

- **Location:** `*_test.go` trong cùng package
- **Coverage:** Target >80%
- **Naming:** `TestFunctionName`
- **Table-driven tests:** Sử dụng cho multiple test cases

**Example:**
```go
func TestService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        req     *CreateUserRequest
        wantErr bool
    }{
        {
            name: "valid user",
            req: &CreateUserRequest{
                Email: "test@example.com",
                Name:  "Test User",
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            req: &CreateUserRequest{
                Email: "invalid",
                Name:  "Test User",
            },
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

### Frontend Tests

- **Location:** `*.test.tsx` hoặc `*.spec.tsx`
- **Framework:** Vitest hoặc Jest
- **Coverage:** Target >80%

**Example:**
```typescript
import { render, screen } from '@testing-library/react';
import { UserCard } from './UserCard';

describe('UserCard', () => {
  it('renders user name', () => {
    render(<UserCard userId="1" name="Test User" />);
    expect(screen.getByText('Test User')).toBeInTheDocument();
  });
});
```

### Chạy Tests

```bash
# Go tests
cd backend
go test ./... -v -cover

# Frontend tests
cd frontend
npm test

# Chaincode tests
cd teaTraceCC
npm test
```

---

## Commit Messages

Sử dụng [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: Tính năng mới
- `fix`: Sửa lỗi
- `docs`: Cập nhật tài liệu
- `style`: Formatting, không ảnh hưởng code
- `refactor`: Refactor code
- `test`: Thêm tests
- `chore`: Maintenance tasks

### Examples

```bash
# Feature
git commit -m "feat(auth): add JWT refresh token support"

# Bug fix
git commit -m "fix(chaincode): fix hash verification logic"

# Documentation
git commit -m "docs: update API documentation"

# Refactor
git commit -m "refactor(service): extract common validation logic"
```

---

## Pull Request Process

### Checklist Trước Khi Tạo PR

- [ ] Code tuân thủ [Tiêu Chuẩn Code](#tiêu-chuẩn-code)
- [ ] Tất cả tests pass
- [ ] Coverage >80% cho code mới
- [ ] Documentation đã được cập nhật
- [ ] Commit messages theo [Conventional Commits](#commit-messages)
- [ ] Không có merge conflicts với `main`
- [ ] Đã test locally

### PR Template

Khi tạo PR, vui lòng điền đầy đủ thông tin:

```markdown
## Mô Tả
[Miêu tả ngắn gọn về thay đổi]

## Loại Thay Đổi
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Cách Test
[Miêu tả cách test thay đổi này]

## Checklist
- [ ] Code tuân thủ style guide
- [ ] Tests đã được thêm/cập nhật
- [ ] Documentation đã được cập nhật
- [ ] Không có breaking changes (hoặc đã document)

## Screenshots (nếu có)
[Thêm screenshots nếu là UI changes]
```

### Review Process

1. **Automated Checks:** CI/CD sẽ chạy tests và linting
2. **Code Review:** Ít nhất 1 reviewer phải approve
3. **Merge:** Sau khi approved, PR sẽ được merge vào `main`

---

## Cấu Trúc Dự Án

```
ibn/
├── backend/              # Go backend API
│   ├── internal/
│   │   ├── handlers/     # HTTP handlers
│   │   ├── services/     # Business logic
│   │   └── infrastructure/ # Database, cache, gateway
│   └── cmd/server/       # Entry point
├── api-gateway/          # API Gateway service
├── frontend/             # React frontend
├── teaTraceCC/           # Chaincode
└── docs/                 # Documentation
```

---

## Tài Liệu Tham Khảo

- [Backend Architecture](docs/v1.0.1/backend.md)
- [API Gateway](docs/v1.0.1/gateway.md)
- [Network Architecture](docs/v1.0.1/network.md)
- [Chaincode Documentation](teaTraceCC/README.md)

---

## Câu Hỏi?

Nếu bạn có câu hỏi, vui lòng:

1. Tạo một [Discussion](https://github.com/Callmeduobgne/luongthiOlympic/discussions)
2. Tạo một issue với label `question`
3. Liên hệ maintainers

---

**Cảm ơn bạn đã đóng góp cho IBN Network! 🎉**


