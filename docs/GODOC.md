# 📚 Visualizando Documentação GoDoc

## Desenvolvimento Local

### Opção 1: servidor godoc (Clássico)

```bash
# Instalar godoc
go install golang.org/x/tools/cmd/godoc@latest

# Iniciar servidor
godoc -http=:6060

# Abrir navegador
open http://localhost:6060/pkg/github.com/fabyo/gordon-watcher/
```

### Opção 2: pkgsite (Moderno - pkg.go.dev localmente)

```bash
# Instalar pkgsite
go install golang.org/x/pkgsite/cmd/pkgsite@latest

# Iniciar servidor
pkgsite -http=:8080

# Abrir navegador
open http://localhost:8080/github.com/fabyo/gordon-watcher
```

### Opção 3: go doc (Terminal)

```bash
# Ver documentação do pacote
go doc github.com/fabyo/gordon-watcher/internal/watcher

# Ver tipo específico
go doc github.com/fabyo/gordon-watcher/internal/watcher.Config

# Ver método específico
go doc github.com/fabyo/gordon-watcher/internal/watcher.Watcher.Start
```

## Online (após publicação)

Uma vez publicado no GitHub, a documentação estará automaticamente disponível em:
- https://pkg.go.dev/github.com/fabyo/gordon-watcher

## Gerando HTML Estático

```bash
# Gerar documentação HTML
godoc -url=/pkg/github.com/fabyo/gordon-watcher/ > docs/godoc.html
```
