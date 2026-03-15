package rates

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type Service interface {
	GetCurrentRates() (SystemRates, error)
}

type oracleService struct {
	client *http.Client
}

func NewService(c *http.Client) *oracleService {
	return &oracleService{
		client: c,
	}
}

func (s *oracleService) GetCurrentRates() (SystemRates, error) {
	rates := SystemRates{}

	if err := s.fetchDolarApi(
		"https://ve.dolarapi.com/v1/dolares",
		&rates.USDBCV,
		&rates.USDParallel,
	); err != nil {
		return rates, err
	}

	if err := s.fetchDolarApi(
		"https://ve.dolarapi.com/v1/euros",
		&rates.EURBCV,
		&rates.EURParallel,
	); err != nil {
		return rates, err
	}

	if err := s.fetchBinanceApi(&rates.Binance); err != nil {
		return rates, err
	}

	return rates, nil
}

func (s *oracleService) fetchDolarApi(url string, bcv *float64, parallel *float64) error {
	resp, err := s.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data DolarApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	for _, item := range data {
		switch item.Source {
			case "oficial": *bcv = item.Price
			case "paralelo": *parallel = item.Price

			default: continue
		}
	}

	return nil
}

func (s *oracleService) fetchBinanceApi(binanceRate *float64) error {
	url := "https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search"
	payload := []byte(`{"page":1,"rows":5,"asset":"USDT","tradeType":"SELL","fiat":"VES","merchantCheck":false}`)

	// 1. Cambiamos a POST
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	// 2. Agregamos el Header obligatorio
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data BinanceP2PResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	if len(data.Data) == 0 {
		return errors.New("no se encontraron ofertas en binance")
	}

	var sum float64
	for _, item := range data.Data {
		price, _ := strconv.ParseFloat(item.Adv.Price, 64)
		sum += price
	}

	*binanceRate = sum / float64(len(data.Data))
	return nil
}
