package worker

type Chunker interface {
	SplitText(text string, chunkSize int, overlap int) []string
}

type chunker struct{}

func NewChunker() Chunker {
	return &chunker{}
}

// SplitText divide un texto extenso en fragmentos con superposición de caracteres
func (c *chunker) SplitText(text string, chunkSize int, overlap int) []string {
	var chunks []string
	runes := []rune(text)
	totalLength := len(runes)

	if totalLength == 0 {
		return chunks
	}

	start := 0
	for start < totalLength {
		end := start + chunkSize
		if end > totalLength {
			end = totalLength
		}

		chunk := string(runes[start:end])
		chunks = append(chunks, chunk)

		// Si llegamos al final del texto, terminamos
		if end == totalLength {
			break
		}

		// Avanzamos manteniendo el solapamiento (overlap)
		start += (chunkSize - overlap)
	}

	return chunks
}
