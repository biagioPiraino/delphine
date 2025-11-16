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

type ArticleMetadata struct {
	Url       string        `json:"url"`
	Author    string        `json:"author"`
	Published string        `json:"published"`
	Domain    ArticleDomain `json:"domain"`
}

type Article struct {
	Url     string        `json:"url"`
	Content string        `json:"content"`
	Domain  ArticleDomain `json:"domain"`
}
