package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/fumiama/pics1f"
)

func main() {
	index := flag.Uint("s", 1, "start index")
	dir := flag.String("d", "./dl", "download dir")
	retry := flag.Uint("r", 3, "retry times")
	override := flag.Bool("o", true, "override")
	pageid := flag.Int("p", 0, "only download this page")
	mult := flag.Uint("m", 4, "download thread multiplier")
	flag.Parse()
	if *pageid > 0 {
		p, err := pics1f.NewPageShortLink(*pageid)
		if err != nil {
			fmt.Println("[ERROR] fetching page id", *pageid, "->", err)
			os.Exit(1)
		}
		err = p.DownloadContentsTo(*dir, (int)(*retry), *override, int(*mult), func(err error) {
			fmt.Println("[ERROR] downloading page id", *pageid, "->", err)
		})
		if err != nil {
			fmt.Println("[ERROR] calling to download page id", *pageid, "->", err)
		}
		return
	}
	pl := pics1f.NewPostList((int)(*index))
	err := pl.Fetch()
	for err == nil {
		fmt.Println("[IFNO] start download index", pl.Index, "/", pl.Total, "...")
		for i, p := range pl.Pages {
			err = p.Fetch()
			if err != nil {
				fmt.Println("[ERROR] fetching index", pl.Index, "page", i+1, "->", err)
			}
			err = p.DownloadContentsTo(*dir, (int)(*retry), *override, int(*mult), func(err error) {
				fmt.Println("[ERROR] downloading index", pl.Index, "page", i+1, "->", err)
			})
			if err != nil {
				fmt.Println("[ERROR] calling to download index", pl.Index, "page", i+1, "->", err)
			}
			fmt.Printf("\r[INFO] progress: %d / %d                    ", i+1, len(pl.Pages))
		}
		fmt.Print("\n")
		err = pl.Next()
	}
	if err != nil && err != io.EOF {
		fmt.Println("[ERROR] fetching index", pl.Index, "->", err)
	}
}
