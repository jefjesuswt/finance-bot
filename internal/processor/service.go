package processor

import (
	"context"
	"fmt"

	"github.com/jefjesuswt/finance-bot/internal/github"
	"github.com/jefjesuswt/finance-bot/internal/parser"
	"github.com/jefjesuswt/finance-bot/internal/rates"
	"github.com/jefjesuswt/finance-bot/internal/reports"
)

type Service interface {
	ProcessTransaction(ctx context.Context, text string) (*Result, error)
}

type processorService struct {
	ratesService rates.Service
	ghClient 	 *github.Client
	folder 		 string
}

func NewService(rs rates.Service, ghc *github.Client, folder string) *processorService {
	return &processorService{
		ratesService: rs,
		ghClient: ghc,
		folder: folder,
	}
}

func (s *processorService) ProcessTransaction(ctx context.Context, text string) (*Result, error) {
	tx, err := parser.Parse(text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}

	if err := tx.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate transaction: %w", err)
	}

	currentRates, err := s.ratesService.GetCurrentRates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current rates: %w", err)
	}

	note, err := reports.BuildNote(tx, currentRates, s.folder)
	if err != nil {
		return nil, fmt.Errorf("failed to build note: %w", err)
	}

	commitMsg := fmt.Sprintf("🤖 bot: registra transacción %s", tx.Action)

	if err := s.ghClient.PushFile(ctx, note, commitMsg); err != nil {
		return nil, fmt.Errorf("failed to push file: %w", err)
	}

	res := &Result{
		Note: note,
		Warnings: tx.Warnings,
	}

	if tx.Action == parser.ActionCobro {
		err := s.ghClient.MarkLoanAsPaid(ctx, note.Folder, tx.Debtor, tx.Concept)
		if err != nil {
			res.ExtraMessage = fmt.Sprintf("failed to mark loan as paid: %v", err)
		} else {
			res.ExtraMessage = fmt.Sprintf("loan marked as paid: %s", tx.Debtor)
		}
	}

	return res, nil
}
