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
```

Il modulo non importa emulatori, CP/M, API web o componenti UI. Gli adattatori
vivono nei repo che ne hanno bisogno.

## Confini

- Il terminale non conosce registri CPU, porte I/O o funzioni BDOS.
- Il buffer raw conserva i byte scritti, compresi escape ANSI.
- Lo schermo testuale interpreta solo un subset ANSI generico.
- Websocket e xterm.js saranno adattatori futuri, non dipendenze del core.

## Copyright

Il progetto non include ROM terminali, font storici, terminfo o manuali copiati.
I comportamenti implementati sono convenzioni testuali generiche e i test sono
sintetici.
