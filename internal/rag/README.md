# Módulo RAG (Retrieval-Augmented Generation)

Este documento explica el flujo completo del módulo RAG del ERP: desde que se registra la
ruta en `cmd/api/router.go`, pasa por la carga de un documento Word/PDF, se indexa como
vectores en PostgreSQL (pgvector), y termina cuando un usuario hace una pregunta en el chat.

```
┌────────────────────────────────────────────────────────────────────────────┐
│                        INGESTA (indexación de documentos)                  │
│                                                                            │
│  POST /documents         POST /documents/versions                          │
│      │                        │                                            │
│      ▼                        ▼                                            │
│  documents.Handler ──► documents.Service ──► (goroutine) rag.Service        │
│                                              IndexDocument()               │
│                                                  │                         │
│                                                  ▼                         │
│  1. worker.Parser: descarga el archivo (URL) y extrae texto (.docx/.pdf/..)│
│  2. worker.Chunker: divide el texto en fragmentos (chunks) de ~1000 chars  │
│  3. worker.Embedder: genera un vector de 1536 dims por chunk (OpenAI)      │
│  4. rag.Repository.SaveDocumentEmbedding(): INSERT en document_embeddings  │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────────────┐
│                     CONSULTA (pregunta al chat)                            │
│                                                                            │
│  POST /rag/chat                                                            │
│      │                                                                     │
│      ▼                                                                     │
│  rag.Handler ──► rag.Service.Ask()                                         │
│      │                                                                     │
│      ▼                                                                     │
│  1. Embedder: vector de la pregunta (1536 dims)                            │
│  2. Repository.SearchSimilarChunks(): búsqueda vectorial (pgvector)        │
│  3. Se construye el contexto + system prompt                               │
│  4. OpenAI /v1/responses (gpt-5.6-luna) → respuesta                        │
│  5. Se responde con { answer, sources }                                    │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## 1. Registro de rutas y armado de dependencias (`cmd/api/router.go`)

En `SetupRoutes(db *sql.DB)` se crea la cadena de dependencias del RAG:

```go
openAIKey := os.Getenv("OPENAI_API_KEY")

ragEmbedder := worker.NewOpenAIEmbedder(openAIKey)   // worker: genera embeddings
ragRepository := rag.NewRepository(db)               // repo: habla con PostgreSQL
ragService := rag.NewService(ragRepository, ragEmbedder, openAIKey)
ragHandler := rag.NewHandler(ragService)

