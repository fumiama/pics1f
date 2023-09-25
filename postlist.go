package pics1f

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

var (
	ErrPostListIndexMismatch = errors.New("post list index mismatch")
	ErrPostListEmptyPage     = errors.New("post list empty page")
)

const (
	pagelinktemplate = "http://pic.s1f.pw/index.php/page/%d/"
)

const (
	postlistrestr     = `<h2 class="post-list-title">[\n\s]*<a\s*href="(http://pic.s1f.pw/index.php/[/%\w-]+)" title=".*">([^\n\t]+)</a>`
	currentindexrestr = `<a class="page-link" href="#">(\d+)</a>`
	lastindexrestr    = `<a class="page-link" href='http://pic.s1f.pw/index.php/page/(\d+)/'>[\n\s]*<i class="fa fa-angle-double-right"></i>`
)

var (
	postlistre     = regexp.MustCompile(postlistrestr)
	currentindexre = regexp.MustCompile(currentindexrestr)
	lastindexre    = regexp.MustCompile(lastindexrestr)
)

// PostList 一页文章索引
type PostList struct {
	Index int    // Index is http://pic.s1f.pw/index.php/page/Index/
	Total int    // Total is the total page number
	Pages []Page // Pages in this index
}

// NewPostList ...
func NewPostList(index int) (pl PostList) {
	pl.Index = index
	return
}

// Fetch ...
func (pl *PostList) Fetch() error {
	u := fmt.Sprintf(pagelinktemplate, pl.Index)
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
	_ = os.WriteFile("t.html", data, 0644)
	matched := currentindexre.FindAllStringSubmatch(dats, -1)
	if len(matched) != 1 {
		return fmt.Errorf("unexpected current index matched len: %d", len(matched))
	}
	index, err := strconv.Atoi(matched[0][1])
	if err != nil {
		return err
	}
	if index != pl.Index {
		return ErrPostListIndexMismatch
	}
	matched = lastindexre.FindAllStringSubmatch(dats, -1)
	if len(matched) != 1 {
		return fmt.Errorf("unexpected last index matched len: %d", len(matched))
	}
	pl.Total, err = strconv.Atoi(matched[0][1])
	if err != nil {
		return err
	}
	matched = postlistre.FindAllStringSubmatch(dats, -1)
	if len(matched) == 0 {
		return ErrPostListEmptyPage
	}
	if pl.Pages == nil {
		pl.Pages = *pagepool.SelectFromPool()
	}
	if len(pl.Pages) > len(matched) {
		pl.Pages = pl.Pages[:len(matched)]
	} else if len(pl.Pages) < len(matched) {
		pl.Pages = append(pl.Pages, make([]Page, len(matched)-len(pl.Pages))...)
	}
	for i, pairs := range matched {
		pl.Pages[i].CanonicalURL = pairs[1]
	}
	return nil
}

// Next ...
func (pl *PostList) Next() error {
	pl.Index++
	if pl.Pages != nil {
		parr := (pagearray)(pl.Pages)
		pagepool.PutIntoPool(&parr)
		pl.Pages = nil
	}
	if pl.Total > 0 && pl.Index > pl.Total {
		return io.EOF
	}
	return pl.Fetch()
}
