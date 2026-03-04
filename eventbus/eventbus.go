package eventbus

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Package eventbus fornisce un event bus in-memory, leggero e thread-safe
// per coordinare componenti interni di un'applicazione.
//
// Il package resta volutamente generico: non impone domini, trasporto,
// persistenza o semantiche applicative. Gli eventi vengono instradati per topic
// (EventID) e il payload resta completamente a carico del chiamante.
//
// Il bus supporta:
//   - subscribe/unsubscribe concorrenti;
//   - publish sincrono o asincrono;
//   - deadline e cancellazione cooperative via context.Context;
//   - raccolta degli errori restituiti dagli handler;
//   - intercettazione dei panic tramite hook opzionale.
//
// Il timeout configurato sul bus non interrompe forzatamente gli handler:
// limita soltanto il tempo massimo durante il quale il publisher attende il
// completamento dei subscriber. Per rendere efficace la deadline, gli handler
// devono osservare ctx.Done() e terminare in modo cooperativo.
//
// EventID identifica il topic di instradamento di un evento.
type EventID string

// Event rappresenta qualunque payload pubblicabile sul bus.
// Ogni evento dichiara il topic tramite EventID.
type Event interface {
	EventID() EventID
}

// EventHandler è la funzione invocata per ogni evento consegnato a un
// subscriber.
//
// Il context ricevuto eredita quello passato al publish e può includere una
// deadline applicata dal bus. Il valore di ritorno viene raccolto nel
// PublishResult e, in caso di errore, può essere osservato anche tramite
// FailureHook.
type EventHandler func(ctx context.Context, event Event) error

// Subscription identifica una sottoscrizione attiva.
// Il valore ritornato da Subscribe va conservato se si vuole poi eseguire
// Unsubscribe in modo preciso.
type Subscription struct {
	eventID EventID
	id      uint64
}

// PublishResult riassume l'esito di una publish.
//
// Delivered conta gli handler che hanno restituito un risultato prima della
// chiusura del contesto di publish. Pending conta gli handler ancora in corso
// quando il contesto è scaduto o è stato cancellato. Errors contiene gli errori
// restituiti dagli handler completati (inclusi i panic convertiti in error).
// Err rappresenta invece l'errore "globale" della publish, tipicamente dovuto a
// timeout o cancellazione del context.
type PublishResult struct {
	Delivered int
	Pending   int
	Errors    []error
	Err       error
}

// HandlerFailure descrive il fallimento di un singolo subscriber.
// Viene emesso quando un handler restituisce un errore oppure genera un panic.
type HandlerFailure struct {
	Subscription Subscription
	Event        Event
	Err          error
	Panic        any
}

// FailureHook osserva errori e panic dei subscriber.
// L'hook viene invocato nel goroutine che gestisce il subscriber fallito.
type FailureHook func(HandlerFailure)

// BusSubscriber espone le operazioni di subscribe e unsubscribe.
type BusSubscriber interface {
	Subscribe(eventID EventID, cb EventHandler) Subscription
	Unsubscribe(id Subscription)
}

// BusPublisher espone le operazioni di pubblicazione.
//
// PublishSync attende la conclusione della publish (o la scadenza del context).
// PublishAsync avvia la publish e restituisce immediatamente un canale dal quale
// leggere il risultato finale.
type BusPublisher interface {
	PublishSync(ctx context.Context, event Event) PublishResult
	PublishAsync(ctx context.Context, event Event) <-chan PublishResult
}

// Bus combina sottoscrizione e pubblicazione in un'unica interfaccia.
type Bus interface {
	BusSubscriber
	BusPublisher
}

// Option configura il comportamento del bus alla creazione.
type Option func(*bus)

// WithPublishTimeout imposta un timeout massimo per ogni publish.
//
// Il timeout non interrompe forzatamente gli handler già avviati: limita solo
// il tempo di attesa del publisher e viene applicato tramite context.WithTimeout.
// Per ottenere una terminazione puntuale, gli handler devono rispettare il
// context ricevuto.
func WithPublishTimeout(timeout time.Duration) Option {
	return func(b *bus) {
		b.publishTimeout = timeout
	}
}

// WithFailureHook registra un hook chiamato quando un subscriber fallisce.
//
// L'hook osserva sia gli errori restituiti dagli handler, sia i panic
// intercettati dal bus e convertiti in error.
func WithFailureHook(hook FailureHook) Option {
	return func(b *bus) {
		b.failureHook = hook
	}
}

// New crea un nuovo event bus in-memory.
//
// Se non vengono passate opzioni, il bus non applica timeout di default e non
// installa alcun hook di failure.
func New(opts ...Option) Bus {
	b := &bus{
		infos: make(map[EventID]subscriptionInfoList),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(b)
		}
	}
	return b
}

type subscriptionInfo struct {
	id uint64
	cb EventHandler
}

type subscriptionInfoList []*subscriptionInfo

type bus struct {
	lock           sync.Mutex
	nextID         uint64
	publishTimeout time.Duration
	failureHook    FailureHook
	infos          map[EventID]subscriptionInfoList
}

