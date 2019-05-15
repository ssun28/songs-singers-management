package data

import (
	"fmt"
	"sync"
)

type SynSongsLibrary struct {
	slMap map[int64]string
	mux   sync.Mutex
}

func (sl *SynSongsLibrary) Initial() {
	sl.slMap = make(map[int64]string)
}

func (sl *SynSongsLibrary) Add(songJson string) {
	sl.mux.Lock()
	num := len(sl.slMap)
	sl.slMap[int64(num+1)] = songJson
	sl.mux.Unlock()
}

func (sl *SynSongsLibrary) Show() string {
	sl.mux.Lock()
	defer sl.mux.Unlock()
	rs := "Here are the songs in the songsLibrary:\n"
	for num, songJson := range sl.slMap {
		rs += fmt.Sprintf("Number %v: ", num)
		rs += fmt.Sprintf("SongInfo:%s\n", songJson)
	}
	return rs
}
