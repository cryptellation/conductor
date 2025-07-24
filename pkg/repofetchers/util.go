package repofetchers

func parseOwnerAndRepo(url string) (owner, repo string) {
	const prefix = "github.com/"
	idx := findPrefixIndex(url, prefix)
	if idx == -1 {
		return "", ""
	}
	rest := url[idx:]
	if len(rest) == 0 {
		return "", ""
	}
	rest = trimGitSuffix(rest)
	parts := splitOwnerRepo(rest)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func findPrefixIndex(url, prefix string) int {
	for i := 0; i < len(url)-len(prefix); i++ {
		if url[i:i+len(prefix)] == prefix {
			return i + len(prefix)
		}
	}
	return -1
}

func trimGitSuffix(s string) string {
	if len(s) > 4 && s[len(s)-4:] == ".git" {
		return s[:len(s)-4]
	}
	return s
}

func splitOwnerRepo(rest string) []string {
	parts := make([]string, 0, 2)
	for _, p := range rest {
		if p == '/' {
			parts = append(parts, "")
			continue
		}
		if len(parts) == 0 {
			parts = append(parts, "")
		}
		parts[len(parts)-1] += string(p)
	}
	return parts
}
