package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/notnil/chess"
)

type GameState struct {
	FEN        string `json:"fen"`
	LastPlayer string `json:"last_player"`
}

const (
	StateFile  = "state.json"
	ReadmeFile = "README.md"
	SvgFile    = "board.svg"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Sử dụng: go run main.go 'Chess: <nước_đi_hoặc_reset>' [tên_người_chơi]")
		os.Exit(1)
	}

	inputArg := os.Args[1]
	player := "Local Player"
	if len(os.Args) >= 3 {
		player = os.Args[2]
	}

	parts := strings.Split(inputArg, ":")
	if len(parts) < 2 {
		fmt.Println("Định dạng đầu vào không hợp lệ.")
		os.Exit(1)
	}
	command := strings.TrimSpace(strings.ToLower(parts[1]))

	// Xử lý khi người dùng yêu cầu reset trận cờ
	if command == "reset" {
		game := chess.NewGame()
		saveGame(game, "Chua_Co")
		generateSVGBoard(game, SvgFile)
		updateReadme(game, "Chua_Co")
		fmt.Println("Đã reset bàn cờ về trạng thái ban đầu thành công!")
		return
	}

	// 1. Tải trạng thái game hiện tại từ state.json
	game, _ := loadGame()

	// 2. Validate nước đi cực kỳ nghiêm ngặt theo danh sách ValidMoves thực tế
	var move *chess.Move
	validMoves := game.ValidMoves()
	moveStr := strings.TrimSpace(parts[1])

	for _, m := range validMoves {
		if m.String() == moveStr || m.String() == strings.ReplaceAll(moveStr, "-", "") {
			move = m
			break
		}
	}

	if move == nil {
		decodedMove, decodeErr := chess.AlgebraicNotation{}.Decode(game.Position(), moveStr)
		if decodeErr == nil && decodedMove != nil {
			for _, m := range validMoves {
				if m.String() == decodedMove.String() {
					move = m
					break
				}
			}
		}
	}

	if move == nil {
		fmt.Printf("Nước đi không hợp lệ hoặc phạm luật theo bàn cờ hiện tại (%s)\n", moveStr)
		os.Exit(1)
	}

	err := game.Move(move)
	if err != nil {
		fmt.Printf("Không thể thực hiện nước đi (%s): %v\n", moveStr, err)
		os.Exit(1)
	}

	// 3. Chỉ lưu trạng thái khi nước đi hợp lệ
	saveGame(game, player)
	generateSVGBoard(game, SvgFile)
	updateReadme(game, player)

	fmt.Printf("Thực hiện thành công nước đi: %s bởi @%s\n", moveStr, player)
}

func loadGame() (*chess.Game, string) {
	if _, err := os.Stat(StateFile); os.IsNotExist(err) {
		return chess.NewGame(), "Chưa có"
	}

	data, err := os.ReadFile(StateFile)
	if err != nil {
		return chess.NewGame(), "Chưa có"
	}

	var state GameState
	if err := json.Unmarshal(data, &state); err != nil || state.FEN == "" {
		return chess.NewGame(), "Chưa có"
	}

	fenFunc, err := chess.FEN(state.FEN)
	if err != nil {
		return chess.NewGame(), state.LastPlayer
	}

	lastPlayer := state.LastPlayer
	if lastPlayer == "" || lastPlayer == "Chua_Co" {
		return chess.NewGame(), "Chưa có"
	}

	return chess.NewGame(fenFunc), lastPlayer
}

func saveGame(game *chess.Game, lastPlayer string) {
	state := GameState{
		FEN:        game.FEN(),
		LastPlayer: lastPlayer,
	}
	data, _ := json.MarshalIndent(state, "", "  ")
	_ = os.WriteFile(StateFile, data, 0644)
}

var pieceUnicode = map[chess.Piece]string{
	chess.WhiteKing:   "♔",
	chess.WhiteQueen:  "♕",
	chess.WhiteRook:   "♖",
	chess.WhiteBishop: "♗",
	chess.WhiteKnight: "♘",
	chess.WhitePawn:   "♙",
	chess.BlackKing:   "♚",
	chess.BlackQueen:  "♛",
	chess.BlackRook:   "♜",
	chess.BlackBishop: "♝",
	chess.BlackKnight: "♞",
	chess.BlackPawn:   "♟",
}

