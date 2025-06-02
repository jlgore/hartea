package har

import "time"

type HAR struct {
	Log Log `json:"log"`
}

type Log struct {
	Version string  `json:"version"`
	Creator Creator `json:"creator"`
	Browser Browser `json:"browser,omitempty"`
	Pages   []Page  `json:"pages,omitempty"`
	Entries []Entry `json:"entries"`
	Comment string  `json:"comment,omitempty"`
}

type Creator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Comment string `json:"comment,omitempty"`
}

type Browser struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Comment string `json:"comment,omitempty"`
}

type Page struct {
	StartedDateTime time.Time    `json:"startedDateTime"`
	ID              string       `json:"id"`
	Title           string       `json:"title"`
	PageTimings     PageTimings  `json:"pageTimings"`
	Comment         string       `json:"comment,omitempty"`
}

type PageTimings struct {
	OnContentLoad int `json:"onContentLoad,omitempty"`
	OnLoad        int `json:"onLoad,omitempty"`
	Comment       string `json:"comment,omitempty"`
}

type Entry struct {
	PageRef         string      `json:"pageref,omitempty"`
	StartedDateTime time.Time   `json:"startedDateTime"`
	Time            float64     `json:"time"`
	Request         Request     `json:"request"`
	Response        Response    `json:"response"`
	Cache           Cache       `json:"cache"`
	Timings         Timings     `json:"timings"`
	ServerIPAddress string      `json:"serverIPAddress,omitempty"`
	Connection      string      `json:"connection,omitempty"`
	Comment         string      `json:"comment,omitempty"`
}

type Request struct {
	Method      string      `json:"method"`
	URL         string      `json:"url"`
	HTTPVersion string      `json:"httpVersion"`
	Cookies     []Cookie    `json:"cookies"`
	Headers     []Header    `json:"headers"`
	QueryString []QueryItem `json:"queryString"`
	PostData    *PostData   `json:"postData,omitempty"`
	HeadersSize int         `json:"headersSize"`
	BodySize    int         `json:"bodySize"`
	Comment     string      `json:"comment,omitempty"`
}

type Response struct {
	Status      int      `json:"status"`
	StatusText  string   `json:"statusText"`
	HTTPVersion string   `json:"httpVersion"`
	Cookies     []Cookie `json:"cookies"`
	Headers     []Header `json:"headers"`
	Content     Content  `json:"content"`
	RedirectURL string   `json:"redirectURL"`
	HeadersSize int      `json:"headersSize"`
	BodySize    int      `json:"bodySize"`
	Comment     string   `json:"comment,omitempty"`
}

type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Path     string    `json:"path,omitempty"`
	Domain   string    `json:"domain,omitempty"`
	Expires  time.Time `json:"expires,omitempty"`
	HTTPOnly bool      `json:"httpOnly,omitempty"`
	Secure   bool      `json:"secure,omitempty"`
	Comment  string    `json:"comment,omitempty"`
}

type Header struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}

type QueryItem struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}

type PostData struct {
	MimeType string  `json:"mimeType"`
	Params   []Param `json:"params"`
	Text     string  `json:"text"`
	Comment  string  `json:"comment,omitempty"`
}

type Param struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	FileName    string `json:"fileName,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

type Content struct {
	Size        int    `json:"size"`
	Compression int    `json:"compression,omitempty"`
	MimeType    string `json:"mimeType"`
	Text        string `json:"text,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

type Cache struct {
	BeforeRequest *CacheState `json:"beforeRequest,omitempty"`
	AfterRequest  *CacheState `json:"afterRequest,omitempty"`
	Comment       string      `json:"comment,omitempty"`
}

type CacheState struct {
	Expires    time.Time `json:"expires,omitempty"`
	LastAccess time.Time `json:"lastAccess"`
	ETag       string    `json:"eTag"`
	HitCount   int       `json:"hitCount"`
	Comment    string    `json:"comment,omitempty"`
}

type Timings struct {
	Blocked int     `json:"blocked,omitempty"`
	DNS     int     `json:"dns,omitempty"`
	Connect int     `json:"connect,omitempty"`
	Send    int     `json:"send"`
	Wait    int     `json:"wait"`
	Receive int     `json:"receive"`
	SSL     int     `json:"ssl,omitempty"`
	Comment string  `json:"comment,omitempty"`
}