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
```

Il modulo non importa emulatori, CP/M, API web o componenti UI. Gli adattatori
vivono nei repo che ne hanno bisogno.

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

## Confini

- Il terminale non conosce registri CPU, porte I/O o funzioni BDOS.
- Il buffer raw conserva i byte scritti, compresi escape ANSI.
- Lo schermo testuale interpreta solo un subset ANSI generico.
- `Snapshot` e `DrainOutput` sono il contratto pensato per CLI e websocket.
- Il live CLI e' un adattatore locale sopra lo snapshot, non una dipendenza del
  core.
- Il package `live` non conosce CP/M, BBS o API: conosce solo un handler di byte.
- Websocket e xterm.js saranno adattatori futuri, non dipendenze del core.

## Copyright

Il progetto non include ROM terminali, font storici, terminfo o manuali copiati.
I comportamenti implementati sono convenzioni testuali generiche e i test sono
sintetici.