// Subscribe registra un handler per uno specifico topic e restituisce un token
// di subscription riutilizzabile per l'unsubscribe.
//
// Il callback non può essere nil; in quel caso il metodo va in panic, perché si
// tratta di un errore di programmazione rilevabile subito.
func (bus *bus) Subscribe(eventID EventID, cb EventHandler) Subscription {
	if cb == nil {
		panic("eventbus: nil handler")
	}

	bus.lock.Lock()
	defer bus.lock.Unlock()
	id := bus.nextID
	bus.nextID++
	sub := &subscriptionInfo{
		id: id,
		cb: cb,
	}
	bus.infos[eventID] = append(bus.infos[eventID], sub)
	return Subscription{
		eventID: eventID,
		id:      id,
	}
}

// Unsubscribe rimuove una subscription esistente.
//
// Se la subscription non esiste più o non appartiene a questo bus, la chiamata
// non produce effetti.
func (bus *bus) Unsubscribe(subscription Subscription) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	if infos, ok := bus.infos[subscription.eventID]; ok {
		for idx, info := range infos {
			if info.id == subscription.id {
				infos = append(infos[:idx], infos[idx+1:]...)
				break
			}
		}
		if len(infos) == 0 {
			delete(bus.infos, subscription.eventID)
		} else {
			bus.infos[subscription.eventID] = infos
		}
	}
}

// PublishSync pubblica un evento e attende il completamento della delivery.
//
// Il metodo blocca fino a quando tutti gli handler hanno terminato oppure il
// context passato (eventualmente combinato con il timeout del bus) scade.
// Se event è nil il metodo va in panic, perché anche questo è un errore di
// programmazione.
func (bus *bus) PublishSync(ctx context.Context, event Event) PublishResult {
	return <-bus.PublishAsync(ctx, event)
}

// PublishAsync pubblica un evento e restituisce immediatamente un canale che
// produrrà il PublishResult finale.
//
// Ogni subscriber viene eseguito in un goroutine separato. Questo rende la
// consegna concorrente e quindi non garantisce un ordine deterministico di
// completamento tra handler diversi.
func (bus *bus) PublishAsync(ctx context.Context, event Event) <-chan PublishResult {
	if event == nil {
		panic("eventbus: nil event")
	}

	infos := bus.copySubscriptions(event.EventID())
	resultCh := make(chan PublishResult, 1)
	if len(infos) == 0 {
		resultCh <- PublishResult{}
		close(resultCh)
		return resultCh
	}

	pubCtx, cancel := bus.publishContext(ctx)
	results := make(chan error, len(infos))

	for _, info := range infos {
		info := info
		// Ogni handler gira in modo indipendente così un subscriber lento non
		// impedisce agli altri di partire immediatamente.
		go bus.invokeHandler(pubCtx, event, info, results)
	}

	go func() {
		defer close(resultCh)
		if cancel != nil {
			defer cancel()
		}

		result := PublishResult{}
		remaining := len(infos)

		for remaining > 0 {
			select {
			case err := <-results:
				remaining--
				result.Delivered++
				if err != nil {
					result.Errors = append(result.Errors, err)
				}
			case <-pubCtx.Done():
				// Alla scadenza smettiamo di attendere, ma gli handler già avviati
				// continuano finché non terminano o non rispettano ctx.Done().
				result.Pending = remaining
				result.Err = pubCtx.Err()
				resultCh <- result
				return
			}
		}

		resultCh <- result
	}()

	return resultCh
}

// invokeHandler esegue un singolo subscriber, converte gli eventuali panic in
// error e notifica sempre il risultato sul canale interno di raccolta.
func (bus *bus) invokeHandler(
	ctx context.Context,
	event Event,
	info *subscriptionInfo,
	results chan<- error,
) {
	var err error
	var panicValue any

	defer func() {
		if recovered := recover(); recovered != nil {
			panicValue = recovered
			err = fmt.Errorf("eventbus: handler panic for %q: %v", event.EventID(), recovered)
		}
		if err != nil {
			bus.handleFailure(HandlerFailure{
				Subscription: Subscription{
					eventID: event.EventID(),
					id:      info.id,
				},
				Event: event,
				Err:   err,
				Panic: panicValue,
			})
		}
		results <- err
	}()

	err = info.cb(ctx, event)
}

// handleFailure inoltra il fallimento all'hook registrato, se presente.
func (bus *bus) handleFailure(failure HandlerFailure) {
	if bus.failureHook != nil {
		bus.failureHook(failure)
	}
}

// publishContext combina il context del chiamante con l'eventuale timeout
// configurato sul bus.
func (bus *bus) publishContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if bus.publishTimeout <= 0 {
		return parent, nil
	}
	return context.WithTimeout(parent, bus.publishTimeout)
}

// copySubscriptions crea una snapshot dei subscriber registrati per il topic.
//
// La copia è necessaria perché codice esterno può fare subscribe/unsubscribe
// mentre una publish è in corso: iterare direttamente sullo slice condiviso
// esporrebbe a race e a comportamenti non deterministici.
func (bus *bus) copySubscriptions(eventID EventID) subscriptionInfoList {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	if infos, ok := bus.infos[eventID]; ok {
		cloned := make(subscriptionInfoList, len(infos))
		copy(cloned, infos)
		return cloned
	}
	return subscriptionInfoList{}
}
