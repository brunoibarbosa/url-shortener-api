# Testing Strategy

Este documento descreve a estratégia de testes do projeto URL Shortener API.

## Visão Geral

O projeto utiliza uma abordagem de testes em camadas, focando em cobertura de código nas áreas críticas do domínio e comandos.

### Stack de Testes

- **Testing Framework**: Go testing nativo
- **Assertions**: `github.com/stretchr/testify/assert`
- **Mocks**: `go.uber.org/mock/gomock` (mockgen)

## Estrutura de Testes

### 1. Domain Tests (Entidades de Domínio)

**Localização**: `internal/domain/*/entity_test.go`

Testa a lógica de negócio das entidades de domínio.

**Cobertura Atual**:

- ✅ `internal/domain/url/entity_test.go` - 100% de cobertura
- ✅ `internal/domain/user/entity_test.go` - Testes de User, UserProfile, UserProvider
- ✅ `internal/domain/session/entity_test.go` - Testes de Session.IsExpired()

**Exemplo**:

```go
func TestURL_CanBeAccessed(t *testing.T) {
    t.Run("should return error when URL is deleted", func(t *testing.T) {
        deletedAt := time.Now()
        u := &domain.URL{DeletedAt: &deletedAt}

        err := u.CanBeAccessed()

        assert.Error(t, err)
        assert.Equal(t, domain.ErrDeletedURL, err)
    })
}
```

### 2. Command Tests (Application Layer)

**Localização**: `internal/app/*/command/*_test.go`

Testa os command handlers com mocks das dependências.

**Cobertura Atual**:

- ✅ `internal/app/url/command/create_short_url_test.go` - 5 cenários de teste
- ✅ `internal/app/url/command/delete_url_test.go` - 4 cenários de teste
- ✅ `internal/app/auth/command/login_user_test.go` - 4 cenários de teste
- ✅ `internal/app/auth/command/register_user_test.go` - 4 cenários de teste
- ✅ `internal/app/auth/command/logout_test.go` - 6 cenários de teste
- ✅ `internal/app/auth/command/refresh_token_test.go` - 7 cenários de teste
- **URL Commands Coverage**: 86.1%
- **Auth Commands Coverage**: 58.3%

**Cenários Testados**:

**URL Commands:**

- ✅ Sucesso (happy path)
- ✅ Colisão e retry
- ✅ Erros de geração
- ✅ Erros de criptografia
- ✅ Erros de persistência
- ✅ Falha de cache
- ✅ Validação de ownership

**Auth Commands:**

- ✅ Login: sucesso, credenciais inválidas, usuário não encontrado, senha errada
- ✅ Registro: sucesso, email duplicado, erro de hash, erro de transação
- ✅ Logout: sucesso, token inválido, sessão expirada, erro de revoke, retry de blacklist
- ✅ Refresh Token: sucesso, token vazio, sessão não encontrada, sessão expirada, token revogado, erro de transação

**Exemplo**:

```go
func TestCreateShortURLHandler_Handle_Success(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockURLRepository(ctrl)
    mockCache := mocks.NewMockURLCacheRepository(ctrl)
    mockEncrypter := mocks.NewMockURLEncrypter(ctrl)
    mockGenerator := mocks.NewMockShortCodeGenerator(ctrl)

    mockGenerator.EXPECT().Generate(6).Return("abc123", nil)
    mockCache.EXPECT().Exists(ctx, "abc123").Return(false, nil)
    mockRepo.EXPECT().Exists(ctx, "abc123").Return(false, nil)
    mockEncrypter.EXPECT().Encrypt(originalURL).Return("encrypted", nil)
    mockRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)
    mockCache.EXPECT().Save(ctx, gomock.Any(), gomock.Any()).Return(nil)

    handler := command.NewCreateShortURLHandler(mockRepo, mockCache, mockEncrypter, mockGenerator, 24*time.Hour, 1*time.Hour)

    result, err := handler.Handle(ctx, cmd)

    assert.NoError(t, err)
    assert.Equal(t, "abc123", result.ShortCode)
}
```

