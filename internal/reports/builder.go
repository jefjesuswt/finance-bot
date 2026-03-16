package reports

import (
	"fmt"
	"strings"

	"github.com/jefjesuswt/finance-bot/internal/parser"
	"github.com/jefjesuswt/finance-bot/internal/rates"
)

func BuildNote(tx parser.Transaction, r rates.SystemRates, folder string) (MarkdownNote, error) {
	var sb strings.Builder

	//inicio
	fmt.Fprintln(&sb, "---")
	fmt.Fprintf(&sb, "tipo: %s\n", tx.Action)

	// deudas
	if tx.Debtor != "" {
		fmt.Fprintf(&sb, "deudor: %s\n", tx.Debtor)
	}
	if tx.Status != "" {
		fmt.Fprintf(&sb, "estado: %s\n", tx.Status)
	}

	// detales
	fmt.Fprintf(&sb, "concepto: \"%s\"\n", tx.Concept)
	fmt.Fprintf(&sb, "monto: %.2f\n", tx.Amount)
	fmt.Fprintf(&sb, "moneda: %s\n", tx.Currency)
	fmt.Fprintf(&sb, "cuenta: %s\n", tx.Account)

	// fecha y hora
	fmt.Fprintf(&sb, "fecha: %s\n", tx.Date.Format("2006-01-02"))
	fmt.Fprintf(&sb, "hora: %s\n", tx.Date.Format("15:04:05"))

	// economics
	fmt.Fprintf(&sb, "tasa_oficial_bcv: %.4f\n", r.USDBCV)
	fmt.Fprintf(&sb, "tasa_binance_p2p: %.2f\n", r.Binance)
	fmt.Fprintf(&sb, "tasa_paralelo: %.2f\n", r.USDParallel)

	bsAmount := calculateBsAmount(tx, r)

	if bsAmount > 0 {
		fmt.Fprintf(&sb, "monto_bs_equivalente: %.2f\n", bsAmount)
		fmt.Fprintf(&sb, "monto_usd_bcv_equivalente: %.2f\n", bsAmount/r.USDBCV)
		fmt.Fprintf(&sb, "monto_usdt_equivalente: %.2f\n", bsAmount/r.Binance)
	}

	if tx.Action == parser.ActionPrestamo {
		fmt.Fprintln(&sb, "\n# --- Estado de Deuda ---")

		if tx.IndexedTo != "" {
			fmt.Fprintf(&sb, "indexado_a: %s\n", tx.IndexedTo)
		}
		if tx.InterestRate > 0 {
			fmt.Fprintf(&sb, "interes_aplicado: %.2f\n", tx.InterestRate)
		}

		// metrica para el dataview
		debtInUSDT := calculateDebtUSDT(tx, bsAmount, r)
		fmt.Fprintf(&sb, "deuda_activa_usdt: %.2f\n", debtInUSDT)
	}

	fmt.Fprintln(&sb, "---")

	safeConcept := strings.ReplaceAll(strings.ToLower(tx.Concept), " ", "-")
	if len(safeConcept) > 20 {
		safeConcept = safeConcept[:20]
	}

	var fileName string
	if tx.Debtor != "" {
		fileName = fmt.Sprintf("%s-%s-%s-%s.md", tx.Date.Format("20060102-150405"), tx.Action, tx.Debtor, safeConcept)
	} else {
		fileName = fmt.Sprintf("%s-%s-%s.md", tx.Date.Format("20060102-150405"), tx.Action, safeConcept)
	}

	return MarkdownNote{
		Filename: fileName,
		Folder: folder,
		Content: sb.String(),
	}, nil
}

func calculateBsAmount(tx parser.Transaction, r rates.SystemRates) float64 {
	if tx.Currency == parser.CurrencyBS {
		return tx.Amount
	}

	// al usar 'por', se olvida el resto
	if tx.TargetAmount > 0 && tx.TargetCurrency == parser.CurrencyBS {
		return tx.TargetAmount
	}

	var rate float64
	if tx.ExplicitRate > 0 {
		rate = tx.ExplicitRate // al forzar la tasa
	} else {
		rate = getDefaultRate(tx.Currency, r) // si no, scrapeamos
	}

	return tx.Amount * rate
}

func getDefaultRate(c parser.Currency, r rates.SystemRates) float64 {
	switch c {
	case parser.CurrencyUSDBCV:
		return r.USDBCV
	case parser.CurrencyEURBCV:
		return r.EURBCV
	case parser.CurrencyUSDCash, parser.CurrencyUSDMenudeo:
		return r.USDParallel
	case parser.CurrencyEURCash:
		return r.EURParallel
	case parser.CurrencyUSDT:
		return r.Binance
	default:
		return 0 // criptos sin target
	}
}

func calculateDebtUSDT(tx parser.Transaction, bsAmount float64, r rates.SystemRates) float64 {
	var debt float64

	// capital base
	if tx.TargetCurrency == parser.CurrencyUSDT && tx.TargetAmount > 0 {
		debt = tx.TargetAmount
	} else if tx.Currency == parser.CurrencyUSDT {
		debt = tx.Amount
	} else if bsAmount > 0 {
		debt = bsAmount / r.Binance
	}

	// tasa de interes en caso de que aplique
	if tx.InterestRate > 0 {
		debt = debt * (1 + (tx.InterestRate / 100))
	}

	return debt
}
