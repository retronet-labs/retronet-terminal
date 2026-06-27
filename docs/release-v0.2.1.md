# Release v0.2.1

Patch di usabilita' per il comando live.

Problema risolto:

- in alcune console Windows il prompt `READY>` poteva lampeggiare o comparire
  solo dopo Invio
- il testo digitato poteva non risultare visibile in modo stabile
- il fallback automatico a modalita' linea rendeva poco chiaro quando raw mode
  non era davvero disponibile

Correzioni:

- durante l'uso interattivo il comando invia alla console solo i nuovi byte
  prodotti da `Terminal.DrainOutput`, senza ridisegnare tutto lo snapshot a ogni
  tasto
- `Backspace` non cancella piu' il prompt quando la riga corrente e' vuota
- se raw mode non e' disponibile, il comando si ferma con un messaggio esplicito
- `-line` resta disponibile come fallback manuale per ambienti non interattivi
- raw mode Windows prova prima l'input VT e poi un fallback senza input VT
- l'output Windows richiede esplicitamente una console con VT processing

Uso consigliato:

```powershell
go run ./cmd/retronet-terminal-live -demo
```

Fallback manuale:

```powershell
go run ./cmd/retronet-terminal-live -demo -line
```

Verifica:

```powershell
go test -count=1 ./...
go vet ./...
go run ./cmd/retronet-terminal-live -width 40 -height 6 -script "CIAO`r`nREADY"
git diff --check
```
