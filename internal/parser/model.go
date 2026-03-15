package parser

import "time"

type Action string
type Status string
type Currency string
type Account string

const (
	// acciones
	ActionGasto 	Action = "gasto"
	ActionIngreso   Action = "ingreso"
	ActionCambio    Action = "cambio"
	ActionPrestamo  Action = "prestamo"
	ActionCobro     Action = "cobro"
	ActionInversion Action = "inversion"
	ActionRetorno   Action = "retorno"

	// estados (deudas)
	StatusPendiente Status = "pendiente"
	StatusPagado 	Status = "pagado"
	StatusNone      Status = "" // Para transacciones que no son deudas

	// cuentas
	AccountCash 	 Account = "cash"
	AccountBinance 	 Account = "binance"
	AccountCoinbase  Account = "coinbase"
	AccountBancamiga Account = "bancamiga"

	// fiat
	CurrencyUSDCash Currency = "usd cash"
	CurrencyUSDBCV  Currency = "usd bcv"
	CurrencyEURCash Currency = "eur cash"
	CurrencyEURBCV  Currency = "eur bcv"
	CurrencyBS      Currency = "bs"

	// cripto
	CurrencyUSDT  Currency = "usdt"
	CurrencyBTC   Currency = "btc"
	CurrencyETH   Currency = "eth"
	CurrencyBNB   Currency = "bnb"
	CurrencyMATIC Currency = "matic"
)

type Transaction struct {
	Action 		Action
	Concept 	string
	Amount 		float64
	Currency 	Currency
	Account 	Account

	// modificadores
	ExplicitRate 	float64 // "tasa X"
	TargetAmount 	float64 // "por X"
	TargetCurrency 	Currency // "por TargetAmmount X",  por ejemplo "por 100 EUR"
	IndexedTo 		Currency // "indexado a X"
	InterestRate 	float64 // "interes X"

	// deudas
	Debtor string
	Status Status

	Date time.Time

	Warnings []string
}
