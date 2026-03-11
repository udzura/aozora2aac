package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// 1. APIキーの設定（環境変数から取得）
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY を設定してください")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 2. モデルの初期化（Gemini 2.0 Flash を推奨）
	model := client.GenerativeModel("gemini-2.0-flash")

	// 3. テキストファイルの読み込み
	// input.txt を読み込む前提
	data, err := os.ReadFile("input.txt")
	if err != nil {
		log.Fatal("input.txt が見つかりません:", err)
	}
	text := string(data)

	// 4. 生成プロンプトの作成
	// 音声出力を指示するためにシステムプロンプト的な文脈を与えます
	prompt := []genai.Part{
		genai.Text("以下のテキストを、通勤中に聞き取りやすい落ち着いたナレーターの声で朗読してください。出力は音声データのみを返してください。"),
		genai.Text(text),
	}

	// 5. 音声生成リクエスト
	// ※注意: 現時点のSDKでは、モデルの設定で response_mime_type を audio/mpeg に指定します
	model.ResponseMIMEType = "audio/mpeg"

	fmt.Println("音声を生成中...（時間がかかる場合があります）")
	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		log.Fatal("生成エラー:", err)
	}

	// 6. 結果をファイルに保存
	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			if blob, ok := part.(genai.Blob); ok {
				err := os.WriteFile("output.mp3", blob.Data, 0644)
				if err != nil {
					log.Fatal("ファイル保存エラー:", err)
				}
				fmt.Println("完了！ output.mp3 を作成しました。")
				return
			}
		}
	}
	fmt.Println("音声データが取得できませんでした。")
}