documentsService := documents.NewService(documentsRepository, ragService) // inyecta el indexador
```

- **`OPENAI_API_KEY`** se lee del `.env`. Sirve para dos cosas: generar embeddings
  (modelo `text-embedding-3-small`, 1536 dimensiones) y el chat generativo vía la
  **Responses API**.
- **`OPENAI_CHAT_MODEL`** (opcional) define el modelo de chat. Default: `gpt-5.6-luna`.
  Otras opciones: `gpt-5.6-terra` (equilibrado) y `gpt-5.6-sol` (flagship, más caro).
- `ragService` se inyecta en **dos lugares**:
  - En `rag.NewHandler(...)` → expone el endpoint de chat.
  - En `documents.NewService(...)` → indexa automáticamente cada documento que se suba.

### Rutas involucradas

| Método | Ruta | Handler | Middlewares |
|--------|------|---------|-------------|
| `POST` | `/documents` | `documentsHandler.CreateDocument` | `protected(allRoles)` → token JWT + suscripción + rol |
| `POST` | `/documents/versions` | `documentsHandler.UpdateVersion` | `protected(allRoles)` |
| `POST` | `/rag/chat` | `ragHandler.HandleChat` | `protected(siteRoles)` → Gerente, Ingeniero, Supervisor |

El chain de middlewares (`auth → RequireActiveSubscription → RequireRole`) coloca el
`company_id` y `user_id` en el contexto de la request, que luego usan handlers y services
para aislar los datos por empresa (**multi-tenant**).

---

## 2. Ingesta: de subir un archivo a insertar los vectores

### 2.1 Recepción del archivo (`internal/documents/handler.go`)

`CreateDocument` recibe un JSON (no multipart): el frontend ya subió el archivo a
Supabase Storage y solo envía la URL pública:

```json
{
  "project_id": "...",
  "document_type_id": "...",
  "title": "Contrato obra A",
  "file_url": "https://xxxx.supabase.co/storage/v1/object/public/docs/contrato.docx",
  "file_extension": "docx",
  "file_size": 24576,
  "change_log": "v1"
}
```

El handler:
1. Lee `companyID` y `userID` del contexto (puestos por los middlewares).
2. Construye `Document` (contenedor lógico) y `DocumentVersion` (archivo físico).
3. Llama a `service.UploadInitialDocument(...)`.

### 2.2 Delegación en background (`internal/documents/service.go`)

`UploadInitialDocument` persiste el documento y su versión en la BD, y luego dispara la
indexación RAG **en una goroutine** para no bloquear la respuesta HTTP:

```go
func (s *Service) indexInBackground(companyID, docID, versionID, fileURL, fileExtension string) {
    go func() {
        idxCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
        defer cancel()

        // Reemplaza los embeddings de versiones anteriores del mismo documento
        s.indexer.DeleteDocumentEmbeddings(idxCtx, companyID, docID)
        s.indexer.IndexDocument(idxCtx, companyID, docID, versionID, fileURL, fileExtension)
    }()
}
```

> Nota: al subir una **nueva versión** (`UploadNewVersion`) se ejecuta el mismo proceso,
> pero primero se borran los embeddings viejos del documento para no mezclar versiones.

### 2.3 Extracción del texto (`internal/rag/worker/parser.go`)

`rag.Service.IndexDocument` llama al **parser**, que descarga el archivo desde la URL
(`http.Get(fileURL)`) y extrae el texto según la extensión:

| Extensión | Método |
|-----------|--------|
| `txt`, `md`, `csv`, `json` | Se usa el contenido tal cual (`[]byte` → `string`) |
| `docx` | `extractDocxText`: abre el ZIP con `github.com/nguyenthenguyen/docx` y extrae solo el texto de las etiquetas `<w:t>` con `encoding/xml` |
| `pdf` | `extractPDFText`: parseo naive de operadores `Tj`/`TJ` (soporte básico, ver TODO) |
| otros | Error: `"formato de archivo no soportado para RAG"` |

### 2.4 Troceado o "chunking" (`internal/rag/worker/chunker.go`)

El texto extraído se divide en fragmentos solapados:

```go
chunks := s.chunker.SplitText(text, 1000, 200)
```

- Tamaño de chunk: **1000 caracteres**.
- Solapamiento (overlap): **200 caracteres** (para que el corte no pierda contexto entre
  fragmentos consecutivos).
- Convierte a `[]rune` para respetar caracteres unicode (acentos, ñ).

### 2.5 Generación del embedding (`internal/rag/worker/embedder.go`)

Por **cada chunk** se llama a la API de embeddings de OpenAI:

```text
POST https://api.openai.com/v1/embeddings
Authorization: Bearer <OPENAI_API_KEY>
{ "input": "<texto del chunk>", "model": "text-embedding-3-small" }
```

Devuelve un vector de **1536 floats** (`[]float32`). Ese número de dimensiones es el que
exige la columna `embedding vector(1536)` de la tabla.

### 2.6 Inserción en la BD (`internal/rag/repository.go`)

Cada chunk con su vector se persiste:

```sql
INSERT INTO document_embeddings
    (company_id, document_id, document_version_id, chunk_index, chunk_content, embedding)
