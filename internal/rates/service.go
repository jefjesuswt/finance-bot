package rates

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"golang.org/x/sync/errgroup"
)

const (
	urlDolarAPIUSD = "https://ve.dolarapi.com/v1/dolares"
	urlDolarAPIEUR = "https://ve.dolarapi.com/v1/euros"
	urlBinanceP2P  = "https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search"
	payloadBinance = `{"page":1,"rows":5,"asset":"USDT","tradeType":"SELL","fiat":"VES","merchantCheck":false}`
)

type Service interface {
	GetCurrentRates(ctx context.Context) (SystemRates, error)
}

type oracleService struct {
	client *http.Client
}

func NewService(c *http.Client) *oracleService {
	return &oracleService{
		client: c,
	}
}

func (s *oracleService) GetCurrentRates(ctx context.Context) (SystemRates, error) {
	rates := SystemRates{}

	g, gCtx := errgroup.WithContext(ctx)

	// dolar
	g.Go(func() error {
		return s.fetchDolarApi(
			gCtx,
			urlDolarAPIUSD,
			&rates.USDBCV,
			&rates.USDParallel,
		)
	})

	// euro
	g.Go(func() error {
		return s.fetchDolarApi(
			gCtx,
			urlDolarAPIUSD,
			&rates.EURBCV,
			&rates.EURParallel,
		)
	})

	// binance
	g.Go(func() error {
		return s.fetchBinanceApi(gCtx, &rates.Binance)
	})

	if err := g.Wait(); err != nil {
		return SystemRates{}, nil
	}

	return rates, nil
}

func (s *oracleService) fetchDolarApi(
	ctx context.Context,
	url string,
	bcv *float64,
	parallel *float64,
) error {
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

func (s *oracleService) fetchBinanceApi(ctx context.Context, binanceRate *float64) error {

	// 1. Cambiamos a POST
	req, err := http.NewRequest(http.MethodPost, urlBinanceP2P, bytes.NewBuffer([]byte(payloadBinance)))
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
