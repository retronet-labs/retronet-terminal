# Contratto Del Terminale

`retronet-terminal` e' un terminale testuale byte-oriented. Il suo compito e'
stare tra un programma che produce/consuma byte e un adattatore esterno, per
esempio CLI, shell CP/M-like, BBS o websocket.

## Cosa Garantisce

- input queue FIFO, letta con `ReadByte` o `Read`
- output raw preservato byte per byte
- `DrainOutput`, per consegnare al chiamante solo i nuovi byte e poi svuotare il
  buffer raw
- schermo testuale derivato, con dimensioni configurabili
- `Snapshot`, vista immutabile con righe a larghezza fissa, cursore, dimensioni,
  input pendente e byte raw in attesa
- `Resize`, che preserva il contenuto visibile in alto a sinistra e sposta il
  cursore dentro i nuovi limiti
- subset ANSI documentato, senza dichiarare compatibilita' VT completa

## Cosa Non Garantisce

- non e' un VT100 completo
- non interpreta terminfo
- non contiene font o ROM terminali
- non implementa scrollback storico
- non conosce CPU, BDOS, BIOS, BBS o websocket

## Esempio API

```go
term := terminal.New(terminal.Config{Width: 80, Height: 24, ANSI: true})
term.QueueInputString("DIR\r")

_, _ = term.Write([]byte("A>DIR\r\nHELLO.COM\r\n"))

snapshot := term.Snapshot()
for _, row := range snapshot.Rows {
    fmt.Println(row)
}

delta := term.DrainOutput()
_ = delta // byte da inviare a una CLI o websocket
```

## Uso Futuro Con Websocket

Un server websocket puo' usare due flussi:

- client -> server: tasti o byte ricevuti dal browser, accodati con
  `QueueInput`
- server -> client: byte prodotti dal programma, letti con `DrainOutput`, piu'
  snapshot periodici per riallineare lo schermo

Il core terminale resta indipendente dal trasporto: non apre socket, non parla
HTTP e non importa librerie UI.

## Licenze E Provenienza

Il comportamento implementato e' un subset generico ASCII/ANSI scritto da zero.
Il repo non include ROM, font, terminfo, manuali storici copiati o dump binari.
