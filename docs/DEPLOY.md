# Deploy — Coolify (Docker Compose)

Guia oficial pra subir o Univer Tracker no Coolify usando o `docker-compose.yml`
do repositório como fonte da verdade. Toda a stack (Postgres, Redis, migrations,
API, worker, frontend) sobe num único recurso e Coolify cuida de SSL, rede
e auto-deploy via webhook.

## Pré-requisitos

- Coolify ≥ 4.0 rodando numa VPS com Docker.
- Domínios apontando para o IP do Coolify:
  - `tracker.lizzon.com.br` — frontend
  - `api.tracker.lizzon.com.br` — backend
- Conta GitHub com acesso ao repositório `univerbeauty777/univer-tracker`.

## 1. Criar o recurso

1. Coolify → **Projects** → seu projeto → **+ New Resource** → **Docker Compose Empty**.
   - Alternativa: **Public Repository** se for repo público / **Private Repository (with Github App)** com a App do Coolify instalada.
2. Cole o repositório: `https://github.com/univerbeauty777/univer-tracker`.
3. **Branch:** `main`.
4. **Base directory:** `/`.
5. **Docker Compose file location:** `/docker-compose.yml`.
6. Salve.

Coolify lê o compose, detecta os serviços (`postgres`, `redis`, `migrate`,
`backend`, `worker`, `frontend`) e mostra cada um na UI.

## 2. Configurar domínios

A mágica do Coolify acontece pelas envs `SERVICE_FQDN_BACKEND_8080` e
`SERVICE_FQDN_FRONTEND_3000` que já estão no compose: ao definir o domínio,
o Coolify gera os labels Traefik e provisiona SSL automaticamente.

- **backend** → **Domains** → `https://api.tracker.lizzon.com.br`
- **frontend** → **Domains** → `https://tracker.lizzon.com.br`

Não defina domínio para `postgres`, `redis`, `migrate`, `worker` — eles ficam
na rede interna `ut_internal`.

## 3. Variáveis de ambiente

Em **Environment Variables** do recurso, cole o bloco abaixo e preencha os
valores reais (cada `<...>` é placeholder).

```bash
# Postgres
POSTGRES_USER=univertracker
POSTGRES_PASSWORD=<gere com `openssl rand -base64 32`>
POSTGRES_DB=univertracker

# App
APP_ENV=production

# Auth
JWT_SECRET=<gere com `openssl rand -base64 64`>
JWT_EXPIRES_IN=7d

# WooCommerce
WC_URL=https://lizzon.com.br
WC_CONSUMER_KEY=<copiar do WP → WooCommerce → Settings → Advanced → REST API>
WC_CONSUMER_SECRET=<idem>
WC_WEBHOOK_SECRET=<gere com `openssl rand -hex 32`>

# Frenet
FRENET_API_TOKEN=<token Frenet>
FRENET_PANEL_EMAIL=<email Frenet>
FRENET_PANEL_PASSWORD=<senha Frenet>

# WAHA (gateway WhatsApp; ZapGrup hospeda nossa instância)
WAHA_URL=https://zapgrup.univerzap.cloud
WAHA_API_KEY=<chave do painel WAHA>
```

> **Importante:** Coolify resolve `SERVICE_URL_BACKEND_8080` e
> `SERVICE_URL_FRONTEND_3000` automaticamente a partir dos domínios da
> seção 2 — não precisa setar essas duas manualmente.

## 4. Deploy

Clique em **Deploy**. Sequência esperada:

1. `postgres` e `redis` sobem e ficam `healthy`.
2. `migrate` roda `migrate up`, aplica `0001_init.up.sql` e sai com código 0.
3. `backend` sobe, healthcheck `/healthz` passa.
4. `worker` sobe, healthcheck `pgrep` confirma o processo.
5. `frontend` faz build com `NEXT_PUBLIC_API_URL=https://api.tracker.lizzon.com.br`,
   sobe e healthcheck passa.

Tempo total na primeira vez: ~3–5 min (build do Go + Next.js).
Em deploys subsequentes: ~30–60s (BuildKit cache).

## 5. Verificar

```bash
curl -fsS https://api.tracker.lizzon.com.br/healthz
# {"status":"ok","time":"..."}

curl -fsS https://api.tracker.lizzon.com.br/api/v1/version
# {"name":"univer-tracker","version":"0.1.0"}
```

E acesse `https://tracker.lizzon.com.br` no navegador.

## 6. Auto-deploy (webhook)

No recurso, **Webhook** → copie a URL gerada → no GitHub
`Settings → Webhooks → Add webhook`:

- **Payload URL:** a URL copiada
- **Content type:** `application/json`
- **Secret:** se o Coolify pedir, cole a mesma
- **Events:** `Just the push event`

A partir daqui todo `git push origin main` redeploya o stack.

## 7. Logs e troubleshooting

- **Logs por serviço:** Coolify → recurso → aba do serviço → **Logs**.
- **Migrate falhou:** abra o log do serviço `migrate`. Se já rodou antes em
  estado parcial, force `migrate down 1` localmente apontando pra produção
  ou aplique `migrate force <versão>` (ver `migrate/migrate` docs).
- **Healthcheck flapping no backend:** geralmente env var faltando. `docker
  logs <container>` mostra `config: ... required`.
- **CORS bloqueando frontend:** confira que `PUBLIC_URL` no backend resolveu
  pro domínio público do frontend (deveria, via `SERVICE_URL_FRONTEND_3000`).

## 8. Deploy local (dev)

```bash
cp .env.example .env
# preencha o mínimo (POSTGRES_*, JWT_SECRET, WC_*, FRENET_*, WAHA_*)
docker compose up -d
```

- Frontend: http://localhost:3000
- API: http://localhost:8080/healthz

Os fallbacks `localhost` nas envs `SERVICE_URL_*` cobrem o cenário de dev.