VALUES ($1, $2, $3, $4, $5, $6);
```

Con `$6 = pgvector.NewVector(vector)` (la librería `pgvector-go` convierte `[]float32` al
tipo nativo `vector` de PostgreSQL).

**Estructura de la tabla** (`migrations/026_create_rag.sql`):

| Columna | Tipo | Descripción |
|---------|------|-------------|
| `id` | `UUID PK` | `gen_random_uuid()` |
| `company_id` | `UUID FK → companies` | Multi-tenant |
| `document_id` | `UUID FK → documents` | Documento padre |
| `document_version_id` | `UUID FK → document_versions` | Versión indexada |
| `chunk_index` | `INT` | Posición del fragmento (0, 1, 2…) |
| `chunk_content` | `TEXT` | El texto del fragmento |
| `embedding` | `vector(1536)` | El vector semántico |
| `created_at` | `TIMESTAMPTZ` | Fecha de indexación |

Índices creados en la migración:
- `idx_document_embeddings_vector` → índice **HNSW** con `vector_cosine_ops`
  (búsqueda aproximada de similitud coseno muy rápida).
- `idx_document_embeddings_company` → índice BTREE por empresa para filtrar el multi-tenant.

> El flujo de **inserción de vectores termina aquí**. Al final de la indexación, la tabla
> `document_embeddings` contiene un registro por cada fragmento del documento, listo para
> ser consultado.

---

## 3. Consulta: el chat con inteligencia artificial

### 3.1 Endpoint y handler (`internal/rag/handler.go`)

`POST /rag/chat` recibe:

```json
{
  "question": "¿Cuál es el presupuesto del proyecto?",
  "project_id": "opcional-para-filtrar"
}
```

`HandleChat` valida el JSON, toma el `companyID` del contexto y delega en
`service.Ask(...)`.

### 3.2 El flujo de `Ask` (`internal/rag/service.go`)

1. **Embedding de la pregunta**: `s.embedder.GenerateEmbedding(ctx, req.Question)`
   → vector de 1536 dims con el **mismo** modelo usado en la ingesta (imprescindible:
   embeddings de distinto modelo no son comparables).

2. **Búsqueda vectorial**: `s.repo.SearchSimilarChunks(ctx, companyID, req.ProjectID, vector, 4)`
   - Query con pgvector: `(1 - (e.embedding <=> $1)) AS similarity`
     (`<=>` es la **distancia coseno**; 1 − distancia = similitud).
   - Filtra por `e.company_id = $2` (multi-tenant).
   - Si viene `project_id`, agrega `AND d.project_id = $n` (filtro por proyecto).
   - Ordena por cercanía y trae los **top 4** fragmentos.

3. **Si no hay resultados**: responde un mensaje cortés indicando que no hay información.

4. **Construcción del contexto**: arma un bloque de texto con los fragmentos y guarda las
   fuentes en un mapa (`document_id → título`).

5. **System prompt especializado**: un prompt en español que le dice al modelo que
   responda **solo** basándose en los documentos del contexto y que no invente datos.

6. **Llamada a la Responses API de OpenAI** (GPT-5.x no acepta `chat/completions`):

   ```text
   POST https://api.openai.com/v1/responses
   Authorization: Bearer <OPENAI_API_KEY>
   {
     "model": "gpt-5.6-luna",
     "input": [
       { "role": "system", "content": "<system prompt + contexto>" },
       { "role": "user", "content": "<pregunta>" }
     ],
     "reasoning_effort": "low"
   }
   ```

   El razonamiento bajo (`low`) es apropiado para respuestas factuales: más rápido y barato.
   El texto se extrae de los items tipo `message/output_text` de la sección `output`.

7. **Respuesta al frontend**:

   ```json
   {
     "answer": "El presupuesto total es de $250,000 según el documento...",
     "sources": [
       { "document_id": "...", "title": "Contrato obra A" }
     ],
     "created_at": "2026-08-14T..."
   }
   ```

---

## 4. Diagrama de archivos

```
cmd/api/router.go                 ← rutas + ensamblado de dependencias
internal/documents/
  handler.go                       ← recibe POST /documents y /documents/versions
  service.go                       ← persiste documento + dispara indexación en background
internal/rag/
  handler.go                       ← recibe POST /rag/chat
  service.go                       ← orquestador: Ask() e IndexDocument()
  repository.go                    ← SQL contra PostgreSQL (pgvector)
  model.go                         ← DTOs (ChatRequest, ChatResponse, DocumentSource)
  worker/
    parser.go                      ← descarga archivo y extrae texto (.docx/.pdf/.txt…)
    chunker.go                     ← divide texto en fragmentos solapados
    embedder.go                    ← genera vectores de 1536 dims (OpenAI)
migrations/
  018_create_documents.sql         ← tables documents / document_versions
  026_create_rag.sql               ← table document_embeddings + índices pgvector
```

---

## 5. Puntos clave a recordar

1. **Mismo modelo de embeddings en ingesta y consulta** (`text-embedding-3-small`,
   1536 dims): si cambias el modelo, cambia también `vector(1536)` y reindexa todo.
2. **Multi-tenant**: todos los queries filtran por `company_id` del JWT.
3. **Indexación async**: la subida del documento responde rápido; los embeddings se generan
   en una goroutine con timeout de 5 minutos.
4. **Reindexado por versión**: subir una nueva versión borra los embeddings viejos del
   documento antes de indexar la nueva.
5. **`vector(1536)` es un tipo de PostgreSQL** aportado por la extensión `pgvector`
   (activada en `026_create_rag.sql` con `CREATE EXTENSION IF NOT EXISTS vector;`).
