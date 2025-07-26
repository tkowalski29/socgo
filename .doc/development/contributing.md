# Contributing Guide

Przewodnik dla kontrybutorów chcących przyczynić się do rozwoju SocGo.

## 🤝 Jak zacząć

### 1. Setup środowiska
```bash
# Fork repository na GitHub
# Clone your fork
git clone https://github.com/YOUR_USERNAME/socgo.git
cd socgo

# Add upstream remote
git remote add upstream https://github.com/tkowalski/socgo.git

# Setup development environment
cp config.yml.example config.yml
go mod download
go test ./...
```

### 2. Wybierz zadanie
- Sprawdź [Issues](https://github.com/tkowalski/socgo/issues)
- Szukaj etykiet `good first issue` lub `help wanted`
- Skomentuj issue żeby zadeklarować pracę nad nim
- Poczekaj na potwierdzenie od maintainerów

## 📋 Typy kontrybucji

### 🐛 Bug Fixes
- Reprodukuj błąd
- Napisz test case
- Napraw błąd
- Sprawdź czy test przechodzi

### ✨ Nowe funkcje
- Otwórz Feature Request issue
- Czekaj na dyskusję i zatwierdzenie
- Implementuj zgodnie z feedback
- Dodaj dokumentację

### 📚 Dokumentacja
- Popraw istniejącą dokumentację
- Dodaj brakujące przykłady
- Tłumacz na inne języki
- Aktualizuj API reference

### 🧪 Testy
- Dodaj brakujące unit testy
- Napisz integration testy
- Popraw test coverage
- Dodaj performance testy

## 🔄 Workflow

### Branch Strategy
```bash
# Create feature branch
git checkout -b feature/oauth-improvements
git checkout -b bugfix/rate-limit-issue
git checkout -b docs/api-examples
```

### Commit Guidelines

Używamy [Conventional Commits](https://conventionalcommits.org/):

```bash
# Format: type(scope): description
git commit -m "feat(oauth): add refresh token support"
git commit -m "fix(database): resolve connection pool leak"
git commit -m "docs(api): add OAuth examples"
git commit -m "test(post): add integration tests"
git commit -m "refactor(handlers): extract validation logic"
```

### Typy commitów:
- `feat`: Nowa funkcjonalność
- `fix`: Naprawa błędu
- `docs`: Zmiany w dokumentacji
- `style`: Formatowanie kodu (bez zmian logiki)
- `refactor`: Refaktoring kodu
- `test`: Dodanie/modyfikacja testów
- `chore`: Zmiany w build process, narzędziach, etc.

### Scopes (opcjonalne):
- `oauth`: OAuth functionality
- `database`: Database layer
- `api`: API endpoints
- `ui`: User interface
- `providers`: Social media providers
- `config`: Configuration

## 📝 Standardy kodowania

### Go Code Style

```go
// Use gofmt and golint
go fmt ./...
golint ./...

// Follow Go naming conventions
type PostService struct {          // Exported types: PascalCase
    repo PostRepository            // Exported fields: PascalCase
    logger *slog.Logger           // Unexported fields: camelCase
}

func (s *PostService) CreatePost() {} // Exported methods: PascalCase
func (s *PostService) validatePost() {} // Unexported methods: camelCase

const MaxRetries = 3              // Exported constants: PascalCase
const defaultTimeout = 30        // Unexported constants: camelCase
```

### Error Handling

```go
// Always handle errors
result, err := someFunction()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Use custom error types when appropriate
type ValidationError struct {
    Field string
    Value interface{}
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field %s with value %v", e.Field, e.Value)
}
```

### Documentation

```go
// Document all exported functions and types
// CreatePost creates a new post in the system.
// It validates the input, saves to database, and triggers notifications.
//
// Returns the created post ID and any error that occurred during creation.
func CreatePost(content string, userID string) (uint, error) {
    // Implementation
}

// Use examples in documentation
// Example:
//   postID, err := CreatePost("Hello world!", "user123")
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Printf("Created post with ID: %d\n", postID)
```

## 🧪 Wymagania testowe

### Test Coverage
- Nowy kod musi mieć minimum 80% coverage
- Krytyczne funkcje wymagają 100% coverage
- Dodaj testy przed fixem dla bug reportów

### Typy testów

#### Unit Tests
```go
func TestPostService_Create(t *testing.T) {
    // Arrange
    mockRepo := &MockPostRepository{}
    service := NewPostService(mockRepo)
    
    // Act
    postID, err := service.CreatePost("test content", "user123")
    
    // Assert
    assert.NoError(t, err)
    assert.Greater(t, postID, uint(0))
    mockRepo.AssertExpectations(t)
}
```

#### Integration Tests
```go
//go:build integration
// +build integration

func TestPostWorkflow_Integration(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Test full workflow
    service := NewPostService(db)
    postID, err := service.CreatePost("test", "user123")
    
    assert.NoError(t, err)
    
    // Verify in database
    post, err := service.GetPost(postID)
    assert.NoError(t, err)
    assert.Equal(t, "test", post.Content)
}
```

### Running Tests
```bash
# Unit tests only
go test ./...

# With integration tests
go test -tags=integration ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/handlers

# Verbose
go test -v ./...
```

## 📚 Documentation Guidelines

### API Documentation
```go
// OpenAPI/Swagger comments
// @Summary Create new post
// @Description Creates a new social media post
// @Tags posts
// @Accept json
// @Produce json
// @Param post body PostRequest true "Post data"
// @Success 201 {object} PostResponse
// @Failure 400 {object} ErrorResponse
// @Router /posts [post]
func (h *PostHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### Markdown Documentation
- Używaj nagłówków H1-H6 hierarchicznie
- Dodawaj przykłady kodu
- Używaj diagramów Mermaid dla przepływów
- Linkuj do powiązanych sekcji

### VitePress Documentation
```bash
# Start documentation server
cd .doc
npm install
npm run docs:dev

# Build documentation
npm run docs:build
```

## 🔍 Code Review Process

### Przed submission
1. **Self-review**: Przejrzyj swój kod
2. **Tests**: Sprawdź czy wszystkie testy przechodzą
3. **Documentation**: Zaktualizuj dokumentację
4. **Linting**: Uruchom linters

```bash
# Pre-submission checklist
go fmt ./...
go vet ./...
golint ./...
go test ./...
go test -race ./...
```

### Pull Request Template

Użyj tego template dla PR:

```markdown
## Description
Brief description of what this PR does.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## How Has This Been Tested?
- [ ] Unit tests
- [ ] Integration tests
- [ ] Manual testing

Describe the tests that you ran to verify your changes.

## Checklist:
- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes

## Screenshots (if applicable):
Add screenshots to help explain your changes.

## Additional Notes:
Any additional information that reviewers should know.
```

### Review Criteria

Reviewers sprawdzają:

1. **Correctness**: Czy kod robi to co ma robić?
2. **Performance**: Czy nie ma performance regressions?
3. **Security**: Czy nie wprowadza luk bezpieczeństwa?
4. **Maintainability**: Czy kod jest czytelny i łatwy w utrzymaniu?
5. **Tests**: Czy są odpowiednie testy?
6. **Documentation**: Czy dokumentacja jest aktualna?

## 🐛 Bug Report Guidelines

### Dobry bug report zawiera:

```markdown
## Bug Description
Clear and concise description of what the bug is.

## To Reproduce
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

## Expected Behavior
Clear description of what you expected to happen.

## Actual Behavior
What actually happened.

## Screenshots
If applicable, add screenshots to help explain your problem.

## Environment:
- OS: [e.g. macOS 12.0]
- Go Version: [e.g. 1.21.0]
- SocGo Version: [e.g. 1.0.0]
- Browser: [e.g. Chrome 91.0]

## Additional Context
Add any other context about the problem here.

## Possible Solution
If you have ideas for fixing the bug, mention them here.
```

## ✨ Feature Request Guidelines

### Dobry feature request zawiera:

```markdown
## Feature Description
Clear and concise description of what you want to happen.

## Problem/Use Case
Explain the problem this feature would solve or the use case it addresses.

## Proposed Solution
Detailed description of how you think this should work.

## Alternatives Considered
Alternative solutions or features you've considered.

## Additional Context
Any other context, screenshots, or examples about the feature request.

## Implementation Ideas
If you have ideas about how to implement this, share them.
```

## 🏆 Recognition

Kontrybutorzy są doceniani poprzez:

- **Contributors list** w README
- **Changelog mentions** dla znaczących zmian
- **Social media shoutouts** 
- **Maintainer status** dla aktywnych kontrybutorów

## 📞 Komunikacja

### Gdzie szukać pomocy:
- **GitHub Issues**: Pytania techniczne
- **GitHub Discussions**: Ogólne dyskusje
- **Discord/Slack**: Real-time chat (jeśli dostępne)

### Guidelines komunikacji:
- Bądź uprzejmy i profesjonalny
- Używaj jasnego, klarownego języka
- Podaj kontekst gdy zadajesz pytania
- Szanuj czas innych kontrybutorów

## 🔒 Security

### Zgłaszanie luk bezpieczeństwa:
- **NIE** twórz publicznie issues dla luk bezpieczeństwa
- Wyślij email na: security@socgo.example.com
- Opisz lukę szczegółowo
- Daj czas na fix przed publicznym ujawnieniem

### Security checklist:
- Nie commituj secrets/keys
- Waliduj wszystkie user inputs
- Używaj HTTPS wszędzie
- Szyfruj sensitive data

## 📈 Performance Guidelines

### Nie optymalizuj przedwcześnie
- Najpierw napisz działający kod
- Potem measure performance
- Dopiero potem optymalizuj bottlenecks

### Performance best practices:
- Używaj database indexes
- Implementuj caching gdzie sensowne
- Avoid N+1 queries
- Profile memory usage
- Monitor API response times

## 🎯 Development Roadmap

### Priorytetowe obszary:
1. **Provider Extensions**: Nowe platformy (LinkedIn, Twitter)
2. **Analytics**: Metryki i reporting
3. **Bulk Operations**: Mass publishing
4. **Mobile API**: REST API improvements
5. **Performance**: Scaling improvements

### Jak pomagać:
- Sprawdź roadmap issues
- Zaproponuj implementację
- Dyskutuj approach z maintainerami
- Implementuj po zatwierdzeniu

---

Dziękujemy za zainteresowanie rozwojem SocGo! 🚀