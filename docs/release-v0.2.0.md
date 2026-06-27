# Release v0.2.0

Release dedicata al primo terminale locale interattivo.

Novita':

- nuovo comando `go run ./cmd/retronet-terminal-live`
- raw mode su Windows e Linux usando solo libreria standard Go
- fallback in modalita' linea quando raw mode non e' disponibile
- repaint ANSI dello snapshot del core terminale
- controlli locali: Backspace, `Ctrl+L`, `Ctrl+Q`, `Ctrl+C`, `Ctrl+D`
- flag `-width`, `-height`, `-demo`, `-script`
- test automatizzati del comando live tramite modalita' scriptata
- documentazione didattica in italiano

Uso:

```powershell
go run ./cmd/retronet-terminal-live
go run ./cmd/retronet-terminal-live -demo
go run ./cmd/retronet-terminal-live -width 100 -height 30
go run ./cmd/retronet-terminal-live -script "CIAO`r`nREADY"
```

Licenza e provenienza:

- nessuna ROM, font, terminfo, manuale storico copiato o dump incluso
- comportamento ASCII/ANSI generico scritto da zero
- nessuna dipendenza esterna aggiunta

Verifica:

```powershell
go test -count=1 ./...
go vet ./...
go run ./cmd/retronet-terminal-live -width 40 -height 6 -script "CIAO`r`nREADY"
git diff --check
```
