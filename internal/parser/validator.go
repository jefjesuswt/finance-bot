package parser

import "errors"

type Rule func(t *Transaction) error

func (t *Transaction) Validate() error {
	// chequear las reglas
	rules := []Rule{
		validateTargetAmount,
		validateInterest,
		validateExchangeRates,
	}

	// ciclar las reglas, conseguir el error
	for _, rule := range rules {
		if err := rule(t); err != nil {
			return err
		}
	}

	return nil
}

func validateTargetAmount(t *Transaction) error {
	if t.TargetAmount > 0 && (t.Action == ActionGasto || t.Action == ActionIngreso) {
		return errors.New("el modificador 'por' no tiene sentido en un gasto o ingreso. Úsalo en cambios o préstamos")
	}
	return nil
}

func validateInterest(t *Transaction) error {
	if t.InterestRate > 0 && t.Action != ActionPrestamo {
		return errors.New("el modificador 'interés' no tiene sentido en una transacción que no sea un préstamo")
	}
	return nil
}

func validateExchangeRates(t *Transaction) error {
	isTradeAction := (t.Action == ActionCambio || t.Action == ActionInversion)

	if !isTradeAction || t.ExplicitRate > 0 || t.TargetAmount > 0 {
		return nil
	}

	switch t.Currency {

	case CurrencyUSDCash, CurrencyEURCash:
		return errors.New("es obligatorio colocar la 'tasa' negociada al operar con dólares/euros en físico")

	case CurrencyUSDT, CurrencyBTC, CurrencyETH, CurrencyBNB, CurrencyMATIC:
		t.Warnings = append(
			t.Warnings,
			"⚠️ No colocaste la tasa para "+string(t.Currency)+". Usaré el promedio del día. Si deseas corregirlo, responde a este mensaje.",
		)
	}

	return nil
}