func generateSVGBoard(game *chess.Game, filename string) {
	const boardSize = 480
	const tileSize = 50
	const offset = 40

	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="100%%" height="100%%">`, boardSize, boardSize))

	svg.WriteString(`
		<style>
			.tile-light { fill: #f0d9b5; }
			.tile-dark { fill: #b58863; }
			.label { font-family: monospace; font-size: 14px; fill: #666; font-weight: bold; }
			.piece { font-size: 38px; text-anchor: middle; dominant-baseline: central; }
		</style>`)

	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#312e2b" rx="8"/>`, boardSize, boardSize))

	board := game.Position().Board()
	files := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	ranks := []string{"8", "7", "6", "5", "4", "3", "2", "1"}

	for r := 0; r < 8; r++ {
		for f := 0; f < 8; f++ {
			x := offset + f*tileSize
			y := offset + r*tileSize
			tileClass := "tile-light"
			if (r+f)%2 == 1 {
				tileClass = "tile-dark"
			}
			svg.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" class="%s"/>`, x, y, tileSize, tileSize, tileClass))
		}
	}

	for i := 0; i < 8; i++ {
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label" text-anchor="middle">%s</text>`, offset+i*tileSize+tileSize/2, 25, files[i]))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label" text-anchor="middle">%s</text>`, offset+i*tileSize+tileSize/2, offset+8*tileSize+25, files[i]))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label" dominant-baseline="middle">%s</text>`, 15, offset+i*tileSize+tileSize/2, ranks[i]))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label" dominant-baseline="middle">%s</text>`, offset+8*tileSize+12, offset+i*tileSize+tileSize/2, ranks[i]))
	}

	for r := 0; r < 8; r++ {
		for f := 0; f < 8; f++ {
			sq := chess.Square((7-r)*8 + f)
			piece := board.Piece(sq)
			if piece != chess.NoPiece {
				x := offset + f*tileSize + tileSize/2
				y := offset + r*tileSize + tileSize/2
				unicodeChar := pieceUnicode[piece]

				svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="piece" fill="rgba(0,0,0,0.3)">%s</text>`, x+1, y+2, unicodeChar))
				fillColor := "#ececec"
				if piece.Color() == chess.Black {
					fillColor = "#1a1a1a"
				}
				svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="piece" fill="%s">%s</text>`, x, y, fillColor, unicodeChar))
			}
		}
	}

	svg.WriteString(`</svg>`)
	_ = os.WriteFile(filename, []byte(svg.String()), 0644)
}

func formatMoveHistory(game *chess.Game) string {
	moves := game.Moves()
	if len(moves) == 0 {
		return "_Chưa có nước đi nào._"
	}

	var sb strings.Builder
	for i, move := range moves {
		if i%2 == 0 {
			sb.WriteString(fmt.Sprintf("%d. %s ", (i/2)+1, move.String()))
		} else {
			sb.WriteString(fmt.Sprintf("%s  ", move.String()))
		}
	}
	return sb.String()
}

func updateReadme(game *chess.Game, lastPlayer string) {
	fen := game.FEN()
	turn := "⚪ Quân Trắng (White)"
	if game.Position().Turn() == chess.Black {
		turn = "⚫ Quân Đen (Black)"
	}

	statusStr := "Đang diễn ra"
	outcome := game.Outcome()
	if outcome != chess.NoOutcome {
		if outcome == chess.WhiteWon {
			statusStr = "🏆 Trắng thắng (White Win)!"
		} else if outcome == chess.BlackWon {
			statusStr = "🏆 Đen thắng (Black Win)!"
		} else if outcome == chess.Draw {
			statusStr = "🤝 Ván đấu kết thúc với kết quả Hòa (Draw)!"
		}
	}

	historyStr := formatMoveHistory(game)

	playerDisplay := "Chưa có"
	if lastPlayer != "" && lastPlayer != "Chua_Co" {
		playerDisplay = fmt.Sprintf("[@`%s`](https://github.com/%s)", lastPlayer, lastPlayer)
	}

	content := "# Chào mừng bạn đến với Profile của tôi! 🚀\n\n" +
		"## ♟️ Trợ thủ Cờ vua trực tiếp trên README\n\n" +
		"Tạo một **Issue** mới với tiêu đề dạng `Chess: <nước_đi>` (ví dụ: `Chess: e4`, `Chess: e2e4`) hoặc `Chess: reset` để làm mới bàn cờ!\n\n" +
		"### Trạng thái ván đấu:\n" +
		"- **Lượt đi:** " + turn + "\n" +
		"- **Người chơi gần nhất:** 👤 " + playerDisplay + "\n" +
		"- **Tình trạng:** " + statusStr + "\n\n" +
		"### Bàn cờ hiện tại:\n\n" +
		"<p align=\"center\">\n" +
		"  <img src=\"board.svg\" width=\"400\" alt=\"Chess Board\">\n" +
		"</p>\n\n" +
		"### 📜 Lịch sử nước đi:\n" +
		"```text\n" + historyStr + "\n```\n\n" +
		"**FEN hiện tại:** `" + fen + "`\n\n" +
		"---\n" +
		"*Được vận hành tự động bởi **Golang** & **GitHub Actions**.*\n"

	_ = os.WriteFile(ReadmeFile, []byte(content), 0644)
}
