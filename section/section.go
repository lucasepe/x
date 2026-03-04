// Package section fornisce un parser minimale per file di configurazione
// testuali organizzati in sezioni, con una sintassi ispirata ai file INI.
//
// Il package ha un obiettivo volutamente ristretto: individua e separa le
// sezioni, preservando le righe di contenuto così come appaiono nel file,
// senza interpretare key/value, tipi, quoting o interpolazioni.
//
// Questa scelta lascia al chiamante il pieno controllo sul parsing del
// contenuto di ogni sezione. In pratica, il package agisce come "section
// splitter" affidabile e leggero, utile quando formati diversi condividono la
// stessa struttura a blocchi ma richiedono regole di parsing differenti.
package section

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Section espone il contenuto di un documento organizzato in sezioni.
//
// Le sezioni vengono chiamate "unit" per mantenere l'API neutrale rispetto al
// dominio. Una unit corrisponde a una sezione del file; inoltre esiste una
// sezione implicita di root che raccoglie le righe trovate prima della prima
// intestazione esplicita.
type Section interface {
	// Units restituisce i nomi delle sezioni esplicite, nell'ordine in cui sono
	// state incontrate nel file. La sezione implicita di root non è inclusa.
	Units() []string
	// Content restituisce le righe raw associate a una sezione.
	//
	// Se unit è vuoto o contiene solo spazi, viene usata la sezione di root.
	// Il package non interpreta le righe: vengono restituite così come lette,
	// preservando spazi e formattazione originali.
	Content(unit string) []string
	// Has indica se una sezione contiene almeno una entry registrata nella
	// struttura interna del parser.
	//
	// Se unit è vuoto o contiene solo spazi, il controllo viene eseguito sulla
	// sezione di root.
	Has(unit string) bool
}

// Parse legge un flusso testuale e costruisce una Section organizzata per
// sezioni.
//
// Regole supportate:
//   - righe vuote ignorate;
//   - commenti full-line che iniziano con '#' o ';' ignorati;
//   - intestazioni di sezione nel formato "[section]";
//   - tutte le altre righe vengono conservate come contenuto raw della sezione
//     corrente.
//
// Le righe prima della prima sezione esplicita appartengono alla sezione
// implicita di root.
//
// Il parser non esegue parsing di key/value: ogni riga utile viene lasciata al
// chiamante, che potrà poi interpretarla come preferisce.
func Parse(r io.Reader) (Section, error) {
	scanner := bufio.NewScanner(r)

	p := &configParser{
		data:  make(map[string][]string),
		order: []string{rootUnit}, // sezione implicita di default
	}

	current := rootUnit

	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		// Le righe vuote non vengono preservate perché qui interessa solo la
		// struttura logica per sezioni, non la formattazione originale completa.
		if line == "" {
			continue
		}

		// I commenti full-line vengono scartati. I commenti inline, invece, fanno
		// parte del contenuto perché la loro interpretazione dipende dal formato
		// che il chiamante applicherà alla singola sezione.
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Un'intestazione valida cambia la sezione corrente.
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.TrimSpace(line[1 : len(line)-1])

			if section == "" {
				return nil, fmt.Errorf("empty section name")
			}

			if _, exists := p.data[section]; exists {
				return nil, fmt.Errorf("duplicate section [%s]", section)
			}

			current = section
			p.data[section] = []string{}
			p.order = append(p.order, section)
			continue
		}

		// Qualunque altra riga viene conservata così com'è, senza parsing
		// semantico. Questo è il punto chiave del design del package.
		p.data[current] = append(p.data[current], raw)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return p, nil
}

const (
	rootUnit = "__ROOT__"
)

var _ Section = (*configParser)(nil)

type configParser struct {
	order []string
	data  map[string][]string
}

// Units restituisce i nomi delle sole sezioni esplicite, preservando l'ordine
// di definizione nel file sorgente.
func (p *configParser) Units() []string {
	out := make([]string, 0, len(p.order))
	for _, u := range p.order {
		if u != rootUnit {
			out = append(out, u)
		}
	}
	return out
}

// Content restituisce le righe associate alla sezione indicata.
//
// Una stringa vuota seleziona la sezione implicita di root.
func (p *configParser) Content(unit string) []string {
	if strings.TrimSpace(unit) == "" {
		unit = rootUnit
	}

	return p.data[unit]
}

// Has indica se la sezione richiesta è presente nella mappa interna.
//
// Una stringa vuota seleziona la sezione implicita di root.
func (p *configParser) Has(unit string) bool {
	if strings.TrimSpace(unit) == "" {
		unit = rootUnit
	}

	_, ok := p.data[unit]
	return ok
}
