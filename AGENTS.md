# Contesto operativo per agenti

## Obiettivo

Implementare e mantenere `retronet-terminal`: terminale testuale riusabile da
emulatori, CP/M-like, BBS, API websocket e UI future.

## Decisioni da preservare

- Nessuna dipendenza da CPU, BDOS, BIOS, BBS o API web.
- Nessuna ROM, font, terminfo, manuale storico o asset proprietario incluso.
- ANSI base scritto da zero e documentato come subset, non come VT completo.
- API byte-oriented: input queue, output raw, screen buffer derivato.
- Documentazione pubblica in italiano.

## Verifica

```powershell
go test -count=1 ./...
go vet ./...
go run ./cmd/retronet-terminal -demo -screen
git diff --check
```
