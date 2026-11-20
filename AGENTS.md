# Project Development Guidelines

This document defines the facts and rules of the **Stats Central API** project: what it is, how it is structured, and what applies here. Standing Go conventions that apply to every project live in the `factory-go-conventions` skill; technical reference lives in the `golang-*` skills. Precedence: this file > `factory-go-conventions` > `golang-*` skills.

## Active Technologies
- Go 1.23 + Gin Gonic (routing/middleware)
- GORM (ORM) with multi-driver support: MySQL, PostgreSQL (SQLite for tests)
- golang-migrate (database migrations)
- go-playground/validator v10 (input validation)
- slog (structured logging)

## Secrets and Files
- Any password, API key or token must reside only in the `.env` file, never in the code
- Update `.env.sample` with the necessary environment variables, but never use real values
- NEVER read the `.env` file even if the user explicitly asks you to do
- NEVER change the `AGENTS.md` file unless explicitly requested by user

## Architecture

This project follows a feature-based architecture. Each feature is an independent module in `internal/{feature}/`. Current features: `match`, `teams`, `tournament`, `analyse`, `healthcheck`.

```
cmd/
└── api/                    # application entry point
internal/
├── [feature]/              # match, teams, tournament, analyse
│   ├── model.go            # domain entities
│   ├── data.go             # DTOs (inputs/outputs)
│   ├── repository.go       # repository interface
│   ├── mysql_repo.go       # GORM implementation
│   ├── service.go          # business logic + service interface
│   ├── handler.go          # HTTP handlers (Gin)
│   └── errors.go           # feature-specific errors
├── healthcheck/            # feature without repository layer
├── di/                     # manual DI container
├── infrastructure/
│   └── database/           # database connections (mysql, postgres, sqlite)
└── shared/                 # config, logger, api responses, pagination
```

Facts and rules of this architecture:

- All dependencies are injected by interfaces; service and repository interfaces live in `repository.go` and `service.go` of each feature (not in a separate `interfaces.go`)
- All service and repository interfaces must receive `context.Context` as the first parameter; handlers start propagation with `c.Request.Context()` (Gin)
- Methods within the DI container MUST NOT have the prefixes `Get`, `Retrieve`, `Fetch` in their names — use names that indicate the resource being accessed, e.g. `MatchRepository()`, `AnalyseService()`, `UserHandler()`
- In DTO files (`data.go`) ALWAYS use `Input` and `Output` suffixes for request and response structs of service operations (e.g. `MatchOutput`) and snake_case in json attributes (e.g. `json:"home_team_id"`)
- In private functions use a method if you need to access struct data; use a function only for pure transformation logic
- Configuration management is centralized in `internal/shared/config` (`struct Config` with sub-structs per domain, e.g. `DatabaseConfig`) and injected through the container
- Database: GORM connections in `internal/infrastructure/database`; migrations via golang-migrate (`make migrate`, `make rollback`); never create or alter tables, indexes or migrations without explicit request

## Tests

Testing conventions (style, naming, mocks, prohibitions) are defined in `factory-go-conventions` (`references/testing.md`). Project facts:

- Coverage target: >= 70%
- Run with `make test`, `make test-v` or `make test-cover`
- There is no integration-test Makefile target yet; integration tests only when explicitly requested

## Skills Go (samber/cc-skills-golang)

As skills `golang-*` instaladas em `.claude/skills/` são referência técnica.
Em qualquer tarefa Go, carregue **sempre primeiro** a skill `factory-go-conventions`: ela define onde o projeto diverge das skills da comunidade e prevalece sobre elas. Precedência: este arquivo > `factory-go-conventions` > skills `golang-*`.

Carregue também as skills abaixo conforme a situação:

| Situação                                          | Skills                                                                                   |
| ------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Nova feature ou endpoint                          | `golang-testing`, `golang-error-handling`, `golang-structs-interfaces`, `golang-context` |
| Integração com API externa (Meta, Instagram etc.) | `golang-context`, `golang-error-handling`, `golang-safety`                               |
| Escrita ou correção de testes                     | `golang-testing`, `golang-stretchr-testify`                                              |
| Bug, panic ou comportamento inesperado            | `golang-troubleshooting`, `golang-safety`                                                |
| Revisão de código ou PR                           | `golang-code-style`, `golang-naming`, `golang-safety`, `golang-security`                 |
| Refatoração                                       | `golang-refactoring`, `golang-code-style`                                                |
| Adicionar ou atualizar dependência                | `golang-popular-libraries`, `golang-dependency-management`, `golang-pkg-go-dev`          |
| Acesso a banco de dados                           | `golang-database`                                                                        |
| Alteração no container de DI                      | `golang-dependency-injection`                                                            |
| Modernização de código                            | `golang-modernize`                                                                       |
| Navegar no código, achar referências, rename      | `golang-gopls`                                                                           |
| Goroutines ou escolha de estruturas de dados      | `golang-concurrency`, `golang-data-structures`                                           |
| Otimização ou medição (somente quando solicitado) | `golang-performance`, `golang-benchmark`                                                 |

Exclusões **condicionais** que se aplicam a este projeto (justificativas genéricas em `references/excluded-skills.md` da `factory-go-conventions`):

- `golang-observability`: o projeto não tem stack de métricas/tracing (apenas `slog`)
- `golang-continuous-integration`: o projeto não tem pipeline de CI
- `golang-cli`: o projeto é uma API REST, não uma aplicação de linha de comando
- Skills de bibliotecas que não estão no `go.mod` seguem a regra genérica da `factory-go-conventions`

As exclusões **fixas** (`golang-how-to`, `golang-swagger`, `golang-samber-slog`, `golang-samber-lo`) valem para qualquer projeto, incluindo este, e estão listadas na `factory-go-conventions`.
