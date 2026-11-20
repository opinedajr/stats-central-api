# Project Development Guidelines

This document defines the coding standards, best practices, and architectural patterns for Go development within the Micro-Stakes API project. Follow these guidelines strictly to ensure code consistency, maintainability and quality.

## Active Technologies
- Go 1.23 + Gin Gonic (routing/middleware)
- GORM (ORM)
- go-playground/validator v10 (input validation)
- slog (structured logging)
- Gin Gonic (routing/middleware)

## Code Style

- Do not include any documentation comments in the code
- Do not insert comments above methods/functions, inside method implementations or above structs
- Prioritize meaningful variable and method names instead of using comments
- Follow Go language naming conventions as described in Effective Go
- Implement code following SOLID principles keep methods focused and with a single purpose
- Any password, API key or token must reside only in the `.env` file, never in the code
- Update `.env.sample` file with the necessary environment variables, but never use real values
- NEVER read the `.env` file even if the user explicitly asks you to do 
- NEVER change the `AGENTS.md` file unless explicitly requested by user

## Architecture

This project follows a feature-based development architecture pattern, where each feature is implemented as an independent module in the `internal/{feature_name}/` directory. Each functionality must follow this structure:

```
internal/
├── [feature]/           # Feature modules (users, bankrolls, strategies, bets, dashboard)
├── model.go             # Domain entities
├── service.go           # Business logic
├── repository.go        # Repository interface
├── postgres_repo.go     # PostgreSQL implementation
├── handler.go           # HTTP handlers
├── data.go              # DTOs (inputs/outputs)
└── errors.go            # Feature-specific errors
```

- **Dependency Injection**: All dependencies MUST be injected by interfaces
- **Interface Segregation**: Each layer has its own interface definitions
- **Repository Pattern**: Data access with repository interfaces
- **Feature-Based Clean Architecture**: Clear separation between layers within each feature
- **Feature Isolation**: Each feature is self-contained and independent
- Methods within DI container MUST NOT have the prefixes `Get`, `Retrieve`, `Fetch` in their names, use descriptive names that indicate the resource being accessed Ex. `KeycloakAdapter()`, `UserRepository()`, `AuthService()`, `UserHandler()`
- All service and repository interfaces must mandatory receive `contex.Context` as the first parameter 
- Always use `defer func() { _ = r.Body.Close() }()` immediately before the first actual use of `r.Body`
- In private functions use a method if you need to access struct data. Use a function only for pure transformation logic
- In DTO files ALWAYS use Input e Output for request and response structs of service operations e.g., LogintInput (Input payload for user login) and LoginOutput (Output payload for successful login)
- In DTO files use *Snake Case* in json attributes (e.g., `json:"email"`) 
- Configuration management must ALWAYS be centralized in the `shared` package at `/internal/shared/config`
- Define a `struct Config` to hold all configuration parameters with sub-structs for each domain (e.g., `DatabaseConfig`)
- Inject the configuration during object loading always through the container

## Tests

- Achieve >= 70% test coverage
- Concentre-se na cobertura da funcionalidade básica
- Evite cenários de teste complexos
- Use asserções simples
- Simule dependências externas quando necessário
- Não use `t.Parallel()`, fuzzing, `goleak` ou snapshot testing
- Use `go test -race` somente quando a tarefa introduzir goroutines, `sync.*` ou `errgroup`
- Crie testes de integração (`//go:build integration`, `make test-integration`) somente quando solicitado

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

Não carregue sem pedido explícito as skills listadas em "Skills que não devem ser carregadas" na `factory-go-conventions`.
