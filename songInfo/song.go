package songInfo

import (
	"encoding/json"
	"log"
)

type Song struct {
	SongName         string `json:"songName"`
	Artist           string `json:"artist"`
	Album            string `json:"album"`
	AlbumArtist      string `json:"albumArtist"`
	AlbumTrackNumber int32  `json:"albumTrackNumber"`
	Composer         string `json:"composer"`
	Genre            string `json:"genre"`
	Year             string `json:"year"`
	Lyrics           string `json:"lyrics"`
	Comments         string `json:"comments"`
	TransactionFee   string `json:"transactionFee"`
}

func NewSong(songName string, artist string, album string, albumArtist string, albumTrackNumber int32,
	composer string, genre string, year string, lyrics string, comments string,
	transactionFee string) Song {
	newSong := Song{songName, artist, album, albumArtist, albumTrackNumber,
		composer, genre, year, lyrics, comments, transactionFee}

	return newSong
}

func (song *Song) EncodeSongToJSON() (string, error) {
	songJson, err := json.Marshal(song)
	if err != nil {
		log.Fatal("encode to songJson error:", err)
	}

	return string(songJson), err
}

func DecodeSongFromJson(songJson string) (Song, error) {
	var song Song
	err := json.Unmarshal([]byte(songJson), &song)
	if err != nil {
		log.Fatal("decode songJson to songInfo error:", err)
	}

	return song, err
}
