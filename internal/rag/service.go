package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"erp-constructora/internal/rag/worker"
)

type Service interface {
	Ask(ctx context.Context, companyID string, req ChatRequest) (*ChatResponse, error)
	IndexDocument(ctx context.Context, companyID, docID, versionID, fileURL, fileExtension string) error
	DeleteDocumentEmbeddings(ctx context.Context, companyID, docID string) error
}

type service struct {
	repo       Repository
	embedder   worker.Embedder
	parser     worker.Parser
	chunker    worker.Chunker
	openAIKey  string
	openAIBase string
	chatModel  string
}

func NewService(repo Repository, embedder worker.Embedder, openAIKey, chatModel string) Service {
	if chatModel == "" {
		chatModel = "gpt-5.6-luna"
	}
	return &service{
		repo:       repo,
		embedder:   embedder,
		parser:     worker.NewParser(),
		chunker:    worker.NewChunker(),
		openAIKey:  openAIKey,
		openAIBase: "https://api.openai.com/v1",
		chatModel:  chatModel,
	}
}

// Estructuras internas para la llamada HTTP a la Responses API de OpenAI
type responseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// apiRequest sigue el formato de POST /v1/responses (GPT-5.x no acepta chat/completions).
type apiRequest struct {
	Model           string            `json:"model"`
	Input           []responseMessage `json:"input"`
	ReasoningEffort string            `json:"reasoning_effort"`
}

type responseContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responseItem struct {
	Type    string            `json:"type"`
	Content []responseContent `json:"content"`
}

type apiResponse struct {
	Output []responseItem `json:"output"`
}

func (s *service) Ask(ctx context.Context, companyID string, req ChatRequest) (*ChatResponse, error) {
	// 1. Convertir la pregunta del usuario a vector usando el embedder
	queryVector, err := s.embedder.GenerateEmbedding(ctx, req.Question)
	if err != nil {
		return nil, fmt.Errorf("error al generar el embedding de la consulta: %w", err)
	}

	// 2. Traer los 4 fragmentos con mayor similitud matemática
	chunks, err := s.repo.SearchSimilarChunks(ctx, companyID, req.ProjectID, queryVector, 4)
	if err != nil {
		return nil, fmt.Errorf("error al recuperar fragmentos de contexto: %w", err)
	}

	// 3. Si no hay fragmentos procesados en la empresa
	if len(chunks) == 0 {
		return &ChatResponse{
			Answer:    "No he encontrado documentos o información relevante cargada en el sistema para responder a tu consulta.",
			Sources:   []DocumentSource{},
			CreatedAt: time.Now(),
		}, nil
	}

	// 4. Construir el contexto e identificar las fuentes
	var contextBuilder strings.Builder
	sourcesMap := make(map[string]string) // Para evitar fuentes repetidas (ID -> Título)

	for _, chunk := range chunks {
		contextBuilder.WriteString(fmt.Sprintf("--- Documento: %s ---\n%s\n\n", chunk.DocumentTitle, chunk.ChunkContent))
		sourcesMap[chunk.DocumentID] = chunk.DocumentTitle
	}

	// 5. System Prompt especializado para el rubro de la construcción
	systemPrompt := fmt.Sprintf(`Eres el asistente de Inteligencia Artificial para un software de gestión de empresas constructoras.
Tu objetivo es responder a las preguntas de los usuarios (Ingenieros, Gerentes, Supervisores) basándote ÚNICAMENTE en los fragmentos de documentos internos provistos a continuación.

REGLAS OBLIGATORIAS:
1. Responde de forma precisa, clara y profesional.
2. Si la información solicitada (ejemplo: fechas de caducidad, nombres, montos) está en el contexto, indícala explícitamente.
3. Si la respuesta NO está implícita o explícita en los documentos entregados, responde amablemente indicando que no dispones de esa información en la documentación actual. No inventes ni asumas datos.

CONTEXTO DE DOCUMENTOS EXTRAÍDOS:
%s`, contextBuilder.String())

	// 6. Preparar petición a la Responses API de OpenAI (GPT-5.x solo usa /v1/responses)
	reqBody := apiRequest{
		Model: s.chatModel,
		Input: []responseMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: req.Question},
		},
		ReasoningEffort: "low", // RAG factual: razonamiento bajo = más rápido y barato
	}

	payload, _ := json.Marshal(reqBody)
	chatURL := strings.TrimRight(s.openAIBase, "/") + "/responses"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", chatURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.openAIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error al conectar con OpenAI: %w", err)
	}
	defer httpResp.Body.Close()

	var apiRes apiResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&apiRes); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta de OpenAI: %w", err)
	}

	// 6.1 Recomponer el texto de salida a partir de los items tipo "message" de la respuesta
	answer, err := extractResponseText(apiRes)
	if err != nil {
		return nil, err
	}
	if answer == "" {
		return nil, fmt.Errorf("OpenAI no devolvió ninguna respuesta")
	}

	// 7. Formatear fuentes limpias para el frontend
	var sources []DocumentSource
	for docID, title := range sourcesMap {
		sources = append(sources, DocumentSource{
			DocumentID: docID,
			Title:      title,
		})
	}

	return &ChatResponse{
		Answer:    answer,
		Sources:   sources,
		CreatedAt: time.Now(),
	}, nil
}

// extractResponseText recorre los items de `output` de la Responses API y concatena
// el texto de los bloques de tipo message/output_text.
func extractResponseText(resp apiResponse) (string, error) {
	var b strings.Builder
	for _, item := range resp.Output {
		if item.Type != "message" {
			continue
		}
		for _, c := range item.Content {
			if strings.HasPrefix(c.Type, "output_text") {
				b.WriteString(c.Text)
			}
		}
	}
	return strings.TrimSpace(b.String()), nil
}

func (s *service) IndexDocument(ctx context.Context, companyID, docID, versionID, fileURL, fileExtension string) error {
	text, err := s.parser.ExtractTextFromURL(fileURL, fileExtension)
	if err != nil {
		return fmt.Errorf("error extrayendo texto del documento: %w", err)
	}

	chunks := s.chunker.SplitText(text, 1000, 200)

	for i, chunk := range chunks {
		vector, err := s.embedder.GenerateEmbedding(ctx, chunk)
		if err != nil {
			return fmt.Errorf("error generando embedding del chunk %d: %w", i, err)
		}

		if err := s.repo.SaveDocumentEmbedding(ctx, companyID, docID, versionID, i, chunk, vector); err != nil {
			return fmt.Errorf("error persistiendo embedding del chunk %d: %w", i, err)
		}
	}

	return nil
}

func (s *service) DeleteDocumentEmbeddings(ctx context.Context, companyID, docID string) error {
	return s.repo.DeleteDocumentEmbeddings(ctx, companyID, docID)
}
