package parser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reAmountCurrency = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s+(usd cash|usd bcv|usdt|bs|eur cash|eur bcv|btc|eth|bnb|matic|[a-z]+)`)
	reRate           = regexp.MustCompile(`(?i)tasa\s+(\d+(?:\.\d+)?)`)
	reAccount        = regexp.MustCompile(`(?i)en\s+([a-zA-Z0-9_]+)`)
	reTarget         = regexp.MustCompile(`(?i)por\s+(\d+(?:\.\d+)?)\s+([a-zA-Z0-9_\s]+)`)
	reIndex          = regexp.MustCompile(`(?i)indexado\s+([a-zA-Z0-9_]+)`)
	reInterest       = regexp.MustCompile(`(?i)interes\s+(\d+(?:\.\d+)?)`)
)

type modifiersExtractor struct {
	pattern *regexp.Regexp
	apply func(match []string, tx *Transaction)
}

func Parse(input string) (Transaction, error) {
	tx := Transaction{
		Date: time.Now(),
	}

	text := strings.TrimSpace(input)

	extractors := []modifiersExtractor{
		{reRate, func(m []string, t *Transaction) { t.ExplicitRate, _ = strconv.ParseFloat(m[1], 64) }},
		{reAccount, func(m []string, t *Transaction) { t.Account = Account(strings.ToLower(m[1])) }},
		{reIndex, func(m []string, t *Transaction) { t.IndexedTo = Currency(strings.ToLower(m[1])) }},
		{reInterest, func(m []string, t *Transaction) { t.InterestRate, _ = strconv.ParseFloat(m[1], 64) }},
		{reTarget, func(m []string, t *Transaction) {
			t.TargetAmount, _ = strconv.ParseFloat(m[1], 64)
			t.TargetCurrency = Currency(strings.ToLower(strings.TrimSpace(m[2])))
		}},
	}

	for _, extractor := range extractors {
		if match := extractor.pattern.FindStringSubmatch(text); match != nil {
			extractor.apply(match, &tx)
			text = strings.Replace(text, match[0], "", 1)
		}
	}

	// extraer monto y moneda
	matchAmmountCurrency := reAmountCurrency.FindStringSubmatch(text)
	if matchAmmountCurrency == nil {
		return tx, errors.New("no se pudo extraer monto y moneda")
	}

	amount, err := strconv.ParseFloat(matchAmmountCurrency[1], 64)
	if err != nil {
		return tx, errors.New("monto no válido")
	}

	tx.Amount = amount
	tx.Currency = Currency(strings.ToLower(matchAmmountCurrency[2]))
	text = strings.Replace(text, matchAmmountCurrency[0], "", 1)

	// concepto, acción y deudas
	words := strings.Fields(text)
	if len(words) == 0 {
		return tx, errors.New("falta acción y concepto")
	}

	tx.Action = Action(strings.ToLower(words[0]))
	if tx.Account == "" {
		tx.Account = AccountCash
	}

	switch tx.Action {
		case ActionPrestamo, ActionCobro:
			if len(words) < 3 {
				return tx, errors.New("debes especificar un deudor y un concepto (ej: prestamo mama remedios)")
			}
			tx.Debtor = strings.ToLower(words[1])
			tx.Concept = strings.Join(words[2:], " ")

			if tx.Action == ActionCobro {
				tx.Status = StatusPagado
			} else {
				tx.Status = StatusPendiente
			}

		case ActionGasto, ActionIngreso:
			if len(words) < 2 {
				return tx, errors.New("debes especificar un concepto (ej: gasto audifonos)")
			}
			tx.Concept = strings.Join(words[1:], " ")
			tx.Status = StatusNone

		case ActionCambio, ActionInversion, ActionRetorno:
			// el concepto es opcional para cambios, inversiones o retornos
			if len(words) > 1 {
				tx.Concept = strings.Join(words[1:], " ")
			} else {
				tx.Concept = string(tx.Action) // si concepto vacio, concepto = 'inversion', 'retorno' o 'cambio'
			}
			tx.Status = StatusNone

		default:
			return tx, errors.New("acción desconocida: " + string(tx.Action))
	}

	return tx, nil
}
