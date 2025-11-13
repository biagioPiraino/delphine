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
	Url       string
	Author    string
	Published string
	Domain    ArticleDomain
}

type Article struct {
	Url     string
	Content string
	Domain  ArticleDomain
}
