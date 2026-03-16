package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/jefjesuswt/finance-bot/internal/reports"
)

type Client struct {
	HTTPClient *http.Client
	Token      string
	Owner      string
	Repo       string
	BaseURL    string
}

type fileReq struct {
	Message string `json:"message"`
	Content string `json:"content"`
	Sha     string `json:"sha,omitempty"`
}

type fileInfoRes struct {
	Sha     string `json:"sha"`
	Content string `json:"content"`
	Path    string `json:"path"`
	Name    string `json:"name"`
}


func NewClient(client *http.Client, token, owner, repo string) *Client {
	return &Client{
		HTTPClient: client,
		Token: token,
		Owner: owner,
		Repo: repo,
		BaseURL: "https://api.github.com",
	}
}

func (c *Client) PushFile(ctx context.Context, note reports.MarkdownNote, commitMsg string) error {
	fullPath := fmt.Sprintf("%s/%s", note.Folder, note.Filename)
	escapedPath := escapeGitPath(fullPath)
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.BaseURL, c.Owner, c.Repo, escapedPath)

	sha, err := c.GetFileSha(ctx, fullPath)
	if err != nil {
		return fmt.Errorf("error verificando estado del archivo: %w", err)
	}

	encodedContent := base64.StdEncoding.EncodeToString([]byte(note.Content))
	payload := fileReq{
		Message: commitMsg,
		Content: encodedContent,
		Sha:     sha,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializando payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error ejecutando request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code inesperado de github al empujar: %d", res.StatusCode)
	}

	return nil
}

func (c *Client) GetFileSha(ctx context.Context, path string) (string, error) {
	escapedPath := escapeGitPath(path)
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.BaseURL, c.Owner, c.Repo, escapedPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("no se pudo crear el request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return "", nil // Archivo no existe
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status inesperado leyendo sha: %d", res.StatusCode)
	}

	var info fileInfoRes
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return "", err
	}

	return info.Sha, nil
}

// busca un prestamo pendiente por deudor y concepto, y lo marca como pagado
func (c *Client) MarkLoanAsPaid(ctx context.Context, folder, debtor, concept string) error {
	escapedFolder := escapeGitPath(folder)
	listUrl := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.BaseURL, c.Owner, c.Repo, escapedFolder)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var files []fileInfoRes
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return err
	}

	// buscar el archivo del préstamo
	var targetFile *fileInfoRes
	searchStr := strings.ToLower(fmt.Sprintf("prestamo-%s", debtor))
	safeConcept := strings.ReplaceAll(strings.ToLower(concept), " ", "-")

	for _, file := range files {
		if strings.Contains(strings.ToLower(file.Name), searchStr) && strings.Contains(strings.ToLower(file.Name), safeConcept) {
			targetFile = &file
			break
		}
	}

	if targetFile == nil {
		return fmt.Errorf("no se encontró préstamo previo para '%s' con concepto '%s'", debtor, concept)
	}

	// bajar el contenido exacto para editarlo
	fileUrl := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.BaseURL, c.Owner, c.Repo, targetFile.Path)
	reqFile, err := http.NewRequestWithContext(ctx, http.MethodGet, fileUrl, nil)
	if err != nil {
		return err
	}
	reqFile.Header.Set("Authorization", "Bearer "+c.Token)

	respFile, err := c.HTTPClient.Do(reqFile)
	if err != nil {
		return err
	}
	defer respFile.Body.Close()

	var exactFile fileInfoRes
	if err := json.NewDecoder(respFile.Body).Decode(&exactFile); err != nil {
		return err
	}

	cleanBase64 := strings.ReplaceAll(exactFile.Content, "\n", "")

	decodedContent, err := base64.StdEncoding.DecodeString(cleanBase64)
	if err != nil {
		return err
	}
	fileText := string(decodedContent)

	if !strings.Contains(fileText, "estado: pendiente") {
		return fmt.Errorf("el préstamo ya estaba pagado o no tiene estado pendiente")
	}

	newText := strings.Replace(fileText, "estado: pendiente", "estado: pagado", 1)
	newEncodedContent := base64.StdEncoding.EncodeToString([]byte(newText))

	// Empujar la actualización
	payload := fileReq{
		Message: "🤖 bot: actualiza préstamo de " + debtor + " a pagado",
		Content: newEncodedContent,
		Sha:     exactFile.Sha,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	reqUpdate, err := http.NewRequestWithContext(ctx, http.MethodPut, fileUrl, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	reqUpdate.Header.Set("Authorization", "Bearer "+c.Token)

	respUpdate, err := c.HTTPClient.Do(reqUpdate)
	if err != nil {
		return err
	}
	defer respUpdate.Body.Close()

	if respUpdate.StatusCode != http.StatusOK {
		return fmt.Errorf("falló actualización en github: %d", respUpdate.StatusCode)
	}

	return nil
}

func escapeGitPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}
