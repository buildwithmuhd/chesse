import { Chess, type Color, type PieceSymbol, type Square } from 'chess.js'

export type BoardPiece = {
  square: Square
  label: string
  glyph: string
  color: Color
  type: PieceSymbol
}

const glyphs: Record<Color, Record<PieceSymbol, string>> = {
  w: { p: 'P', n: 'N', b: 'B', r: 'R', q: 'Q', k: 'K' },
  b: { p: 'p', n: 'n', b: 'b', r: 'r', q: 'q', k: 'k' },
}

const names: Record<PieceSymbol, string> = {
  p: 'pawn',
  n: 'knight',
  b: 'bishop',
  r: 'rook',
  q: 'queen',
  k: 'king',
}

export function boardPieces(chess: Chess): BoardPiece[] {
  return chess
    .board()
    .flatMap((rank, rankIndex) =>
      rank.map((piece, fileIndex) => {
        if (!piece) return null
        const square = `${'abcdefgh'[fileIndex]}${8 - rankIndex}` as Square
        return {
          square,
          label: `${piece.color === 'w' ? 'white' : 'black'} ${names[piece.type]} on ${square}`,
          glyph: glyphs[piece.color][piece.type],
          color: piece.color,
          type: piece.type,
        }
      }),
    )
    .filter((piece): piece is BoardPiece => piece !== null)
}

export function initialPieceCount(): number {
  return boardPieces(new Chess()).length
}
