# Terminale API

`retronet-terminal-api` collega una console locale a una sessione
`retronet-api` tramite WebSocket. E' il primo ponte pratico tra terminale,
backend API e ambiente CP/M-like.

## Avvio Rapido

Avviare `retronet-api`:

```powershell
cd C:\work\source\retronet-api
go run ./cmd/retronet-api -addr 127.0.0.1:8080
```

Collegare il terminale:

```powershell
cd C:\work\source\retronet-terminal
go run ./cmd/retronet-terminal-api -api http://127.0.0.1:8080
```

Se `-session` non viene indicato, il comando crea una nuova sessione API e si
collega al suo websocket.

## Modalita' A Righe

Quando raw mode non e' disponibile, oppure quando si vuole una prova piu'
prevedibile:

```powershell
go run ./cmd/retronet-terminal-api -api http://127.0.0.1:8080 -line
```

In questa modalita' ogni riga digitata viene inviata come riga CP/M-like
terminata da carriage return.

## Sessione Esistente

Per creare la sessione manualmente:

```powershell
$s = Invoke-RestMethod -Method Post http://127.0.0.1:8080/sessions
go run ./cmd/retronet-terminal-api -api http://127.0.0.1:8080 -session $s.id
```

Per passare direttamente l'URL websocket:

```powershell
go run ./cmd/retronet-terminal-api `
  -url "ws://127.0.0.1:8080/sessions/$($s.id)/ws"
```

## Script Breve

`-script` invia byte al websocket e poi termina dopo `-script-wait`.
Serve per smoke test e demo non interattive:

```powershell
go run ./cmd/retronet-terminal-api -api http://127.0.0.1:8080 `
  -script "HELP`r" `
  -script-wait 1s
```

## Protocollo

Il comando usa il protocollo JSON di `retronet-api v0.2`:

Client -> server:

```json
{"type":"input","data":"DIR\r"}
```

Server -> client:

```json
{"type":"output","data":"A>DIR\r\n"}
{"type":"state","state":"idle"}
{"type":"snapshot","snapshot":{}}
```

Il terminale locale stampa i messaggi `output` come byte raw. I messaggi
`state` e `snapshot` vengono usati solo per capire quando la sessione e' chiusa;
la futura UI web potra' invece sfruttare gli snapshot per ridisegnare lo schermo.

## Perche' E' Utile

Questo comando separa tre responsabilita':

- `retronet-api` gestisce sessioni, drive temporanei e CP/M-like
- `retronet-terminal-api` gestisce console host e trasporto WebSocket
- il core `retronet-terminal` resta un componente byte-oriented riusabile

Lo stesso schema potra' essere usato da altri emulatori RetroNet. Per un futuro
`retronet-pc`, per esempio, l'adapter dovra' tradurre tra terminale/API e video
testuale MDA piu' tastiera; per un 8086 nudo potra' bastare un adapter seriale o
I/O byte-oriented.

## Limiti

- niente autenticazione o TLS lato client locale
- websocket minimale, senza frame frammentati
- niente upload file: il drive resta gestito da `retronet-api`
- il core terminale non diventa un emulatore VT completo
- nessuna ROM, font, terminfo, manuale o asset storico viene incluso
