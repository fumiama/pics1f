package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/fumiama/pics1f"
)

func main() {
	index := flag.Uint("s", 1, "start index")
	dir := flag.String("d", "./dl", "download dir")
	retry := flag.Uint("r", 3, "retry times")
	override := flag.Bool("o", true, "override")
	flag.Parse()
	pl := pics1f.NewPostList((int)(*index))
	err := pl.Fetch()
	for err == nil {
		fmt.Println("[IFNO] start download index", pl.Index, "/", pl.Total, "...")
		for i, p := range pl.Pages {
			err = p.Fetch()
			if err != nil {
				fmt.Println("[ERROR] fetching index", pl.Index, "page", i, "->", err)
			}
			err = p.DownloadContentsTo(*dir, (int)(*retry), *override)
			if err != nil {
				fmt.Println("[ERROR] downloading index", pl.Index, "page", i, "->", err)
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
