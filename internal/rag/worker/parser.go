package worker

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nguyenthenguyen/docx"
)

type Parser interface {
	ExtractTextFromURL(fileURL string, extension string) (string, error)
}

type parser struct{}

func NewParser() Parser {
	return &parser{}
}

func (p *parser) ExtractTextFromURL(fileURL string, extension string) (string, error) {
	// 1. Descargar el archivo desde la URL (Supabase Storage / S3)
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("error al descargar el archivo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error al descargar el archivo: HTTP %d", resp.StatusCode)
	}

	fileBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error al leer el contenido del archivo: %w", err)
	}

	ext := strings.ToLower(strings.TrimPrefix(extension, "."))

	switch ext {
	case "txt", "md", "csv", "json":
		return string(fileBytes), nil

	case "docx":
		return p.extractDocxText(fileBytes)

	case "pdf":
		// TODO: Integrar un extractor de PDF real (pdfcpu, unipdf o similar).
		// Por ahora intentamos extraer el texto simple contenido en el binario.
		return p.extractPDFText(fileBytes)

	default:
		return "", fmt.Errorf("formato de archivo no soportado para RAG: %s", extension)
	}
}

func (p *parser) extractDocxText(data []byte) (string, error) {
	doc, err := docx.ReadDocxFromMemory(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("no se pudo leer el documento Word: %w", err)
	}
	defer doc.Close()

	text, err := extractWordText(doc.Editable().GetContent())
	if err != nil {
		return "", err
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("el documento Word no contiene texto visible")
	}
	return text, nil
}

// extractWordText recorre el XML de document.xml y devuelve el texto
// contenido en las etiquetas <w:t> (los párrafos/tablas se separan con salto de línea).
func extractWordText(docXML string) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(docXML))

	var b strings.Builder
	inText := false
	paragraphDone := false // evita saltos de línea duplicados por el cierre de <w:p>

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("no se pudo interpretar el XML del documento Word: %w", err)
		}

		switch el := tok.(type) {
		case xml.StartElement:
			switch el.Name.Local {
			case "t":
				inText = true
			case "p", "br":
				if !paragraphDone {
					b.WriteString("\n")
					paragraphDone = true
				}
			}
		case xml.EndElement:
			switch el.Name.Local {
			case "t":
				inText = false
			case "p":
				paragraphDone = false
			}
		case xml.CharData:
			if inText {
				b.WriteString(string(el))
			}
		}
	}

	return b.String(), nil
}

func (p *parser) extractPDFText(data []byte) (string, error) {
	// Búsqueda naive de operadores "Tj" y "TJ" (strings de texto en PDF).
	// No es robusto para PDFs comprimidos/imágenes; solo da soporte básico.
	matches := strings.Split(string(data), "Tj")
	var b strings.Builder
	for _, m := range matches {
		line := strings.TrimSpace(m)
		if line == "" {
			continue
		}
		// Normalizamos los paréntesis de la sintaxis PDF: borra saltos.
		b.WriteString(line)
		b.WriteString("\n")
	}
	text := strings.TrimSpace(b.String())
	if text == "" {
		return "", fmt.Errorf("no se pudo extraer texto del PDF (puede estar comprimido o ser una imagen)")
	}
	return text, nil
}
