package main

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/skratchdot/open-golang/open"
)

const (
	chromeHistoryPath = "/Users/hitoshi/Library/Application Support/Google/Chrome/Default/History"
	tmpFilePath       = ""
	tmpFileName       = "tmpHistory"
	querySize         = 10
)

func copyHistoryFile() string {
	// chromeの検索履歴はsqlite形式
	// lockされてるため、tmpFileにコピーして用いる

	// chromeの履歴を読み込み(sqlite)
	chromeHistory, err := os.Open(chromeHistoryPath)
	if err != nil {
		log.Fatal(err)
	}
	defer chromeHistory.Close()

	tmpFile, err := ioutil.TempFile(tmpFilePath, tmpFileName)
	// tmpFileの削除はmainでする
	if err != nil {
		log.Fatal(err)
	}

	// 作成した一時ファイル(tmpFile)にchromeHistoryをコピー
	_, err = io.Copy(tmpFile, chromeHistory)
	if err != nil {
		log.Fatal(err)
	}

	return tmpFile.Name()
}

func queryHistory(filename string) []History {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//var query = "select * from urls ORDER BY id DESC LIMIT 10"
	var query = "select * from urls ORDER BY id DESC LIMIT " + strconv.Itoa(querySize)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var historys = []History{}

	for rows.Next() {
		history := History{}
		err = rows.Scan(
			&history.id,
			&history.url,
			&history.title,
			&history.visitCount,
			&history.typedCount,
			&history.lastVisitTime,
			&history.hidden,
		)
		if err != nil {
			log.Fatal(err)
		}
		historys = append(historys, history)
	}

	return historys
}

type History struct {
	id            int
	url           string
	title         string
	visitCount    int
	typedCount    int
	lastVisitTime int
	hidden        int
}

func main() {
	historyFileName := copyHistoryFile()
	defer os.Remove(historyFileName)

	historys := queryHistory(historyFileName)

	for index, history := range historys {
		fmt.Printf(" # %-5d <-- %s \n", index, history.title)
	}

	fmt.Print("\n --- What # ? >> ")
	var t int
	fmt.Scan(&t)
	fmt.Println(" --- Open: ", historys[t].title)
	//TODO: 入力値に対するエラー処理

	//TODO: indexではなくコマンドで開けるようにする(jj, i, k等)
	//TODO: 閉じたタブだけ表示する
	//TODO: 被ってるのは削除する
	open.Run(historys[t].url)
}
