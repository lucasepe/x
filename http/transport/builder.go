package transport

import "net/http"

// TransportBuilder compone piu' middleware HTTP (RoundTripper wrappers)
// applicandoli nell'ordine in cui vengono dichiarati con [TransportBuilder.Use].
type TransportBuilder struct {
	layers []func(http.RoundTripper) http.RoundTripper
}

// NewTransportBuilder crea un builder vuoto pronto a comporre un transport
// partendo dal transport di default del package.
func NewTransportBuilder() *TransportBuilder {
	return &TransportBuilder{
		layers: make([]func(http.RoundTripper) http.RoundTripper, 0),
	}
}

// Use registra un nuovo layer di wrapping.
// Il primo layer aggiunto sara' anche il primo a ricevere la richiesta in fase
// di esecuzione, mentre l'ultimo aggiunto restera' piu' vicino al transport base.
// Questo rende i preset componibili: ad esempio puoi applicare [ForScraping] e
// poi chiamare [ForDebug] sullo stesso builder per aggiungere logging e request
// id sopra uno stack gia' esistente.
func (b *TransportBuilder) Use(layer func(http.RoundTripper) http.RoundTripper) *TransportBuilder {
	b.layers = append(b.layers, layer)
	return b
}

// Build costruisce la catena finale partendo da [Default] e applicando i layer
// in ordine inverso, cosi' da preservare l'ordine dichiarativo usato con [Use].
func (b *TransportBuilder) Build() http.RoundTripper {
	var rt http.RoundTripper = Default()
	for i := len(b.layers) - 1; i >= 0; i-- {
		rt = b.layers[i](rt)
	}
	return rt
}
