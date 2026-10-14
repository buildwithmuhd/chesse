import assert from 'node:assert/strict'
import test from 'node:test'
import { Chess } from 'chess.js'

function boardPieces(chess) {
  return chess
    .board()
    .flatMap((rank, rankIndex) =>
      rank.map((piece, fileIndex) => {
        if (!piece) return null
        const square = `${'abcdefgh'[fileIndex]}${8 - rankIndex}`
        return { square, color: piece.color, type: piece.type }
      }),
    )
    .filter(Boolean)
}

test('pieces render from the initial chess position', () => {
  const pieces = boardPieces(new Chess())
  assert.equal(pieces.length, 32)
  assert.deepEqual(pieces.find((piece) => piece.square === 'g1'), { square: 'g1', color: 'w', type: 'n' })
})

test('pieces update after moves and undo', () => {
  const chess = new Chess()
  chess.move({ from: 'e2', to: 'e4' })
  assert.equal(boardPieces(chess).some((piece) => piece.square === 'e4' && piece.type === 'p'), true)
  chess.undo()
  assert.equal(boardPieces(chess).some((piece) => piece.square === 'e2' && piece.type === 'p'), true)
})
