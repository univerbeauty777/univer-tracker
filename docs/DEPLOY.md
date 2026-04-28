# Deploy — Coolify

Guia para fazer o deploy do Univer Tracker no Coolify.

## Pré-requisitos

- Coolify rodando em uma VPS (ou Hetzner/Digital Ocean/etc).
- Domínio `tracker.lizzon.com.br` apontando para o IP do Coolify.

## Recursos a criar no Coolify

### 1. PostgreSQL 16

- **Database** → `univertracker`
- **Username** → `univertracker`
- **Password** → gere uma forte (32+ chars)
- **Port** → 5432 (interno)
- **Persistent volume** → habilitado

Anote a `DATABASE_URL` interna do Coolify (algo como `postgresql://univertracker:***@postgres:5432/univertracker`).

### 2. Redis 7

- **Persistent volume** → habilitado
- Anote a `REDIS_URL` interna (algo como `redis://redis:6379/0`).

### 3. Aplicação — backend (API)

- **Source:** GitHub → `univerbeauty777/univer-tracker`
- **Branch:** `main`
- **Build pack:** Dockerfile
- **Base directory:** `/backend`
- **Dockerfile target:** `api`
- **Port:** 8080
- **Domain:** `api.tracker.lizzon.com.br` (ou `/api/*` no mesmo domínio se usar reverse-proxy)

**Variáveis de ambiente:**

```
APP_ENV=production
APP_PORT=8080
APP_URL=https://api.tracker.lizzon.com.br
PUBLIC_URL=https://tracker.lizzon.com.br
DATABASE_URL=<da etapa 1>
REDIS_URL=<da etapa 2>
JWT_SECRET=<gerar 64+ chars>
WC_URL=https://lizzon.com.br
WC_CONSUMER_KEY=<copiar do WP>
WC_CONSUMER_SECRET=<copiar do WP>
WC_WEBHOOK_SECRET=<gerar>
FRENET_API_TOKEN=<copiar>
ZAPGRUP_URL=https://zapgrup.cloud
ZAPGRUP_API_TOKEN=<gerar/copiar>
```

### 4. Aplicação — worker

- Mesmo source que a API
- **Dockerfile target:** `worker`
- **Sem porta exposta**
- Mesmas variáveis de ambiente da API

### 5. Aplicação — frontend (Next.js)

- **Source:** GitHub → `univerbeauty777/univer-tracker`
- **Branch:** `main`
- **Build pack:** Dockerfile
- **Base directory:** `/frontend`
- **Port:** 3000
- **Domain:** `tracker.lizzon.com.br`

**Build args:**

```
NEXT_PUBLIC_API_URL=https://api.tracker.lizzon.com.br
```

## Auto-deploy (webhook)

No Coolify, em cada aplicação:

1. **Settings → Webhooks → GitHub**
2. Copie a URL do webhook que o Coolify gera.
3. No GitHub: `Settings → Webhooks → Add webhook`.
   - Payload URL: a URL copiada.
   - Content type: `application/json`.
   - Trigger: `push` events.
4. Cole a secret se o Coolify pedir.

A partir disso, todo `git push origin main` dispara um redeploy automático.

## Deploy local (alternativa Docker Compose)

```bash
cp .env.example .env
# preencha as variáveis
docker compose up -d
```

Acesse:
- Frontend: http://localhost:3000
- API: http://localhost:8080/healthz
