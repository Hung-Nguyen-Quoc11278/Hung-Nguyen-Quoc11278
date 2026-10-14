package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/notnil/chess"
)

// GameState lưu trữ trạng thái bàn cờ dưới dạng chuỗi FEN
type GameState struct {
	FEN string `json:"fen"`
}

const (
	StateFile  = "state.json"
	ReadmeFile = "README.md"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Sử dụng: go run main.go 'Chess: <nước_đi>'")
		os.Exit(1)
	}

	inputArg := os.Args[1] // Ví dụ: "Chess: e2e4"

	// Tách lấy chuỗi nước đi phía sau chữ "Chess:"
	parts := strings.Split(inputArg, ":")
	if len(parts) < 2 {
		fmt.Println("Định dạng đầu vào không hợp lệ.")
		os.Exit(1)
	}
	moveStr := strings.TrimSpace(parts[1])

	// 1. Tải trạng thái game hiện tại từ state.json (hoặc tạo ván mới nếu chưa có)
	game := loadGame()

	// 2. Thực hiện nước đi người chơi vừa nhập
	err := game.MoveStr(moveStr)
	if err != nil {
		fmt.Printf("Nước đi không hợp lệ (%s): %v\n", moveStr, err)
		os.Exit(1)
	}

	// 3. Lưu trạng thái FEN mới vào file state.json
	saveGame(game)

	// 4. Render lại giao diện bàn cờ và ghi đè file README.md
	updateReadme(game)

	fmt.Printf("Thực hiện thành công nước đi: %s\n", moveStr)
}

// loadGame đọc trạng thái từ file state.json hoặc khởi tạo bàn cờ mặc định
func loadGame() *chess.Game {
	if _, err := os.Stat(StateFile); os.IsNotExist(err) {
		return chess.NewGame()
	}

	data, err := os.ReadFile(StateFile)
	if err != nil {
		return chess.NewGame()
	}

	var state GameState
	if err := json.Unmarshal(data, &state); err != nil || state.FEN == "" {
		return chess.NewGame()
	}

	fenFunc, err := chess.FEN(state.FEN)
	if err != nil {
		return chess.NewGame()
	}

	return chess.NewGame(fenFunc)
}

// saveGame lưu FEN hiện tại của bàn cờ vào state.json
func saveGame(game *chess.Game) {
	state := GameState{FEN: game.FEN()}
	data, _ := json.MarshalIndent(state, "", "  ")
	_ = os.WriteFile(StateFile, data, 0644)
}

// updateReadme vẽ lại bàn cờ và cập nhật nội dung README.md bằng cách nối chuỗi an toàn
func updateReadme(game *chess.Game) {
	boardStr := game.Position().Board().Draw()
	fen := game.FEN()

	content := "# Chào mừng bạn đến với Profile của tôi! 🚀\n\n" +
		"## ♟️ Trò chơi Cờ vua trực tiếp trên README\n\n" +
		"Tạo một **Issue** mới với tiêu đề dạng `Chess: <nước_đi>` (ví dụ: `Chess: e2e4`) để thực hiện nước đi tiếp theo!\n\n" +
		"### Bàn cờ hiện tại:\n" +
		"```text\n" + boardStr + "\n```\n\n" +
		"**FEN hiện tại:** `" + fen + "`\n\n" +
		"---\n" +
		"*Được vận hành tự động bởi **Golang** & **GitHub Actions**.*\n"

	_ = os.WriteFile(ReadmeFile, []byte(content), 0644)
}
