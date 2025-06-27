package entities

type Chunk struct {
	Data []byte
	Err  error
}

type Map struct {
	Data map[string]int
	Err  error
}

type WordCount struct {
	Word  string
	Count int
}
