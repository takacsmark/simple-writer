package main

func (e *editor) wordCount() int {
	count := 0
	for _, line := range e.lines {
		inWord := false
		for _, r := range line {
			if isWordRune(r) {
				if !inWord {
					count++
					inWord = true
				}
				continue
			}
			inWord = false
		}
	}
	return count
}
