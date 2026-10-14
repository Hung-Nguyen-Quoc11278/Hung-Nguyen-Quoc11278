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
	SvgFile    = "board.svg"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Sử dụng: go run main.go 'Chess: <nước_đi>'")
		os.Exit(1)
	}

	inputArg := os.Args[1] // Ví dụ: "Chess: e4" hoặc "Chess: e2e4"

	// Tách lấy chuỗi nước đi phía sau chữ "Chess:"
	parts := strings.Split(inputArg, ":")
	if len(parts) < 2 {
		fmt.Println("Định dạng đầu vào không hợp lệ.")
		os.Exit(1)
	}
	moveStr := strings.TrimSpace(parts[1])

	// 1. Tải trạng thái game hiện tại từ state.json (hoặc tạo ván mới nếu chưa có)
	game := loadGame()

	// 2. Thực hiện nước đi người chơi vừa nhập (hỗ trợ cả SAN lẫn UCI)
	var move *chess.Move
	var err error

	move, err = chess.AlgebraicNotation{}.Decode(game.Position(), moveStr)
	if err != nil {
		for _, m := range game.ValidMoves() {
			if m.String() == moveStr {
				move = m
				err = nil
				break
			}
		}
	}

	if err != nil || move == nil {
		fmt.Printf("Nước đi không hợp lệ (%s): %v\n", moveStr, err)
		os.Exit(1)
	}

	err = game.Move(move)
	if err != nil {
		fmt.Printf("Không thể thực hiện nước đi (%s): %v\n", moveStr, err)
		os.Exit(1)
	}

	// 3. Lưu trạng thái FEN mới vào file state.json
	saveGame(game)

	// 4. Sinh file ảnh SVG bàn cờ và cập nhật file README.md
	generateSVGBoard(game, SvgFile)
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

// Map ký tự quân cờ sang ký hiệu Unicode tương ứng
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

// generateSVGBoard tự động vẽ bàn cờ thành file SVG đẹp mắt
func generateSVGBoard(game *chess.Game, filename string) {
	const boardSize = 480
	const tileSize = 50
	const offset = 40

	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="100%%" height="100%%">`, boardSize, boardSize))

	// Định nghĩa style màu sắc hiện đại (Tone gỗ sáng / tối sang trọng)
	svg.WriteString(`
		<style>
			.tile-light { fill: #f0d9b5; }
			.tile-dark { fill: #b58863; }
			.label { font-family: monospace; font-size: 14px; fill: #666; font-weight: bold; }
			.piece { font-size: 38px; text-anchor: middle; dominant-baseline: central; }
		>
	</style>`)

	// Nền tổng thể
	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#312e2b" rx="8"/>`, boardSize, boardSize))

	board := game.Position().Board()

	// Vẽ các ô vuông bàn cờ và tọa độ A-H, 1-8
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

	// Vẽ nhãn chữ cái và số tọa độ biên
	for i := 0; i < 8; i++ {
		// Nhãn chữ cái (A-H) ở trên và dưới
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label" text-anchor="middle">%s</text>`, offset+i*tileSize+tileSize/2, 25, files[i]))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label" text-anchor="middle">%s</text>`, offset+i*tileSize+tileSize/2, offset+8*tileSize+25, files[i]))

		// Nhãn số (1-8) ở trái và phải
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label" dominant-baseline="middle">%s</text>`, 15, offset+i*tileSize+tileSize/2, ranks[i]))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label" dominant-baseline="middle">%s</text>`, offset+8*tileSize+12, offset+i*tileSize+tileSize/2, ranks[i]))
	}

	// Vẽ các quân cờ lên bàn cờ
	for r := 0; r < 8; r++ {
		for f := 0; f < 8; f++ {
			// Chuyển đổi tọa độ bàn cờ chess (A1-H8)
			sq := chess.Square((7-r)*8 + f)
			piece := board.Piece(sq)
			if piece != chess.NoPiece {
				x := offset + f*tileSize + tileSize/2
				y := offset + r*tileSize + tileSize/2
				unicodeChar := pieceUnicode[piece]

				// Thêm hiệu ứng bóng đổ nhẹ cho quân cờ nổi bật
				svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="piece" fill="rgba(0,0,0,0.3)">%s</text>`, x+1, y+2, unicodeChar))
				// Quân cờ chính thức (Quân trắng màu kem sáng, quân đen màu nâu đậm)
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

// updateReadme cập nhật file README.md hiển thị hình ảnh SVG thay vì text thô
func updateReadme(game *chess.Game) {
	fen := game.FEN()

	content := "# Chào mừng bạn đến với Profile của tôi! 🚀\n\n" +
		"## ♟️ Trò chơi Cờ vua trực tiếp trên README\n\n" +
		"Tạo một **Issue** mới với tiêu đề dạng `Chess: <nước_đi>` (ví dụ: `Chess: e4` hoặc `Chess: e2e4`) để thực hiện nước đi tiếp theo!\n\n" +
		"### Bàn cờ hiện tại:\n\n" +
		"<p align=\"center\">\n" +
		"  <img src=\"board.svg\" width=\"400\" alt=\"Chess Board\">\n" +
		"</p>\n\n" +
		"**FEN hiện tại:** `" + fen + "`\n\n" +
		"---\n" +
		"*Được vận hành tự động bởi **Golang** & **GitHub Actions**.*\n"

	_ = os.WriteFile(ReadmeFile, []byte(content), 0644)
}
