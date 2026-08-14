package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Embedder interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

type openAIEmbedder struct {
	apiKey string
	model  string
}

func NewOpenAIEmbedder(apiKey string) Embedder {
	return &openAIEmbedder{
		apiKey: apiKey,
		model:  "text-embedding-3-small", // Retorna vectores de 1536 dimensiones
	}
}

type openAIResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (e *openAIEmbedder) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"input": text,
		"model": e.model,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al conectar con API de Embeddings: %w", err)
	}
	defer resp.Body.Close()

	var result openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no se generó vector para el texto provisto")
	}

	return result.Data[0].Embedding, nil
}
