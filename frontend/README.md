# Claude Relay Frontend

This is the admin dashboard frontend for the Claude Relay Go project, built with React 18, TypeScript, Vite, and TailwindCSS.

## Tech Stack

- **React 18** - UI library with hooks
- **TypeScript** - Type-safe JavaScript
- **Vite** - Fast build tool and dev server
- **TailwindCSS v4** - Utility-first CSS framework
- **React Router v6** - Client-side routing
- **Zustand** - Lightweight state management
- **Axios** - HTTP client with interceptors
- **Lucide React** - Icon library

## Project Structure

```
src/
├── components/
│   ├── common/          # Reusable UI components
│   │   ├── Button.tsx
│   │   ├── Input.tsx
│   │   ├── Modal.tsx
│   │   ├── Table.tsx
│   │   ├── Card.tsx
│   │   ├── Badge.tsx
│   │   ├── Loading.tsx
│   │   └── Alert.tsx
│   └── layout/          # Layout components
│       ├── Sidebar.tsx
│       ├── Header.tsx
│       └── Layout.tsx
├── pages/               # Page components
│   ├── Dashboard.tsx
│   └── Login.tsx
├── utils/               # Utility functions
│   ├── request.ts       # Axios instance with interceptors
│   ├── format.ts        # Date, number formatting
│   ├── validation.ts    # Form validation
│   ├── constants.ts     # App constants
│   └── cn.ts            # Classname utility
├── store/               # Zustand stores
│   ├── authStore.ts     # Authentication state
│   └── uiStore.ts       # UI state (sidebar, theme)
├── types/               # TypeScript types
│   ├── index.ts
│   └── admin.ts
├── api/                 # API client setup
│   └── index.ts
├── App.tsx              # Main app with routes
├── main.tsx             # Entry point
└── index.css            # TailwindCSS imports
```

## Getting Started

### Prerequisites

- Node.js 18+ and npm

### Installation

```bash
# Install dependencies
npm install

# Copy environment variables
cp .env.example .env

# Update .env with your API URL
```

### Development

```bash
# Start dev server (runs on http://localhost:3000)
npm run dev

# Type check
npm run type-check

# Lint
npm run lint

# Lint and fix
npm run lint:fix

# Format code
npm run format
```

### Build

```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

## Environment Variables

Create a `.env` file in the frontend directory (optional, if you want to override the default `/api/v1`):

```env
# Typical local dev: go through Vite proxy to backend
VITE_API_BASE_URL=/api/v1

# Or direct to backend if you prefer
# VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## Features Implemented

### Authentication
- Login page with form validation
- Token-based authentication
- Auto-logout on 401 errors
- Protected routes

### Layout
- Responsive sidebar navigation
- Collapsible sidebar with state persistence
- Header with user info and logout

### UI Components
- Button (primary, secondary, danger, outline, ghost variants)
- Input (with label, error, helper text)
- Modal (customizable size and content)
- Table (with sorting support)
- Card (with title, description, actions)
- Badge (success, warning, error, info variants)
- Loading (spinner with optional text)
- Alert (dismissible alerts with variants)

### State Management
- Auth store (token, user info, authentication status)
- UI store (sidebar collapsed state with persistence)

### Utilities
- Request interceptors (auto-add auth token)
- Response interceptors (auto-logout on 401)
- Date/time formatting
- Number formatting
- Token formatting
- Form validation helpers
- API key masking

## API Integration

The frontend is configured to proxy API requests to `http://localhost:8080/api`. All API endpoints are defined in `src/api/index.ts`:

- Authentication: `/admin/login`, `/admin/info`
- API Keys: `/api-keys` (CRUD)
- Codex Accounts: `/codex-accounts` (CRUD)
- Usage Records: `/usage-records` (read)
- Admin Management: `/admins` (CRUD)

## Code Quality

- TypeScript strict mode enabled
- ESLint with React and TypeScript rules
- Prettier for consistent formatting
- All components are typed
- Follows React best practices

## Next Steps

The following pages are marked as "Coming Soon" and will be implemented in subsequent tasks:
- API Keys management page
- Codex Accounts management page
- Usage Statistics page
- Admin management page

## License

This project is part of the Claude Relay Go system.
