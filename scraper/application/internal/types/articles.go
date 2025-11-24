package types

type ArticleDomain int

const (
	FinanceDomain ArticleDomain = iota
)

var domains = map[ArticleDomain]string{
	FinanceDomain: "finance",
}

func (d ArticleDomain) ToString() string {
	return domains[d]
}

type Article struct {
	Url       string        `json:"url"`
	Title     string        `json:"title"`
	Author    string        `json:"author"`
	Published string        `json:"published"`
	Provider  string        `json:"provider"`
	Content   string        `json:"content"`
	Domain    ArticleDomain `json:"domain"`
}
