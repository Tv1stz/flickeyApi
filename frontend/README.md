# Flickey Frontend (Svelte 5 + SvelteKit 2 + TypeScript + Tailwind CSS)

A standalone, production-ready frontend for the Flickey rental marketplace.

## Features

- **Svelte 5 & SvelteKit 2**: Built entirely using modern Svelte 5 Runes (`$state`, `$derived`, `$effect`, `$props`, `$bindable`, snippets) and `$app/state`.
- **TypeScript**: Strict type-safety across all components, API requests, and data models matching the backend contracts.
- **shadcn-svelte UI**: Clean, accessible component system styled with Tailwind CSS tokens and custom CSS variables.
- **Direct S3 Upload**: Fast, presigned direct photo uploads with client validation, drag-and-drop, and magic bytes verification.
- **Single-Identity Auth Model**: Seamless switching between **Guest** and **Host** contexts on a single account.
- **6-Step Listing Creation Wizard**: Intuitive multi-step listing creator with live validation and atomic UUIDv4 idempotency key submission.
- **Complete Portability**: 100% decoupled from the backend source code. Can be moved to any other repository or server.

---

## Directory Structure

```text
frontend/
├── src/
│   ├── lib/
│   │   ├── api/          # Typed API modules (client, auth, listings, media, amenities)
│   │   ├── auth/         # Svelte 5 Runes Auth State Class ($state, $derived)
│   │   ├── components/
│   │   │   ├── ui/       # shadcn-svelte UI primitives (Button, Card, Dialog, Toast, etc.)
│   │   │   ├── layout/   # Header, Context Switcher, Footer
│   │   │   ├── auth/     # AuthDialog (OTP SMS & Dev Quick Login)
│   │   │   ├── listings/ # Public Catalog, ListingCard, FilterBar, Gallery
│   │   │   └── wizard/   # 6-Step Multi-step Listing Creation Wizard
│   │   ├── types/        # TypeScript interfaces matching Go backend schema
│   │   └── utils.ts      # Formatting and styling utilities (cn)
│   │
│   ├── routes/           # SvelteKit 2 pages (+layout.svelte, +page.svelte, etc.)
│   ├── app.html
│   └── app.css           # Tailwind CSS & design tokens
├── tests/                # Vitest unit & component tests
├── package.json
├── svelte.config.js
├── vite.config.ts
├── tsconfig.json
├── tailwind.config.js
└── README.md
```

---

## Getting Started

### 1. Prerequisites
- Node.js `^20.0.0` or `^22.0.0`
- npm `^10.0.0` or `^11.0.0`

### 2. Installation
```bash
cd frontend
npm install
```

### 3. Environment Variables
Create a `.env` file based on `.env.example`:
```dotenv
PUBLIC_API_BASE_URL=http://localhost:8000
PUBLIC_APP_NAME="Flickey Rental Marketplace"
```

### 4. Development Server
```bash
npm run dev
```
The application will be available at `http://localhost:5173`.

---

## Testing & Quality Checks

### Run Unit Tests
```bash
npm test
```

### Type Checking & Svelte Checks
```bash
npm run check
```

### Production Build
```bash
npm run build
```

---

## Portability

The entire `frontend/` directory is self-contained. To run it in another repository:
```bash
cp -r frontend/ /path/to/new-project/frontend
cd /path/to/new-project/frontend
npm install
npm run dev
```
Configure `PUBLIC_API_BASE_URL` in `.env` to point to your target Go API instance.
