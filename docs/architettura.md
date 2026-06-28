# Architettura

`retronet-terminal` contiene un solo concetto centrale: un terminale testuale
byte-oriented con input, output raw e schermo derivato.

```text
programma / emulatore / BBS
        |
        v
 terminal.Terminal
        |
        +-- input queue
        +-- output raw
        +-- screen buffer
        +-- snapshot

 live.Run
        |
        +-- raw mode host
        +-- rendering snapshot
        +-- delta output
        +-- handler applicativo

 retronet-terminal-api
        |
        +-- WebSocket client minimale
        +-- input host -> messaggi input
        +-- output API -> stdout raw
```

Il core terminale non importa emulatori, CP/M, API web o componenti UI. Gli
adattatori vivono nei comandi o nei repo che ne hanno bisogno.

## Adattatori CLI

`cmd/retronet-terminal` e' una demo deterministica: scrive byte, stampa raw output
o schermo e termina. `cmd/retronet-terminal-live` e' invece un adattatore
interattivo locale: mette la console in raw mode quando il sistema lo permette,
legge tasti byte per byte, applica input, backspace e piccoli controlli, poi
ridisegna lo snapshot del core con sequenze ANSI generiche.

Il comando live non cambia il core: usa `Terminal.Write`, `Snapshot` e il parser
ANSI gia presenti. Questo e' importante per il futuro websocket, perche' la UI
web potra' usare lo stesso contratto senza duplicare la logica dello schermo.

Da v0.3.0 la parte riusabile vive nel package `live`. Il comando
`cmd/retronet-terminal-live` fornisce solo un handler dimostrativo con prompt
`READY>`, mentre altri repo possono passare handler propri. Per esempio
`retronet-cpm` puo' trasformare Invio in `session.RunCommand`, Backspace in echo
locale e `Ctrl+L` in pulizia schermo.

Da v0.4.0 `cmd/retronet-terminal-api` collega la console host a
`retronet-api`: i tasti diventano messaggi websocket `input`, mentre i messaggi
`output` dell'API vengono scritti direttamente su stdout. Questo comando e' un
adapter di trasporto: non cambia il core `Terminal` e non introduce dipendenze
esterne.

## Confini

- Il terminale non conosce registri CPU, porte I/O o funzioni BDOS.
- Il buffer raw conserva i byte scritti, compresi escape ANSI.
- Lo schermo testuale interpreta solo un subset ANSI generico.
- `Snapshot` e `DrainOutput` sono il contratto pensato per CLI e websocket.
- Il live CLI e' un adattatore locale sopra lo snapshot, non una dipendenza del
  core.
- Il package `live` non conosce CP/M, BBS o API: conosce solo un handler di byte.
- Websocket e xterm.js sono adattatori, non dipendenze del core.
- Il client websocket usa solo libreria standard Go e il protocollo JSON di
  `retronet-api`.

## Copyright

Il progetto non include ROM terminali, font storici, terminfo o manuali copiati.
I comportamenti implementati sono convenzioni testuali generiche e i test sono
sintetici.
