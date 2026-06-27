# retronet-terminal

`retronet-terminal` e' il terminale testuale condiviso dell'ecosistema
RetroNet. Non emula una marca storica specifica e non contiene ROM, font, dump o
asset proprietari: offre solo un core ASCII/ANSI minimo, scritto da zero, utile
per emulatori, CP/M-like, BBS e futuri websocket.

## Perche' Esiste

Senza questo modulo ogni progetto tende a reinventare input queue, output buffer,
CR/LF, stato dello schermo e piccoli escape ANSI. Separarlo rende il terminale:

- indipendente da CPU, BDOS, BIOS e protocolli futuri
- testabile senza avviare emulatori
- riusabile da CLI locali e sessioni websocket
- chiaro dal punto di vista licenze: solo codice RetroNet originale

## Quick Start

```powershell
go test ./...
go run ./cmd/retronet-terminal -demo -screen
```

API minima:

```go
term := terminal.New(terminal.Config{ANSI: true})
term.QueueInputString("A")
value, _ := term.ReadByte()
_ = term.WriteByte(value)
fmt.Println(term.OutputString())
```

## Stato

- coda input byte-oriented
- output raw bufferizzato
- schermo testuale 80x24 di default
- CR, LF, backspace, tab e wrapping
- ANSI CSI minimo: clear screen, home/cursor position, clear line, attributi
  colore ignorati ma preservati nel buffer raw
- CLI demo locale

## Limiti

- Non e' un VT100 completo.
- Non include font, terminfo, ROM o documentazione storica proprietaria.
- Non implementa ancora websocket: il core e' pronto per essere adattato da
  `retronet-api`.

## Licenza

MIT.
