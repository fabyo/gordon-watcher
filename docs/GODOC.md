# 📚 Visualizando Documentação GoDoc

## Local Development

### Option 1: godoc server (Classic)

```bash
# Install godoc
go install golang.org/x/tools/cmd/godoc@latest

# Start server
godoc -http=:6060

# Open browser
open http://localhost:6060/pkg/github.com/fabyo/gordon-watcher/
```

### Option 2: pkgsite (Modern - pkg.go.dev locally)

```bash
# Install pkgsite
go install golang.org/x/pkgsite/cmd/pkgsite@latest

# Start server
pkgsite -http=:8080

# Open browser
open http://localhost:8080/github.com/fabyo/gordon-watcher
```

### Option 3: go doc (Terminal)

```bash
# View package documentation
go doc github.com/fabyo/gordon-watcher/internal/watcher

# View specific type
go doc github.com/fabyo/gordon-watcher/internal/watcher.Config

# View specific method
go doc github.com/fabyo/gordon-watcher/internal/watcher.Watcher.Start
```

## Online (after publishing)

Once published to GitHub, documentation will be automatically available at:
- https://pkg.go.dev/github.com/fabyo/gordon-watcher

## Generating Static HTML

```bash
# Generate HTML documentation
godoc -url=/pkg/github.com/fabyo/gordon-watcher/ > docs/godoc.html
```
