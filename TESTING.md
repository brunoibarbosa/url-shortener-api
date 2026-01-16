# Testing Guide

Este documento explica a estratégia de testes do projeto e como executá-los.

## Índice

- [Estratégia de Testes](#estratégia-de-testes)
- [Comandos Disponíveis](#comandos-disponíveis)
- [Tipos de Testes](#tipos-de-testes)
- [Workflow Recomendado](#workflow-recomendado)
- [CI/CD](#cicd)
- [Estrutura dos Testes](#estrutura-dos-testes)

---

## Estratégia de Testes

Os testes estão organizados em **camadas** para otimizar velocidade e feedback:

### Fast Tests (Testes Rápidos)

- **Domain**: Lógica de negócio pura
- **Application**: Casos de uso e comandos
- **Services**: Serviços de infraestrutura (crypto, jwt, shortcode)
- **Sem Docker**: Rodam localmente sem dependências externas
- **Uso**: Pre-commit hooks, desenvolvimento diário

### Integration Tests (Testes de Integração)

- **Redis Repositories**: Cache e sessões
- **PostgreSQL Repositories**: Session, URL, User
- **Com Docker**: Requer testcontainers
- **Uso**: Antes de push, validação manual, CI/CD

### Property-Based Tests (Testes Baseados em Propriedades)

- **Validação de invariantes**: Testa propriedades que devem sempre ser verdadeiras
- **Geração de dados aleatórios**: Executa centenas de cenários automaticamente
- **Uso**: Encontrar edge cases que testes manuais não cobrem

### Edge Case Tests (Testes de Casos Limites)

- **Valores extremos**: Strings vazias, máximos, mínimos
- **Caracteres especiais**: Unicode, SQL injection, XSS
- **Condições de erro**: Duplicatas, não existentes, operações inválidas
- **Uso**: Garantir robustez em cenários adversos

### E2E Tests (Testes End-to-End)

- **Fluxos completos**: Registro → Login → Criar URL → Redirecionar
- **HTTP Layer**: Testa handlers, middleware, serialização
- **Uso**: Validar integração completa da API

### All Tests (Todos os Testes)

- Executa `test-fast` + `test-integration` em sequência
- **Uso**: CI/CD, validação final antes de merge

---

## Comandos Disponíveis

### `make test` ou `make test-fast`

**Testes rápidos (sem Docker)**

```bash
make test
# ou
make test-fast
```

**Quando usar:**

- Durante desenvolvimento (feedback rápido)
- Pre-commit hooks
- Validação de mudanças simples
- Quando Docker não está disponível

---

### `make test-integration`

**Testes de integração (requer Docker)**

```bash
make test-integration
```

**Quando usar:**

- Antes de fazer push
- Validação de mudanças em repositories
- Testes de infraestrutura
- Validação completa de integração

**Requisitos:**

- Docker Desktop rodando
- Testcontainers funcionando

---

### `make test-property`

**Testes baseados em propriedades (requer Docker)**

```bash
make test-property
```

**O que testa:**

- Invariantes de negócio (CreatedAt <= UpdatedAt)
- Unicidade de constraints (email, short_code)
- Idempotência de operações (soft delete)
- Geração de 100-300 cenários aleatórios por teste

**Quando usar:**

- Validar robustez de repositories
- Encontrar bugs em edge cases
- Antes de releases importantes

---

### `make test-edge`

**Testes de casos limites (requer Docker)**

```bash
make test-edge
```

**O que testa:**

- Strings muito longas (URLs de 2048 chars)
- Strings muito curtas (emails mínimos)
- Caracteres especiais e Unicode
- Operações duplicadas e não existentes
- SQL injection e XSS attempts

**Quando usar:**

- Validar segurança
- Garantir compatibilidade com dados reais
- Antes de deploy em produção

---

### `make test-e2e`

**Testes end-to-end (requer Docker)**

```bash
make test-e2e
```

**O que testa:**

- Fluxo completo de registro e login
- Criação e redirecionamento de URLs encurtadas
- Gerenciamento de sessões (list, refresh, logout)
- Tratamento de erros HTTP

**Quando usar:**

- Validar integração de todos os componentes
- Antes de releases
- Smoke tests em staging

---

### `make test-all`

**Suite completa de testes**

```bash
make test-all
```

**Quando usar:**

- Validação final antes de PR
- CI/CD pipelines
- Release builds
- Validação completa de qualidade

---

## Tipos de Testes

### `make coverage`

**Relatório de cobertura HTML**

```bash
make coverage
```

**Quando usar:**

- Análise de cobertura de código
- Identificar áreas não testadas
- Documentação de qualidade

**Gera:**

- `coverage.out` - Dados de cobertura
- `coverage.html` - Relatório visual (abrir no browser)
