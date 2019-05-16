package data

import (
	"fmt"
	"strconv"
	"sync"
)

type SynSongsLibrary struct {
	slMap map[string]string
	mux   sync.Mutex
}

func (sl *SynSongsLibrary) Initial() {
	sl.slMap = make(map[string]string)
}

func (sl *SynSongsLibrary) Add(id int32, timestamp int64, songJson string) {
	sl.mux.Lock()
	key := strconv.FormatInt(int64(id), 10) + "id" + strconv.FormatInt(timestamp, 10)
	sl.slMap[key] = songJson
	sl.mux.Unlock()
}

func (sl *SynSongsLibrary) Show() string {
	sl.mux.Lock()
	defer sl.mux.Unlock()
	rs := "Here are the songs in the songsLibrary:\n"
	count := 1
	for key, songJson := range sl.slMap {
		rs += fmt.Sprintf("Num:%v: ", count)
		rs += fmt.Sprintf("id:%s: ", key)
		rs += fmt.Sprintf("SongInfo:%s\n", songJson)
		count++
	}
	return rs
}
