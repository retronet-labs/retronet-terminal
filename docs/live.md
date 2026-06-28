# Terminale Live

`retronet-terminal-live` e' il primo adattatore interattivo sopra il core
`terminal.Terminal`. Serve per giocare con lo schermo RetroNet senza avviare
ancora CP/M, BBS o websocket.

La logica riusabile vive nel package Go
`github.com/retronet-labs/retronet-terminal/live`: il comando locale usa quel
runner con un handler dimostrativo, mentre `retronet-cpm` e `retronet-api`
possono fornire handler diversi.

## Avvio

Dal repository:

```powershell
cd C:\work\source\retronet-terminal
go run ./cmd/retronet-terminal-live
```

Con demo iniziale:

```powershell
go run ./cmd/retronet-terminal-live -demo
```

Con dimensioni esplicite:

```powershell
go run ./cmd/retronet-terminal-live -width 100 -height 30
```

## Tasti

- testo stampabile: viene scritto nello schermo del terminale
- Invio: nuova riga e prompt `READY>`
- Backspace: cancella il carattere precedente sullo schermo
- `Ctrl+L`: pulisce lo schermo
- `Ctrl+Q`, `Ctrl+C` o `Ctrl+D`: esce

Quando possibile, il comando entra in raw mode: i tasti vengono letti subito,
senza aspettare Invio. Se la console non supporta raw mode, il comando si ferma
con un messaggio esplicito. In quel caso conviene avviarlo da PowerShell o
Windows Terminal; per una prova non interattiva si puo' usare `-script`, oppure
`-line` per una modalita' a righe meno fedele.

Il live non ridisegna tutto lo schermo a ogni tasto: invia alla console solo i
nuovi byte prodotti dal core terminale. Questo rende stabile il prompt `READY>`
e riduce lo sfarfallio.

## Modalita' Script

Per test, documentazione e CI si puo' usare `-script`: il comando alimenta il
terminale con una stringa, ridisegna lo schermo finale e termina.

```powershell
go run ./cmd/retronet-terminal-live -width 40 -height 6 -script "CIAO`r`nREADY"
```

Questa modalita' e' utile per verificare il renderer senza aprire una sessione
interattiva.

## Cosa Dimostra

Il live CLI dimostra quattro elementi del contratto:

- il core riceve byte e aggiorna lo schermo derivato
- lo snapshot e' sufficiente per ridisegnare una vista interattiva
- il renderer esterno puo' essere sostituito in futuro da websocket/xterm.js
- raw mode e repaint sono adattatori, non logica del core terminale

## Uso Come Package

Un repo applicativo fornisce un handler:

```go
type Handler interface {
    Start(term *terminal.Terminal) error
    HandleByte(term *terminal.Terminal, value byte) (bool, error)
}
```

Poi avvia il runner:

```go
err := live.Run(live.Config{
    Input:   os.Stdin,
    Output:  os.Stdout,
    Raw:     true,
    Handler: handler,
})
```

`Start` scrive lo stato iniziale sul terminale. `HandleByte` riceve un byte alla
volta e restituisce `false` quando la sessione deve terminare. Il runner si
occupa di raw mode, rendering iniziale dello snapshot, `DrainOutput` dopo ogni
tasto e ripristino della console.

## Limiti

- non e' un VT100 completo
- non avvia programmi CP/M-like da solo
- non implementa scrollback
- non apre socket
- non include ROM, font, terminfo, manuali o asset storici proprietari

La prossima integrazione naturale e' usare questo modello in
`retronet-cpm/session` e poi in `retronet-api`.
