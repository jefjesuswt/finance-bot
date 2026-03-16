package parser

import "errors"

type Rule func(t *Transaction) error

func (t *Transaction) Validate() error {
	// chequear las reglas
	rules := []Rule{
		validateTargetAmount,
		validateInterest,
		validateExchangeRates,
		validateAccountCurrencyConsistency,
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

	case CurrencyUSDCash, CurrencyEURCash, CurrencyUSDMenudeo:
		return errors.New("es obligatorio colocar la 'tasa' o el modificador 'por' al operar con efectivo o dólares de menudeo")

	case CurrencyUSDT, CurrencyBTC, CurrencyETH, CurrencyBNB, CurrencyMATIC:
		t.Warnings = append(
			t.Warnings,
			"⚠️ No colocaste la tasa para "+string(t.Currency)+". Usaré el promedio del día. Si deseas corregirlo, responde a este mensaje.",
		)
	}

	return nil
}

func validateAccountCurrencyConsistency(t *Transaction) error {
	// no hay efectivo en cuentas bancarias
	if (t.Currency == CurrencyUSDCash || t.Currency == CurrencyEURCash) && (t.Account == AccountBancamiga || t.Account == AccountBDV) {
		return errors.New("incongruencia: no puedes tener dólares o euros en físico dentro de un banco nacional")
	}

	// no hay criptos en cuentas bancarias (fiat)
	isCrypto := t.Currency == CurrencyUSDT || t.Currency == CurrencyBTC || t.Currency == CurrencyETH || t.Currency == CurrencyBNB || t.Currency == CurrencyMATIC
	if isCrypto && (t.Account == AccountBancamiga || t.Account == AccountBDV || t.Account == AccountCash) {
		return errors.New("incongruencia: las criptomonedas deben ir en 'binance' o 'coinbase'")
	}

	// no manejar menudeo/bcv en efectivo, solo admite bdv o bancamiga
	isDigitalFiat := t.Currency == CurrencyUSDBCV || t.Currency == CurrencyEURBCV || t.Currency == CurrencyUSDMenudeo
	if isDigitalFiat && t.Account == AccountCash {
		return errors.New("incongruencia: no puedes manejar divisas bancarias (bcv/menudeo) en efectivo (cash). Usa 'bdv' o 'bancamiga'")
	}

	return nil
}
