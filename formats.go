package main

/*
 * Copyright 2015 Albert P. Tobey <atobey@datastax.com> @AlTobey
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * pcstat.go - page cache stat
 *
 * uses the mincore(2) syscall to find out which pages (almost always 4k)
 * of a file are currently cached in memory
 *
 */

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tobert/pcstat"
)

type PcStatusList []pcstat.PcStatus

func (a PcStatusList) Len() int {
	return len(a)
}
func (a PcStatusList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a PcStatusList) Less(i, j int) bool {
	return a[j].Cached < a[i].Cached
}

func (stats PcStatusList) FormatUnicode() {
	maxName := stats.maxNameLen()

	// create horizontal grid line
	pad := strings.Repeat("─", maxName+2)
	top := fmt.Sprintf("┌%s┬────────────────┬────────────┬─────────────┬─────────────┬─────────┐", pad)
	hr := fmt.Sprintf("├%s┼────────────────┼────────────┼─────────────┼─────────────┼─────────┤", pad)
	bot := fmt.Sprintf("└%s┴────────────────┴────────────┴─────────────┴─────────────┴─────────┘", pad)

	var size_sum, page_sum, cached_sum int64

	fmt.Println(top)

	// -nohdr may be chosen to save 2 lines of precious vertical space
	if !nohdrFlag {
		pad = strings.Repeat(" ", maxName-4)
		fmt.Printf("│ Name%s │ Size (bytes)   │ File Pages │ Cached Page │ Cached Size │ Percent │\n", pad)
		fmt.Println(hr)
	}

	for _, pcs := range stats {
		pad = strings.Repeat(" ", maxName-len(pcs.Name))

		// %07.3f was chosen to make it easy to scan the percentages vertically
		// I tried a few different formats only this one kept the decimals aligned
		fmt.Printf("│ %s%s │ %-15d│ %-11d│ %-12d│ %-12d│ %07.3f │\n",
			pcs.Name, pad, pcs.Size, pcs.Pages, pcs.Cached, pcs.Cached*4096, pcs.Percent)

		size_sum += pcs.Size
		page_sum += int64(pcs.Pages)
		cached_sum += int64(pcs.Cached)
	}

	fmt.Println(hr)
	pad = strings.Repeat(" ", maxName-len("Sum"))
	fmt.Printf("│ %s%s │ %-15d│ %-11d│ %-12d│ %-12s│ %07.3f │\n",
		"Sum", pad, size_sum, page_sum, cached_sum, Size_to_string(int64(cached_sum*4096)),
		(float64(cached_sum)/float64(page_sum))*100.00)
	fmt.Println(bot)
}

func (stats PcStatusList) FormatText() {
	maxName := stats.maxNameLen()

	// create horizontal grid line
	pad := strings.Repeat("-", maxName+2)
	top := fmt.Sprintf("+%s+----------------+------------+-------------+-------------+---------+", pad)
	hr := fmt.Sprintf("|%s+----------------+------------+-------------+-------------+---------|", pad)
	bot := fmt.Sprintf("+%s+----------------+------------+-------------+-------------+---------+", pad)
	var size_sum, page_sum, cached_sum int64

	fmt.Println(top)

	// -nohdr may be chosen to save 2 lines of precious vertical space
	if !nohdrFlag {
		pad = strings.Repeat(" ", maxName-4)
		fmt.Printf("| Name%s | Size (bytes)   | File Pages | Cached Page | Cached Size | Percent |\n", pad)
		fmt.Println(hr)
	}

	for _, pcs := range stats {
		pad = strings.Repeat(" ", maxName-len(pcs.Name))

		// %07.3f was chosen to make it easy to scan the percentages vertically
		// I tried a few different formats only this one kept the decimals aligned
		fmt.Printf("| %s%s | %-15d| %-11d| %-12d| %-12d| %07.3f |\n",
			pcs.Name, pad, pcs.Size, pcs.Pages, pcs.Cached, pcs.Cached*4096, pcs.Percent)

		size_sum += pcs.Size
		page_sum += int64(pcs.Pages)
		cached_sum += int64(pcs.Cached)
	}
	fmt.Println(hr)
	pad = strings.Repeat(" ", maxName-len("Sum"))
	fmt.Printf("│ %s%s │ %-15d│ %-11d│ %-12d│ %-12s│ %07.3f │\n",
		"Sum", pad, size_sum, page_sum, cached_sum, Size_to_string(int64(cached_sum*4096)),
		(float64(cached_sum)/float64(page_sum))*100.00)
	fmt.Println(bot)
}

