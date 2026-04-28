# Contributing

## Branching

- `main` — produção. Deploy automático no Coolify.
- `develop` — staging. Integração contínua.
- `feature/<nome>` — features novas.
- `fix/<nome>` — correções de bugs.

## Conventional Commits

Todos os commits seguem [Conventional Commits](https://www.conventionalcommits.org/):

```
<tipo>(<escopo opcional>): <descrição>

[corpo opcional]

[footer opcional]
```

**Tipos:**
- `feat` — nova funcionalidade
- `fix` — correção de bug
- `docs` — documentação
- `style` — formatação (sem mudança lógica)
- `refactor` — refatoração
- `perf` — melhoria de performance
- `test` — testes
- `chore` — manutenção
- `ci` — CI/CD

**Exemplos:**
```
feat(api): add tracking events endpoint
fix(frontend): resolve dark mode flicker on initial load
chore(deps): bump go to 1.23.5
```

## Pull Requests

1. Crie branch a partir de `develop`
2. Faça commits seguindo o padrão
3. Abra PR para `develop`
4. Aguarde CI passar e review

## Code Style

- **Go:** `gofmt` + `golangci-lint run` antes de commitar
- **TypeScript:** `prettier --write` + `eslint --fix` antes de commitar
- **SQL:** snake_case, lower-case keywords
