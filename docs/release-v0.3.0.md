# Release v0.3.0

Release dedicata all'estrazione del runner live riusabile.

Novita':

- nuovo package `github.com/retronet-labs/retronet-terminal/live`
- interfaccia `live.Handler` con `Start` e `HandleByte`
- `live.Run` per raw mode, rendering iniziale, output a delta e modalita'
  scriptata
- `live.RenderSnapshot`, `live.WriteDelta` e `live.FeedBytes` testabili senza
  console interattiva
- raw adapter Windows/Linux spostati dal comando al package
- `cmd/retronet-terminal-live` ridotto a handler dimostrativo `READY>`

Perche' serve:

- `retronet-cpm` puo' usare lo stesso raw mode/repaint senza copiare codice
- `retronet-api` potra' riusare il contratto handler anche sopra websocket
- il core `terminal.Terminal` resta indipendente da CP/M, BBS e trasporti

Uso minimo:

```go
err := live.Run(live.Config{
    Input:   os.Stdin,
    Output:  os.Stdout,
    Raw:     true,
    Handler: handler,
})
```

Verifica:

```powershell
go test -count=1 ./...
go vet ./...
go run ./cmd/retronet-terminal-live -width 40 -height 6 -script "CIAO`r`nREADY"
git diff --check
```

Licenza e provenienza:

- nessuna ROM, font, terminfo, manuale storico copiato o dump incluso
- nessuna dipendenza esterna aggiunta
- comportamento ASCII/ANSI generico scritto da zero