func (stats PcStatusList) FormatPlain() {
	maxName := stats.maxNameLen()

	var size_sum, page_sum, cached_sum int64

	// -nohdr may be chosen to save 2 lines of precious vertical space
	if !nohdrFlag {
		pad := strings.Repeat(" ", maxName-4)
		fmt.Printf("Name%s  Size (bytes)    File Pages  Cached Page  Cached Size  Percent \n", pad)
	}

	for _, pcs := range stats {
		pad := strings.Repeat(" ", maxName-len(pcs.Name))

		// %07.3f was chosen to make it easy to scan the percentages vertically
		// I tried a few different formats only this one kept the decimals aligned
		fmt.Printf("%s%s  %-15d %-11d %-12d %-12d %07.3f\n",
			pcs.Name, pad, pcs.Size, pcs.Pages, pcs.Cached, pcs.Cached*4096, pcs.Percent)

		size_sum += pcs.Size
		page_sum += int64(pcs.Pages)
		cached_sum += int64(pcs.Cached)
	}

	pad := strings.Repeat(" ", maxName-len("Sum"))
	fmt.Printf("%s%s  %-15d %-11d %-12d %-12s %07.3f\n",
		"Sum", pad, size_sum, page_sum, cached_sum, Size_to_string(int64(cached_sum*4096)),
		(float64(cached_sum)/float64(page_sum))*100.00)
}

func (stats PcStatusList) FormatTerse() {

	if !nohdrFlag {
		fmt.Println("name,size,timestamp,mtime,filepages,cachedpage,cachesize,percent")
	}
	for _, pcs := range stats {
		time := pcs.Timestamp.Unix()
		mtime := pcs.Mtime.Unix()
		fmt.Printf("%s,%d,%d,%d,%d,%d,%d,%g\n",
			pcs.Name, pcs.Size, time, mtime, pcs.Pages, pcs.Cached, pcs.Cached*4096, pcs.Percent)
	}
}

func (stats PcStatusList) FormatJson(clearpps bool) {
	// clear the per-page status when requested
	// emits an empty "status": [] field in the JSON when disabled, but NBD.
	if clearpps {
		for _, pcs := range stats {
			pcs.PPStat = nil
		}
	}

	b, err := json.Marshal(stats)
	if err != nil {
		log.Fatalf("JSON formatting failed: %s\n", err)
	}
	os.Stdout.Write(b)
	fmt.Println("")
}

// references:
// http://www.unicode.org/charts/PDF/U2580.pdf
// https://github.com/puppetlabs/mcollective-puppet-agent/blob/master/application/puppet.rb#L143
// https://github.com/holman/spark
func (stats PcStatusList) FormatHistogram() {
	ws := getwinsize()
	maxName := stats.maxNameLen()

	// block elements are wider than characters, so only use 1/2 the available columns
	buckets := (int(ws.ws_col)-maxName)/2 - 10

	for _, pcs := range stats {
		pad := strings.Repeat(" ", maxName-len(pcs.Name))
		fmt.Printf("%s%s % 8d ", pcs.Name, pad, pcs.Pages)

		// when there is enough room display on/off for every page
		if buckets > pcs.Pages {
			for _, v := range pcs.PPStat {
				if v {
					fmt.Print("\u2588") // full block = 100%
				} else {
					fmt.Print("\u2581") // lower 1/8 block
				}
			}
		} else {
			bsz := pcs.Pages / buckets
			fbsz := float64(bsz)
			total := 0.0
			for i, v := range pcs.PPStat {
				if v {
					total++
				}

				if (i+1)%bsz == 0 {
					avg := total / fbsz
					if total == 0 {
						fmt.Print("\u2581") // lower 1/8 block = 0
					} else if avg < 0.16 {
						fmt.Print("\u2582") // lower 2/8 block
					} else if avg < 0.33 {
						fmt.Print("\u2583") // lower 3/8 block
					} else if avg < 0.50 {
						fmt.Print("\u2584") // lower 4/8 block
					} else if avg < 0.66 {
						fmt.Print("\u2585") // lower 5/8 block
					} else if avg < 0.83 {
						fmt.Print("\u2586") // lower 6/8 block
					} else if avg < 1.00 {
						fmt.Print("\u2587") // lower 7/8 block
					} else {
						fmt.Print("\u2588") // full block = 100%
					}

					total = 0
				}
			}
		}
		fmt.Println("")
	}
}

/*

	// convert long paths to their basename with the -bname flag
	// this overwrites the original filename in pcs but it doesn't matter since
	// it's not used to access the file again -- and should not be!
	if bnameFlag {
		pcs.Name = path.Base(fname)
	}
*/

// maxNameLen returns the len of longest filename in the stat list
// if the bnameFlag is set, this will return the max basename len
func (stats PcStatusList) maxNameLen() int {
	var maxName int
	for _, pcs := range stats {
		if len(pcs.Name) > maxName {
			maxName = len(pcs.Name)
		}
	}

	if maxName < 5 {
		maxName = 5
	}
	return maxName
}

func Size_to_string(cachesize int64) string {
	var KB, MB, GB, TB, PB int64
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
	PB = 1024 * TB

	if cachesize > PB {
		return fmt.Sprintf("%.2f PB", (float64(cachesize) / float64(PB)))
	}
	if cachesize > TB {
		return fmt.Sprintf("%.2f TB", (float64(cachesize) / float64(TB)))
	}
	if cachesize > GB {
		return fmt.Sprintf("%.2f GB", (float64(cachesize) / float64(GB)))
	}
	if cachesize > MB {
		return fmt.Sprintf("%.2f MB", (float64(cachesize) / float64(MB)))
	}

	return fmt.Sprintf("%.2f KB", (float64(cachesize) / float64(KB)))
}
