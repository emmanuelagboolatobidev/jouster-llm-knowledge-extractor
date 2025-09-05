package analyzer

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type KeywordExtractor struct {
	stopWords map[string]bool
}

func NewKeywordExtractor() *KeywordExtractor {
	stopWords := map[string]bool{
		"the": true, "be": true, "to": true, "of": true, "and": true,
		"a": true, "in": true, "that": true, "have": true, "i": true,
		"it": true, "for": true, "not": true, "on": true, "with": true,
		"he": true, "as": true, "you": true, "do": true, "at": true,
		"this": true, "but": true, "his": true, "by": true, "from": true,
		"they": true, "we": true, "say": true, "her": true, "she": true,
		"or": true, "an": true, "will": true, "my": true, "one": true,
		"all": true, "would": true, "there": true, "their": true, "what": true,
		"so": true, "up": true, "out": true, "if": true, "about": true,
		"who": true, "get": true, "which": true, "go": true, "me": true,
		"when": true, "make": true, "can": true, "like": true, "time": true,
		"no": true, "just": true, "him": true, "know": true, "take": true,
		"people": true, "into": true, "year": true, "your": true, "good": true,
		"some": true, "could": true, "them": true, "see": true, "other": true,
		"than": true, "then": true, "now": true, "look": true, "only": true,
		"come": true, "its": true, "over": true, "think": true, "also": true,
		"back": true, "after": true, "use": true, "two": true, "how": true,
		"our": true, "work": true, "first": true, "well": true, "way": true,
		"even": true, "new": true, "want": true, "because": true, "any": true,
		"these": true, "give": true, "day": true, "most": true, "us": true,
		"is": true, "was": true, "are": true, "been": true, "has": true,
		"had": true, "were": true, "said": true, "did": true, "having": true,
		"may": true, "being": true,
	}
	return &KeywordExtractor{stopWords: stopWords}
}

func (ke *KeywordExtractor) ExtractKeywords(text string, topN int) []string {
	nouns := ke.extractNouns(text)
	
	wordFreq := make(map[string]int)
	for _, noun := range nouns {
		word := strings.ToLower(noun)
		if !ke.stopWords[word] && len(word) > 2 {
			wordFreq[word]++
		}
	}
	
	type wordCount struct {
		word  string
		count int
	}
	
	var counts []wordCount
	for word, count := range wordFreq {
		counts = append(counts, wordCount{word, count})
	}
	
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].count == counts[j].count {
			return counts[i].word < counts[j].word
		}
		return counts[i].count > counts[j].count
	})
	
	result := make([]string, 0, topN)
	for i := 0; i < len(counts) && i < topN; i++ {
		result = append(result, counts[i].word)
	}
	
	return result
}

func (ke *KeywordExtractor) extractNouns(text string) []string {
	wordPattern := regexp.MustCompile(`\b[A-Za-z]+\b`)
	words := wordPattern.FindAllString(text, -1)
	
	var nouns []string
	for _, word := range words {
		if ke.isLikelyNoun(word) {
			nouns = append(nouns, word)
		}
	}
	
	return nouns
}

func (ke *KeywordExtractor) isLikelyNoun(word string) bool {
	word = strings.ToLower(word)
	
	if len(word) < 3 {
		return false
	}
	
	nounSuffixes := []string{
		"tion", "sion", "ment", "ness", "ity", "er", "or",
		"ism", "ist", "ance", "ence", "ship", "hood", "dom",
		"ing", "age", "ery", "ory", "cy", "ty", "ure",
	}
	
	for _, suffix := range nounSuffixes {
		if strings.HasSuffix(word, suffix) {
			return true
		}
	}
	
	firstChar := rune(word[0])
	if unicode.IsUpper(firstChar) {
		return true
	}
	
	commonNouns := map[string]bool{
		"data": true, "system": true, "user": true, "file": true,
		"code": true, "app": true, "web": true, "api": true,
		"server": true, "client": true, "database": true, "service": true,
		"product": true, "company": true, "team": true, "project": true,
		"email": true, "phone": true, "address": true, "name": true,
		"text": true, "image": true, "video": true, "audio": true,
		"document": true, "report": true, "analysis": true, "result": true,
		"model": true, "algorithm": true, "function": true, "method": true,
		"process": true, "task": true, "job": true, "role": true,
		"customer": true, "market": true, "business": true, "industry": true,
		"technology": true, "platform": true, "solution": true, "tool": true,
		"feature": true, "update": true, "version": true, "release": true,
	}
	
	return commonNouns[word]
}

func CalculateConfidence(text string, summary string, topics []string) float64 {
	if text == "" || summary == "" {
		return 0.0
	}
	
	textLen := len(strings.Fields(text))
	summaryLen := len(strings.Fields(summary))
	
	baseConfidence := 0.5
	
	if textLen > 50 {
		baseConfidence += 0.2
	} else if textLen > 20 {
		baseConfidence += 0.1
	}
	
	if summaryLen > 5 && summaryLen < 50 {
		baseConfidence += 0.1
	}
	
	if len(topics) >= 3 {
		baseConfidence += 0.1
	}
	
	compressionRatio := float64(summaryLen) / float64(textLen)
	if compressionRatio > 0.05 && compressionRatio < 0.3 {
		baseConfidence += 0.1
	}
	
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}
	
	return baseConfidence
}