package pics1f

import (
	"os"
	"testing"
)

var postlisttesthtml = func() []byte {
	data, err := os.ReadFile("1.html")
	if err != nil {
		panic(err)
	}
	return data
}()

func TestPostListPostListRegex(t *testing.T) {
	matched := postlistre.FindAllStringSubmatch(BytesToString(postlisttesthtml), -1)
	if len(matched) != 10 {
		t.Fatal("unexpected matched len:", len(matched))
	}
	u := matched[0][1]
	if u != "http://pic.s1f.pw/index.php/2023/09/25/xiuren%e7%a7%80%e4%ba%ba%e7%bd%91-2023-01-19-no-6159-arude%e8%96%87%e8%96%87/" {
		t.Fatal("unexpected matched content url:", u)
	}
}

func TestPostListCurrentIndexRegex(t *testing.T) {
	matched := currentindexre.FindAllStringSubmatch(BytesToString(postlisttesthtml), -1)
	if len(matched) != 1 {
		t.Fatal("unexpected matched len:", len(matched))
	}
	index := matched[0][1]
	if index != "1" {
		t.Fatal("unexpected matched index:", index)
	}
}

func TestPostListLastIndexRegex(t *testing.T) {
	matched := lastindexre.FindAllStringSubmatch(BytesToString(postlisttesthtml), -1)
	if len(matched) != 1 {
		t.Fatal("unexpected matched len:", len(matched))
	}
	index := matched[0][1]
	if index != "284" {
		t.Fatal("unexpected matched index:", index)
	}
}

func TestPostListFetch(t *testing.T) {
	pl := NewPostList(1)
	err := pl.Fetch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPostListNext(t *testing.T) {
	pl := NewPostList(1)
	err := pl.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	err = pl.Next()
	if err != nil {
		t.Fatal(err)
	}
}
