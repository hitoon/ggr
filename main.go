package main

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/skratchdot/open-golang/open"
)

const (
	chromeHistoryPath = "/Users/hitoshi/Library/Application Support/Google/Chrome/Default/History"
	tmpFilePath       = ""
	tmpFileName       = "tmpHistory"
	querySize         = 10
)

var commandKeys = []string{"j", "i", "k", "n"}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func makeCommandList(historySize int) []string {
	var commandList = []string{}
	for i := 0; i < historySize; i++ {
		remainder := i % len(commandKeys)  // 余り
		quotient := i/len(commandKeys) + 1 // 商
		key := commandKeys[remainder]
		commandList = append(commandList, strings.Repeat(key, quotient))
	}
	return commandList
}

func getCommandIndex(str string, list []string) int {
	for index, value := range list {
		if str == value {
			return index
		}
	}
	return -1
}

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

	var query = "SELECT * FROM urls ORDER BY id DESC LIMIT " +
		strconv.Itoa(querySize)

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var historys = []History{}
	var titleList = []string{}
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

		// タイトルが被っているものを削除
		if !stringInSlice(history.title, titleList) {
			titleList = append(titleList, history.title)
			historys = append(historys, history)
		}
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

	// indexではなくコマンドで開けるようにする(jj, i, k等)
	// コマンドのリストを作成(historysの数だけ)
	commandList := makeCommandList(len(historys))

	for index, history := range historys {
		fmt.Printf(" # %-4d %-6s <-- %s \n", index, commandList[index], history.title)
	}

	fmt.Print("\n --- What # ? >> ")
	var t string
	fmt.Scan(&t)
	//TODO: 入力値に対するエラー処理

	cmdidx := getCommandIndex(t, commandList)
	fmt.Println(" --- Open: ", historys[cmdidx].title)

	// 閉じたタブだけ表示する -> 技術的に無理
	/*マルウェアを防ぐため、Current TabsがSSNSという形式でFormatされている
	  Chromagnonがそれをリバースエンジニアリングするプロジェクト*/

	open.Run(historys[cmdidx].url) // 選択したタブを開く
}
