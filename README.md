# Univer Tracker

Sistema de logística inteligente para e-commerces — rastreamento, automação de status, integração com transportadoras e notificações em tempo real.

## Stack

- **Backend:** Go 1.23 (Chi router, sqlc, Postgres)
- **Frontend:** Next.js 15 + TypeScript + Tailwind + shadcn/ui
- **Cache/Queue:** Redis (Asynq)
- **Database:** PostgreSQL 16
- **Deploy:** Coolify (Docker Compose)

## Arquitetura

```
┌──────────────────────────────────────────────────────────┐
│  Frontend (Next.js)                                      │
│  ↓ REST + Server-Sent Events                             │
│  Backend (Go)                                            │
│  ↓                                                        │
│  Postgres + Redis                                        │
│  ↑                                                        │
│  WooCommerce ←→ Frenet ←→ WAHA (WhatsApp)                │
└──────────────────────────────────────────────────────────┘
```

## Desenvolvimento Local

```bash
# Subir infra (Postgres + Redis)
docker-compose up -d postgres redis

# Backend
cd backend
make dev

# Frontend
cd frontend
pnpm install
pnpm dev
```

Frontend: http://localhost:3000
API: http://localhost:8080

## Estrutura

```
univer-tracker/
├── backend/              # API Go + workers
├── frontend/             # Dashboard Next.js
├── docker-compose.yml    # Orquestração local
└── .github/workflows/    # CI/CD
```

## Convenções

- **Commits:** [Conventional Commits](https://www.conventionalcommits.org/)
- **Branching:** `main` (produção) ← `develop` ← `feature/*`
- **Code style:** Go — `gofmt` + `golangci-lint` | TS — `prettier` + `eslint`

## Licença

Proprietário — Univer Beauty Comércio LTDA. Todos os direitos reservados.
