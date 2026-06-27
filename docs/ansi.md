# ANSI Supportato

Il supporto ANSI e' intenzionalmente minimo. Serve a rendere leggibili demo e
programmi testuali, non a dichiarare compatibilita' VT completa.

Sequenze CSI gestite:

- `ESC [ J`: pulizia da cursore a fine schermo
- `ESC [ 1 J`: pulizia da inizio schermo a cursore
- `ESC [ 2 J`: pulizia schermo completa
- `ESC [ H`: cursore in alto a sinistra
- `ESC [ r ; c H`: posizione cursore, 1-based
- `ESC [ A/B/C/D`: cursore su/giu'/destra/sinistra
- `ESC [ K`: pulizia da cursore a fine riga
- `ESC [ 1 K`: pulizia da inizio riga a cursore
- `ESC [ 2 K`: pulizia riga completa
- `ESC [ ... m`: attributi colore/stile ignorati

Tutte le sequenze restano nel buffer raw; solo lo schermo derivato applica il
subset sopra.

Sequenze sconosciute o incomplete non generano panic. Se una sequenza CSI
sconosciuta arriva completa, viene preservata nel raw output e ignorata dallo
schermo derivato.
