package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	// コマンドライン引数の設定
	inputPath := flag.String("i", "input.txt", "入力ファイルのパス")
	chunkSize := flag.Int("size", 2000, "分割の目安となる文字数")
	outputPrefix := flag.String("o", "chunk", "出力ファイル名の接頭辞")
	flag.Parse()

	// 1. ファイル読み込み
	content, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ファイルの読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 2. 前処理（前回作成したクレンジング）
	text := cleanAozoraText(string(content))

	// 3. 分割ロジック
	chunks := splitText(text, *chunkSize)

	// 4. ファイル書き出し
	for i, chunk := range chunks {
		fileName := fmt.Sprintf("%s_%03d.txt", *outputPrefix, i+1)
		err := os.WriteFile(fileName, []byte(chunk), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ファイル %s の書き出しに失敗: %v\n", fileName, err)
			continue
		}
		fmt.Printf("Generated: %s (%d characters)\n", fileName, len([]rune(chunk)))
	}
}

// 青空文庫クレンジング
func cleanAozoraText(input string) string {
	reRuby := regexp.MustCompile(`《.*?》`)
	reRubyStart := regexp.MustCompile(`｜`)
	reNotes := regexp.MustCompile(`［＃.*?］`)

	s := reNotes.ReplaceAllString(input, "")
	s = reRuby.ReplaceAllString(s, "")
	s = reRubyStart.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func splitText(text string, targetSize int) []string {
	var chunks []string
	// Goで日本語を扱うため rune に変換
	runes := []rune(text)
	totalLen := len(runes)

	start := 0
	for start < totalLen {
		end := start + targetSize
		if end >= totalLen {
			chunks = append(chunks, string(runes[start:]))
			break
		}

		// 目標地点から、次の改行（\n）が見つかるまでポインタを進める
		// これにより段落の途中で切れるのを防ぐ
		foundNewline := false
		for i := end; i < totalLen; i++ {
			if runes[i] == '\n' {
				end = i + 1
				foundNewline = true
				break
			}
		}

		// もし最後まで改行がなかったら、仕方ないので目標地点で切る
		if !foundNewline {
			end = start + targetSize
		}

		chunks = append(chunks, string(runes[start:end]))
		start = end
	}
	return chunks
}
