# Release v0.4.0

Release dedicata al collegamento con `retronet-api`.

## Novita'

- nuovo comando `cmd/retronet-terminal-api`
- creazione automatica di una sessione `retronet-api`
- collegamento a sessioni esistenti con `-session`
- collegamento diretto con `-url ws://...`
- raw mode host per inviare tasti byte per byte
- modalita' `-line` per console non raw o prove manuali semplici
- modalita' `-script` per smoke test non interattivi
- package interno `internal/wsclient` con client WebSocket minimale basato solo
  su libreria standard Go
- test su handshake, frame mascherati, URL sessione e normalizzazione input
- documentazione italiana su uso e limiti

## Esempio

Avviare API:

```powershell
cd C:\work\source\retronet-api
go run ./cmd/retronet-api -addr 127.0.0.1:8080
```

Collegare terminale:

```powershell
cd C:\work\source\retronet-terminal
go run ./cmd/retronet-terminal-api -api http://127.0.0.1:8080
```

Modalita' a righe:

```powershell
go run ./cmd/retronet-terminal-api -api http://127.0.0.1:8080 -line
```

Smoke test:

```powershell
go run ./cmd/retronet-terminal-api -api http://127.0.0.1:8080 -script "HELP`r" -script-wait 1s
```

## Nota Architetturale

Il core `terminal.Terminal` non apre socket e non importa HTTP. Il nuovo comando
e' un adapter: prende input dalla console host, lo traduce in messaggi
`{"type":"input"}` e scrive su stdout i messaggi `output` ricevuti da
`retronet-api`.

## Licenze

Non sono state aggiunte dipendenze esterne e non sono stati copiati asset
storici. Il client WebSocket e' codice originale RetroNet basato sulla RFC6455,
limitato al subset necessario per il laboratorio.

## Verifica

```powershell
gofmt -l .
go test -count=1 ./...
go vet ./...
go run ./cmd/retronet-terminal -demo -screen
git diff --check
```
