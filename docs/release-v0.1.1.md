# Release v0.1.1

Release di consolidamento per preparare `retronet-api`.

Novita':

- `Snapshot()` con dimensioni, righe a larghezza fissa, cursore, input pendente
  e byte raw in attesa
- `DrainOutput()` per consegnare i nuovi byte a CLI o websocket
- `Resize(width, height)` con preservazione del contenuto visibile
- ANSI CSI esteso:
  - erase display `J` modalita' `0`, `1`, `2`
  - erase line `K` modalita' `0`, `1`, `2`
  - cursor up/down/right/left `A/B/C/D`
- test per sequenze ANSI sconosciute/incomplete
- test di accesso concorrente
- documentazione del contratto terminale

Licenza e provenienza:

- nessuna ROM, font, terminfo, manuale storico copiato o dump incluso
- subset ASCII/ANSI scritto da zero

Verifica:

```powershell
go test -count=1 ./...
go vet ./...
go run ./cmd/retronet-terminal -demo -screen
git diff --check
```