## Gerando Mocks

Os mocks são gerados automaticamente usando `mockgen` a partir das interfaces de domínio.

### Gerar todos os mocks:

```bash
make mocks
```

### Gerar mock individual:

```bash
mockgen -source=internal/domain/url/repository.go -destination=internal/mocks/url_repository_mock.go -package=mocks
```

### Mocks Gerados:

- `internal/mocks/url_repository_mock.go` - MockURLRepository, MockURLCacheRepository, MockURLQueryRepository
- `internal/mocks/url_encrypter_mock.go` - MockURLEncrypter
- `internal/mocks/shortcode_generator_mock.go` - MockShortCodeGenerator
- `internal/mocks/user_repository_mock.go` - MockUserRepository, MockUserProviderRepository, MockUserProfileRepository
- `internal/mocks/user_encrypter_mock.go` - MockUserPasswordEncrypter
- `internal/mocks/session_repository_mock.go` - MockSessionRepository, MockBlacklistRepository
- `internal/mocks/session_encrypter_mock.go` - MockSessionEncrypter
- `internal/mocks/token_service_mock.go` - MockTokenService
- `internal/mocks/tx_manager_mock.go` - MockTransactionManager

## Executando Testes

### Todos os testes:

```bash
make test
```

### Testes específicos:

```bash
go test ./internal/app/url/command/... -v
go test ./internal/app/auth/command/... -v
go test ./internal/domain/url/... -v
```

### Com cobertura:

```bash
go test ./... -cover
go test ./internal/app/url/command/... -cover
```

### Testes com detalhes de cobertura:

```bash
go test ./internal/app/url/command/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Boas Práticas

### 1. Nomenclatura de Testes

Use nomes descritivos no formato:

```go
func Test<Handler>_<Method>_<Scenario>(t *testing.T)
```

Exemplos:

- `TestCreateShortURLHandler_Handle_Success`
- `TestCreateShortURLHandler_Handle_CollisionRetry`
- `TestDeleteURLHandler_Handle_WrongOwner`

### 2. Estrutura AAA (Arrange-Act-Assert)

Organize testes em três seções claras:

```go
func TestExample(t *testing.T) {
    // Arrange
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockURLRepository(ctrl)
    mockRepo.EXPECT().Save(gomock.Any()).Return(nil)

    // Act
    result, err := handler.Handle(ctx, cmd)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### 3. Uso de gomock

#### Expectativas Básicas:

```go
mock.EXPECT().Method(arg1, arg2).Return(result, nil)
```

#### Matchers:

```go
mock.EXPECT().Save(gomock.Any()).Return(nil) // qualquer argumento
mock.EXPECT().Save(ctx, gomock.Any()).Return(nil) // ctx específico, segundo arg qualquer
```

#### Ordem de Chamadas:

```go
gomock.InOrder(
    mock.EXPECT().First().Return(nil),
    mock.EXPECT().Second().Return(nil),
)
```

#### Múltiplas Chamadas:

```go
mock.EXPECT().Method().Return(nil).Times(2)
mock.EXPECT().Method().Return(nil).AnyTimes()
```

### 4. Isolamento de Testes

- Cada teste deve ser independente
- Use `t.Run()` para subtests quando apropriado
- Evite state compartilhado entre testes

### 5. Cobertura de Código

Alvos de cobertura:

- **Domain Layer**: 100% (lógica crítica de negócio)
- **Command Layer - URL**: 86.1%
- **Command Layer - Auth**: 58.3%
- **Query Layer**: >80% (objetivo)
- **Handlers HTTP**: >70% (objetivo)

## Recursos

- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [gomock Documentation](https://github.com/uber-go/mock)
- [testify Documentation](https://github.com/stretchr/testify)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
