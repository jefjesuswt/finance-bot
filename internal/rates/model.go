package rates

type SystemRates struct {
	USDBCV 		float64
	EURBCV 		float64
	USDParallel float64
	EURParallel float64
	Binance 	float64
}

type DolarApiResponse []struct {
	Source 	string `json:"fuente"`
	Price 	float64 `json:"promedio"`
}

type BinanceP2PResponse struct {
	Data []struct{
		Adv struct {
			Price string `json:"price"`
		} `json:"adv"`
	} `json:"data"`
}
