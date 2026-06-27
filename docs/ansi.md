# ANSI Supportato

Il supporto ANSI e' intenzionalmente minimo. Serve a rendere leggibili demo e
programmi testuali, non a dichiarare compatibilita' VT completa.

Sequenze CSI gestite:

- `ESC [ 2 J`: pulizia schermo
- `ESC [ H`: cursore in alto a sinistra
- `ESC [ r ; c H`: posizione cursore, 1-based
- `ESC [ K`: pulizia da cursore a fine riga
- `ESC [ ... m`: attributi colore/stile ignorati

Tutte le sequenze restano nel buffer raw; solo lo schermo derivato applica il
subset sopra.
