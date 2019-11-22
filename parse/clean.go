package parse

// Whitelist filters for whitelisted ASCII characters: Space to ~, TAB, LF, CR.
// Returns the number of bytes that passed whitelist.
func Whitelist(b []byte, n int) (nclean int) {
	for _, char := range b[:n] {
		// Accept normal printable ASCII (Space to ~), TAB, LF, CR
		if (char >= 32 && char <= 126) || char == 9 || char == 10 || char == 13 {
			b[nclean] = char
			nclean++
		}
	}
	return nclean
}
