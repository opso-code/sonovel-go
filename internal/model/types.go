package model

type Rule struct {
	ID       int      `json:"id"`
	URL      string   `json:"url"`
	Name     string   `json:"name"`
	Comment  string   `json:"comment"`
	Language string   `json:"language"`
	Disabled bool     `json:"disabled"`
	Search   *Search  `json:"search"`
	Book     *Book    `json:"book"`
	Toc      *Toc     `json:"toc"`
	Chapter  *Chapter `json:"chapter"`
	Crawl    *Crawl   `json:"crawl"`
}

type Search struct {
	Disabled       bool   `json:"disabled"`
	BaseURI        string `json:"baseUri"`
	Timeout        int    `json:"timeout"`
	URL            string `json:"url"`
	Method         string `json:"method"`
	Data           string `json:"data"`
	Cookies        string `json:"cookies"`
	Result         string `json:"result"`
	BookName       string `json:"bookName"`
	Author         string `json:"author"`
	Category       string `json:"category"`
	LatestChapter  string `json:"latestChapter"`
	LastUpdateTime string `json:"lastUpdateTime"`
	Status         string `json:"status"`
	WordCount      string `json:"wordCount"`
	Pagination     bool   `json:"pagination"`
	NextPage       string `json:"nextPage"`
}

type Book struct {
	BaseURI        string `json:"baseUri"`
	Timeout        int    `json:"timeout"`
	URL            string `json:"url"`
	BookName       string `json:"bookName"`
	Author         string `json:"author"`
	Intro          string `json:"intro"`
	Category       string `json:"category"`
	CoverURL       string `json:"coverUrl"`
	LatestChapter  string `json:"latestChapter"`
	LastUpdateTime string `json:"lastUpdateTime"`
	Status         string `json:"status"`
}

type Toc struct {
	BaseURI    string `json:"baseUri"`
	Timeout    int    `json:"timeout"`
	URL        string `json:"url"`
	List       string `json:"list"`
	Item       string `json:"item"`
	Desc       bool   `json:"isDesc"`
	Pagination bool   `json:"pagination"`
	NextPage   string `json:"nextPage"`
}

type Chapter struct {
	BaseURI            string `json:"baseUri"`
	Timeout            int    `json:"timeout"`
	Title              string `json:"title"`
	Content            string `json:"content"`
	ParagraphTagClosed bool   `json:"paragraphTagClosed"`
	ParagraphTag       string `json:"paragraphTag"`
	FilterTxt          string `json:"filterTxt"`
	FilterTag          string `json:"filterTag"`
	Pagination         bool   `json:"pagination"`
	NextPage           string `json:"nextPage"`
	NextPageInJS       string `json:"nextPageInJs"`
	NextChapterLink    string `json:"nextChapterLink"`
}

type Crawl struct {
	Concurrency      int `json:"concurrency"`
	MinInterval      int `json:"minInterval"`
	MaxInterval      int `json:"maxInterval"`
	MaxAttempts      int `json:"maxAttempts"`
	RetryMinInterval int `json:"retryMinInterval"`
	RetryMaxInterval int `json:"retryMaxInterval"`
}

type SearchResult struct {
	SourceID       int    `json:"sourceId"`
	SourceName     string `json:"sourceName"`
	URL            string `json:"url"`
	BookName       string `json:"bookName"`
	Author         string `json:"author"`
	Category       string `json:"category"`
	LatestChapter  string `json:"latestChapter"`
	LastUpdateTime string `json:"lastUpdateTime"`
	Status         string `json:"status"`
	WordCount      string `json:"wordCount"`
}

type BookMeta struct {
	URL            string
	BookName       string
	Author         string
	Intro          string
	Category       string
	CoverURL       string
	LatestChapter  string
	LastUpdateTime string
	Status         string
}

type ChapterItem struct {
	Order   int
	Title   string
	URL     string
	Content string
}

type Config struct {
	RulesFile     string
	SourceID      int
	Concurrency   int
	MinIntervalMS int
	MaxIntervalMS int
	EnableRetry   bool
	MaxRetries    int
	RetryMinMS    int
	RetryMaxMS    int
	OutputDir     string
	Format        string
	PreserveCache bool
	SearchLimit   int
	ChapterStart  int
	ChapterEnd    int
	OnProgress    func(done, total int)
}
