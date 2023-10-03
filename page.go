package pics1f

import (
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	ErrNoLoadableLink         = errors.New("no loadable link")
	ErrShortLinkPMismatched   = errors.New("shortlink p mismatched")
	ErrCanonicalURLMismatched = errors.New("canonical url mismatched")
	ErrInvalidShortLinkP      = errors.New("invalid shortlink p")
	ErrTooManyContents        = errors.New("too many contents")
	ErrMaxRetryTimeExceeded   = errors.New("max retry time exceeded")
	ErrInvalidMultiplier      = errors.New("invalid multiplier")
)

const (
	shortlinktemplate = "http://pic.s1f.pw/?p=%d"
)

const (
	shortlinkrestr = `<link rel='shortlink' href='http://pic.s1f.pw/\?p=(\d+)'\s*/>`
	canonicalrestr = `<link rel="canonical" href="([a-zA-z]+://[^\s\n\t]+)"\s*/>`
	titlerestr     = `<h3 class="post-title">\s*([^\n\t]+)\s*</h3>`
	contentrestr   = `<img decoding="async" src="([a-zA-z]+://[^\s\n\t]+)"\s*/>`
)

var (
	shortlinkre = regexp.MustCompile(shortlinkrestr)
	canonicalre = regexp.MustCompile(canonicalrestr)
	titlere     = regexp.MustCompile(titlerestr)
	contentre   = regexp.MustCompile(contentrestr)
)

// Page 一个写真集
type Page struct {
	ShortLinkP   int      // ShortLinkP fills http://pic.s1f.pw/?p=ShortLinkP
	CanonicalURL string   // CanonicalURL is the jumped result from shortlink
	Title        string   // Title is the post title
	ContentURLs  []string // ContentURLs are the post contents
}

// NewPageShortLink 从短链接获得 Page
func NewPageShortLink(p int) (page Page, err error) {
	if p <= 0 {
		err = ErrInvalidShortLinkP
		return
	}
	page.ShortLinkP = p
	err = page.Fetch()
	return
}

// NewPageCanonical 从标准链接获得 Page
func NewPageCanonical(u string) (page Page, err error) {
	page.CanonicalURL = u
	err = page.Fetch()
	return
}

// Fetch 在 ShortLinkP 和 CanonicalURL 不合法时报错, 但不保证 Title 和 ContentURLs 非空
func (p *Page) Fetch() error {
	u := ""
	switch {
	case p.ShortLinkP > 0:
		u = fmt.Sprintf(shortlinktemplate, p.ShortLinkP)
	case p.CanonicalURL != "":
		u = p.CanonicalURL
	default:
		return ErrNoLoadableLink
	}
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	dats := BytesToString(data)
	matched := shortlinkre.FindAllStringSubmatch(dats, -1)
	if len(matched) != 1 {
		return fmt.Errorf("unexpected shortlink matched len: %d", len(matched))
	}
	ps := matched[0][1]
	pn, err := strconv.Atoi(ps)
	if err != nil {
		return err
	}
	if p.ShortLinkP > 0 {
		if pn != p.ShortLinkP {
			return ErrShortLinkPMismatched
		}
	} else {
		p.ShortLinkP = pn
	}
	matched = canonicalre.FindAllStringSubmatch(dats, -1)
	if len(matched) != 1 {
		return fmt.Errorf("unexpected canonical matched len: %d", len(matched))
	}
	u = matched[0][1]
	if p.CanonicalURL != "" {
		if u != p.CanonicalURL {
			return ErrCanonicalURLMismatched
		}
	} else {
		p.CanonicalURL = u
	}
	matched = titlere.FindAllStringSubmatch(dats, -1)
	if len(matched) != 1 {
		return fmt.Errorf("unexpected title matched len: %d", len(matched))
	}
	p.Title = html.UnescapeString(matched[0][1])
	matched = contentre.FindAllStringSubmatch(dats, -1)
	if len(matched) > 0 {
		if p.ContentURLs == nil {
			p.ContentURLs = *stringarraypool.SelectFromPool()
		}
		if len(p.ContentURLs) > len(matched) {
			p.ContentURLs = p.ContentURLs[:len(matched)]
		} else if len(p.ContentURLs) < len(matched) {
			p.ContentURLs = append(p.ContentURLs, make([]string, len(matched)-len(p.ContentURLs))...)
		}
		for i, pairs := range matched {
			p.ContentURLs[i] = pairs[1]
		}
	}
	return nil
}

// DownloadContentsTo 并发下载图片到 dir/title/index.webp
//
// retry 小于 0 表示无穷
func (p *Page) DownloadContentsTo(dir string, retry int, override bool, threadmultiplier int, errcallback func(error)) error {
	if threadmultiplier <= 0 {
		return ErrInvalidMultiplier
	}
	namefmt := path.Join(dir, p.Title)
	err := os.MkdirAll(namefmt, 0755)
	if err != nil {
		return err
	}
	i := len(p.ContentURLs)
	switch {
	case i < 10:
		namefmt = path.Join(namefmt, "%d.webp")
	case i < 100:
		namefmt = path.Join(namefmt, "%02d.webp")
	case i < 1000:
		namefmt = path.Join(namefmt, "%03d.webp")
	case i < 10000:
		namefmt = path.Join(namefmt, "%04d.webp")
	default:
		return ErrTooManyContents
	}
	n := runtime.NumCPU() * threadmultiplier
	batch := len(p.ContentURLs) / n
	if batch == 0 {
		batch = 1
		n = len(p.ContentURLs)
	}
	dlonepage := func(i int, u string) error {
		n := 0
		filepath := fmt.Sprintf(namefmt, i)
		if !override {
			if _, err := os.Stat(filepath); err == nil {
				return nil
			}
		}
		var resp *http.Response
		var err error
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Accept", `image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8`)
		req.Header.Add("Referer", `http://pic.s1f.pw/`)
		req.Header.Add("User-Agent", UserAgent)
		for retry < 0 || n <= retry {
			resp, err = HTTPClient.Do(req)
			if err != nil {
				errcallback(err)
				time.Sleep(time.Millisecond * 100)
				n++
				continue
			}
			break
		}
		if retry >= 0 && n > retry {
			return ErrMaxRetryTimeExceeded
		}
		defer resp.Body.Close()
		f, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, resp.Body)
		return err
	}
	dlbatch := func(wg *sync.WaitGroup, offset int, urls []string) {
		defer wg.Done()
		for i, u := range urls {
			err := dlonepage(offset+i, u)
			if err != nil {
				errcallback(err)
			}
		}
	}
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go dlbatch(&wg, i*batch, p.ContentURLs[i*batch:(i+1)*batch])
		time.Sleep(time.Millisecond * 100)
	}
	if batch > 1 && len(p.ContentURLs) > n*batch {
		wg.Add(1)
		go dlbatch(&wg, n*batch, p.ContentURLs[n*batch:])
	}
	wg.Wait()
	return nil
}

// Reset ...
func (p *Page) Reset() {
	p.ShortLinkP = 0
	p.CanonicalURL = ""
	stringarraypool.PutIntoPool(&p.ContentURLs)
	p.ContentURLs = nil
}
