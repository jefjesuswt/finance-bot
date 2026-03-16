package processor

import "github.com/jefjesuswt/finance-bot/internal/reports"

type Result struct {
	Note 			reports.MarkdownNote
	Warnings 		[]string
	ExtraMessage 	string
}
