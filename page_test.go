package pics1f

import (
	"os"
	"testing"
)

var pagetesthtml = func() []byte {
	data, err := os.ReadFile("5363.html")
	if err != nil {
		panic(err)
	}
	return data
}()

func TestPageShortLinkRegex(t *testing.T) {
	matched := shortlinkre.FindAllStringSubmatch(BytesToString(pagetesthtml), -1)
	if len(matched) != 1 {
		t.Fatal("unexpected matched len:", len(matched))
	}
	ps := matched[0][1]
	if ps != "5363" {
		t.Fatal("unexpected matched p:", ps)
	}
}

func TestPageCanonicalRegex(t *testing.T) {
	matched := canonicalre.FindAllStringSubmatch(BytesToString(pagetesthtml), -1)
	if len(matched) != 1 {
		t.Fatal("unexpected matched len:", len(matched))
	}
	u := matched[0][1]
	if u != "http://pic.s1f.pw/index.php/2023/09/18/xiuren%e7%a7%80%e4%ba%ba%e7%bd%91-2023-08-18-no-7254-%e9%b1%bc%e5%ad%90%e9%85%b1fish/" {
		t.Fatal("unexpected matched canonical url:", u)
	}
}

func TestPageTitleRegex(t *testing.T) {
	matched := titlere.FindAllStringSubmatch(BytesToString(pagetesthtml), -1)
	if len(matched) != 1 {
		t.Fatal("unexpected matched len:", len(matched))
	}
	title := matched[0][1]
	if title != "[XiuRen秀人网] 2023.08.18 No.7254 鱼子酱Fish" {
		t.Fatal("unexpected matched title:", title)
	}
}

func TestPageContentsRegex(t *testing.T) {
	matched := contentre.FindAllStringSubmatch(BytesToString(pagetesthtml), -1)
	if len(matched) != 85 {
		t.Fatal("unexpected matched len:", len(matched))
	}
	u := matched[0][1]
	if u != "https://wp.007irs.com/f/Aw2N7Fy/9e4bb112.webp" {
		t.Fatal("unexpected matched content url:", u)
	}
}

func TestPageFetch(t *testing.T) {
	page := Page{ShortLinkP: 5363}
	err := page.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	if page.ShortLinkP != 5363 {
		t.Fatal("unexpected page p:", page.ShortLinkP)
	}
	if page.CanonicalURL != "http://pic.s1f.pw/index.php/2023/09/18/xiuren%e7%a7%80%e4%ba%ba%e7%bd%91-2023-08-18-no-7254-%e9%b1%bc%e5%ad%90%e9%85%b1fish/" {
		t.Fatal("unexpected page canonical url:", page.CanonicalURL)
	}
	if page.Title != "[XiuRen秀人网] 2023.08.18 No.7254 鱼子酱Fish" {
		t.Fatal("unexpected page title:", page.Title)
	}
	if len(page.ContentURLs) != 85 {
		t.Fatal("unexpected page contents len:", len(page.ContentURLs))
	}
	if page.ContentURLs[0] != "https://wp.007irs.com/f/Aw2N7Fy/9e4bb112.webp" {
		t.Fatal("unexpected first page content url:", page.ContentURLs[0])
	}
}

func TestPageDownloadContentsTo(t *testing.T) {
	page := Page{ShortLinkP: 5183}
	err := page.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	err = page.DownloadContentsTo("tmp", 3, true, 4, func(err error) {
		t.Fatal(err)
	})
	if err != nil {
		t.Fatal(err)
	}
}
