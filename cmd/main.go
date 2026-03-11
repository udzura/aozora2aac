package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"
)

// REF: https://cloud.google.com/text-to-speech/pricing

func main() {
	// 1. フラグの設定
	inputPath := flag.String("i", "input.txt", "入力ファイルのパス")
	outputPrefix := flag.String("o", "out-", "出力ファイル名の接頭辞")
	chunkSize := flag.Int("size", 1500, "1ファイルあたりの最大文字数(目安)")
	flag.Parse()

	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY") // Google Cloud TTSでも共通で使えます
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY を設定してください")
	}

	// 2. TTSクライアントの初期化
	client, err := texttospeech.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 3. テキストの読み込みとクレンジング
	data, err := os.ReadFile(*inputPath)
	if err != nil {
		log.Fatal("ファイル読み込みエラー:", err)
	}
	cleanedText := cleanAozoraText(string(data))

	// 4. テキストを分割
	chunks := splitText(cleanedText, *chunkSize)
	fmt.Printf("全 %d 個のセグメントに分割しました。処理を開始します...\n", len(chunks))

	// 5. 各チャンクをループして音声合成
	for i, chunk := range chunks {
		outputName := fmt.Sprintf("%s%03d.mp3", *outputPrefix, i+1)
		fmt.Printf("[%d/%d] %s を生成中...\n", i+1, len(chunks), outputName)

		req := &texttospeechpb.SynthesizeSpeechRequest{
			Input: &texttospeechpb.SynthesisInput{
				InputSource: &texttospeechpb.SynthesisInput_Text{Text: chunk},
			},
			Voice: &texttospeechpb.VoiceSelectionParams{
				LanguageCode: "ja-JP",
				// ボイス ID を指定
				Name:      "ja-JP-Standard-D",
				ModelName: "chirp_3_hd",
			},
			AudioConfig: &texttospeechpb.AudioConfig{
				AudioEncoding: texttospeechpb.AudioEncoding_MP3,
				SpeakingRate:  1.1,
			},
		}

		var resp *texttospeechpb.SynthesizeSpeechResponse
		for attempt := 1; attempt <= 3; attempt++ {
			resp, err = client.SynthesizeSpeech(ctx, req)
			if err == nil {
				break
			}
			log.Printf("セグメント %d の合成に失敗しました (試行 %d/3): %v", i+1, attempt, err)
		}
		if err != nil {
			log.Fatalf("セグメント %d の合成を3回試行しましたが失敗しました。中断します", i+1)
		}

		if err := os.WriteFile(outputName, resp.AudioContent, 0644); err != nil {
			log.Fatal("保存失敗:", err)
		}
	}

	fmt.Println("すべての処理が完了しました。")
}

// 青空文庫のクレンジング
func cleanAozoraText(input string) string {
	reRuby := regexp.MustCompile(`《.*?》`)
	reRubyStart := regexp.MustCompile(`｜`)
	reNotes := regexp.MustCompile(`［＃.*?］`)
	s := reNotes.ReplaceAllString(input, "")
	s = reRuby.ReplaceAllString(s, "")
	s = reRubyStart.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// テキストを適切な長さで分割
func splitText(text string, targetSize int) []string {
	var chunks []string
	runes := []rune(text)
	totalLen := len(runes)
	start := 0

	for start < totalLen {
		end := start + targetSize
		if end >= totalLen {
			chunks = append(chunks, string(runes[start:]))
			break
		}

		// 区切りのいい改行を探す
		foundNewline := false
		for i := end; i < totalLen && i < end+500; i++ {
			if runes[i] == '\n' {
				end = i + 1
				foundNewline = true
				break
			}
		}

		if !foundNewline {
			end = start + targetSize
		}

		chunks = append(chunks, string(runes[start:end]))
		start = end
	}
	return chunks
}
