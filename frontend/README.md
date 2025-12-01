# IBN Network Frontend

Frontend application cho hệ thống IBN Network - Blockchain-based Supply Chain Management System.

## 🚀 Quick Start

### Prerequisites
- Node.js >= 20.0.0 (hoặc >= 18.19.1 với warnings)
- npm hoặc yarn

### Installation

```bash
# Cài đặt dependencies
npm install

# Tạo file .env từ .env.example
cp .env.example .env

# Chạy development server
npm run dev
```

Ứng dụng sẽ chạy tại:
- **Development**: `http://localhost:5173` (với hot reload)
- **Production (Docker)**: `http://localhost:3001`

## 📁 Cấu trúc Project

```
frontend/
├── src/
│   ├── app/              # App-level setup
│   │   ├── router.tsx    # React Router configuration
│   │   └── stores/       # Zustand stores
│   │
│   ├── features/         # Feature-based modules
│   │   └── authentication/
│   │       ├── components/  # LoginForm, RegisterForm, etc.
│   │       ├── hooks/      # useAuth, useKeycloak
│   │       ├── services/   # authService
│   │       └── types/      # auth.types.ts
│   │
│   └── shared/           # Shared resources
│       ├── components/   # Reusable UI components
│       ├── hooks/        # Shared hooks
│       ├── utils/        # Utilities (api, errorHandler, etc.)
│       └── config/       # Configuration files
│
├── public/               # Static assets
└── package.json
```

## 🛠️ Technology Stack

- **React 19.2** - UI library
- **TypeScript 5.9** - Type safety
- **Vite 7.2** - Build tool
- **Tailwind CSS 4.1** - Styling
- **React Router 7.9** - Routing
- **React Query 5.90** - Server state management
- **Zustand 5.0** - Client state management
- **React Hook Form 7.66** - Form handling
- **Zod 4.1** - Schema validation
- **Axios 1.13** - HTTP client

## 🔐 Authentication

Frontend sử dụng JWT authentication với backend API Gateway.

**API Endpoints:**
- `POST /api/v1/auth/login` - Đăng nhập
- `POST /api/v1/auth/refresh` - Refresh token
- `GET /api/v1/auth/profile` - Lấy thông tin user

**Features:**
- Auto token refresh khi token hết hạn
- Protected routes
- Token storage trong localStorage

## 📝 Environment Variables

Tạo file `.env` với các biến sau:

```env
VITE_API_BASE_URL=http://localhost:9900
```

## 🎨 UI Components

### Base Components
- `Button` - Button với variants (primary, secondary, danger, ghost)
- `Input` - Input field với validation
- `Card` - Card component
- `QRCodeDisplay` - QR code display với download support

### Usage Example

```tsx
import { Button } from '@shared/components/ui/Button'
import { Input } from '@shared/components/ui/Input'

<Button variant="primary" size="md" isLoading={false}>
  Click me
</Button>

<Input label="Email" type="email" error={errors.email?.message} />
```

## 🧪 Development

```bash
# Development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint
npm run lint
```

## 📚 API Integration

API client được cấu hình với:
- Base URL: `http://localhost:9900` (có thể thay đổi qua env)
- Auto token injection
- Auto token refresh on 401
- Error handling với toast notifications

### QR Code API
- `GET /api/v1/qrcode/batches/{batchId}/base64` - Lấy QR code base64 cho batch
- `GET /api/v1/qrcode/packages/{packageId}/base64` - Lấy QR code base64 cho package
- `GET /api/v1/qrcode/transactions/{txId}` - Lấy QR code từ transaction ID

**Example:**

```typescript
import api from '@shared/utils/api'

// GET request
const response = await api.get('/api/v1/blocks/ibnchannel')

// POST request
const response = await api.post('/api/v1/auth/login', {
  email: 'user@example.com',
  password: 'password123'
})
```

## 🗺️ Routing

Routes được định nghĩa trong `src/app/router.tsx`:

- `/login` - Trang đăng nhập
- `/` - Home page (protected)

Protected routes tự động redirect về `/login` nếu chưa authenticated.

## 🔒 Security

- ✅ CSRF protection (sẵn sàng implement)
- ✅ XSS protection với DOMPurify (sẵn sàng implement)
- ✅ Input validation với Zod
- ✅ Secure token storage
- ✅ Auto token refresh

## 📖 Documentation

Xem thêm chi tiết trong:
- `/docs/v1.0.1/frontend.md` - Full architecture design

## 🐛 Troubleshooting

### Port already in use
Thay đổi port trong `vite.config.ts` hoặc dùng:
```bash
npm run dev -- --port 3001
```

### API connection issues
Kiểm tra:
1. Backend API Gateway đang chạy tại `http://localhost:9090`
2. CORS được cấu hình đúng trên backend
3. Environment variables trong `.env`

## 📄 License

Internal project - IBN Network

3. Environment variables trong `.env`

## 📄 License

Internal project - IBN Network
