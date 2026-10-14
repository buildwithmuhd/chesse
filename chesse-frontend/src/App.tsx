import { Chess, type Move, type Square } from 'chess.js'
import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { apiRequest, type SessionUser } from './api'
import { boardPieces } from './chessView'
import { needsAuth, parseRoute, safeReturnTo, signInPath, type Route } from './routing'
import './App.css'

type AuthMode = 'sign-in' | 'register'
type PendingGate = { title: string; message: string; path: string } | null

const navItems = [
  { label: 'Challenge', path: '/challenge', message: 'Sign in to challenge players.' },
  { label: 'Spectate', path: '/spectate', message: 'Sign in to watch live games.' },
  { label: 'Tournaments', path: '/tournaments', message: 'Sign in to host or join tournaments.' },
  { label: 'History', path: '/history', message: 'Sign in to save and review games.' },
]

function App() {
  const [route, setRoute] = useState(() => parseRoute(window.location.pathname, window.location.search))
  const [user, setUser] = useState<SessionUser | null>(null)
  const [accessToken, setAccessToken] = useState('')
  const [status, setStatus] = useState('Checking session...')
  const [busy, setBusy] = useState(true)
  const [gate, setGate] = useState<PendingGate>(null)

  useEffect(() => {
    const onPop = () => setRoute(parseRoute(window.location.pathname, window.location.search))
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  }, [])

  useEffect(() => {
    apiRequest('/auth/refresh', { method: 'POST' })
      .then((session) => {
        setUser(session.user)
        setAccessToken(session.accessToken)
        setStatus(`Signed in as ${session.user.email}`)
      })
      .catch(() => setStatus('Guest mode'))
      .finally(() => setBusy(false))
  }, [])

  useEffect(() => {
    if (!busy && !user && needsAuth(route)) navigate(signInPath(route.path))
  }, [busy, route, user])

  function navigate(path: string) {
    window.history.pushState(null, '', path)
    setRoute(parseRoute(window.location.pathname, window.location.search))
    setGate(null)
  }

  async function logout() {
    setBusy(true)
    try {
      await apiRequest('/auth/logout', { method: 'POST' })
    } finally {
      setUser(null)
      setAccessToken('')
      setStatus('Guest mode')
      navigate('/')
      setBusy(false)
    }
  }

  async function completeAuth(mode: AuthMode, email: string, password: string, returnTo: string) {
    setBusy(true)
    setStatus(mode === 'sign-in' ? 'Signing in...' : 'Creating account...')
    try {
      const session = await apiRequest(mode === 'sign-in' ? '/auth/login' : '/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })
      setUser(session.user)
      setAccessToken(session.accessToken)
      setStatus(`Signed in as ${session.user.email}`)
      navigate(safeReturnTo(returnTo))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className={route.name === 'ai' || route.name === 'game' ? 'app app-game' : 'app'}>
      <TopNav user={user} route={route} onNavigate={navigate} onLogout={logout} onGate={setGate} />
      {route.name === 'home' && <HomePage onNavigate={navigate} user={user} onGate={setGate} />}
      {route.name === 'ai' && <GamePage mode="ai" user={user} />}
      {route.name === 'game' && <GamePage mode="multiplayer" user={user} gameId={route.gameId} />}
      {route.name === 'challenge' && <Placeholder title="Challenge" text="Player search and challenges connect here." />}
      {route.name === 'spectate' && <Placeholder title="Spectate" text="Live games will appear here." empty="No live games yet." />}
      {route.name === 'tournaments' && <Placeholder title="Tournaments" text="Create, join, and track tournaments here." />}
      {route.name === 'history' && <Placeholder title="History" text="Saved games will appear after authenticated play." empty="No history yet." />}
      {route.name === 'sign-in' && <AuthPage mode="sign-in" busy={busy} returnTo={route.returnTo} onSubmit={completeAuth} onNavigate={navigate} />}
      {route.name === 'register' && <AuthPage mode="register" busy={busy} returnTo={route.returnTo} onSubmit={completeAuth} onNavigate={navigate} />}
      {route.name === 'not-found' && <Placeholder title="Not found" text="That room does not exist." />}
      {gate ? <GateDialog gate={gate} onClose={() => setGate(null)} onNavigate={navigate} /> : null}
      <span className="sr-only" aria-live="polite">
        {status}
        {accessToken ? ', authenticated' : ''}
      </span>
    </div>
  )
}

function TopNav({
  user,
  route,
  onNavigate,
  onLogout,
  onGate,
}: {
  user: SessionUser | null
  route: Route
  onNavigate: (path: string) => void
  onLogout: () => void
  onGate: (gate: PendingGate) => void
}) {
  return (
    <header className="top-nav">
      <button className="brand" type="button" onClick={() => onNavigate('/')}>
        <span className="brand-mark">N</span>
        <span>NoChess</span>
      </button>
      <nav aria-label="Primary navigation">
        <button aria-current={route.name === 'ai' ? 'page' : undefined} type="button" onClick={() => onNavigate('/play/ai')}>
          Play AI
        </button>
        {navItems.map((item) => (
          <button
            key={item.path}
            aria-current={route.path === item.path ? 'page' : undefined}
            type="button"
            onClick={() => (user ? onNavigate(item.path) : onGate({ title: item.label, message: item.message, path: item.path }))}
          >
            {!user ? <span aria-hidden="true">Lock </span> : null}
            {item.label}
          </button>
        ))}
      </nav>
      <div className="account-nav">
        {user ? (
          <>
            <button type="button" onClick={() => onNavigate('/history')}>{initials(user.email)}</button>
            <button type="button" onClick={onLogout}>Sign out</button>
          </>
        ) : (
          <>
            <button type="button" onClick={() => onNavigate('/sign-in')}>Sign in</button>
            <button className="nav-primary" type="button" onClick={() => onNavigate('/register')}>Register</button>
          </>
        )}
      </div>
    </header>
  )
}

function HomePage({ user, onNavigate, onGate }: { user: SessionUser | null; onNavigate: (path: string) => void; onGate: (gate: PendingGate) => void }) {
  return (
    <main className="home-page">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Browser chess arena</p>
          <h1>Choose your board.</h1>
          <p className="lede">Play a quick guest AI game, or sign in for challenges, spectating, tournaments, saved history, and voice.</p>
        </div>
        <div className="home-actions">
          <button className="primary-button" type="button" onClick={() => onNavigate('/play/ai')}>Play vs AI</button>
          <button className="secondary-button" type="button" onClick={() => (user ? onNavigate('/challenge') : onGate({ title: 'Challenge', message: 'Sign in to challenge players.', path: '/challenge' }))}>Challenge player</button>
        </div>
      </section>
      <section className="feature-grid" aria-label="Arena features">
        {navItems.map((item) => (
          <button key={item.path} type="button" className="feature-card" onClick={() => (user ? onNavigate(item.path) : onGate({ title: item.label, message: item.message, path: item.path }))}>
            <strong>{item.label}</strong>
            <span>{user ? 'Open' : 'Sign in required'}</span>
          </button>
        ))}
      </section>
    </main>
  )
}

function AuthPage({ mode, busy, returnTo, onSubmit, onNavigate }: { mode: AuthMode; busy: boolean; returnTo: string; onSubmit: (mode: AuthMode, email: string, password: string, returnTo: string) => Promise<void>; onNavigate: (path: string) => void }) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const isRegister = mode === 'register'

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    if (isRegister && password !== confirm) {
      setError('Passwords must match.')
      return
    }
    try {
      await onSubmit(mode, email, password, returnTo)
    } catch (authError) {
      setError(authError instanceof Error ? authError.message : 'Authentication failed')
    }
  }

  return (
    <main className="auth-page">
      <form className="auth-card" onSubmit={submit}>
        <p className="eyebrow">{isRegister ? 'New account' : 'Welcome back'}</p>
        <h1>{isRegister ? 'Register' : 'Sign in'}</h1>
        <label>Email<input autoComplete="email" inputMode="email" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required /></label>
        <label>
          Password
          <span className="password-row">
            <input autoComplete={isRegister ? 'new-password' : 'current-password'} minLength={8} type={showPassword ? 'text' : 'password'} value={password} onChange={(event) => setPassword(event.target.value)} required />
            <button type="button" onClick={() => setShowPassword((showing) => !showing)}>{showPassword ? 'Hide' : 'Show'}</button>
          </span>
        </label>
        {isRegister ? (
          <>
            <label>Confirm password<input autoComplete="new-password" minLength={8} type={showPassword ? 'text' : 'password'} value={confirm} onChange={(event) => setConfirm(event.target.value)} required /></label>
            <p className="hint">Use at least 8 characters. Longer passphrases are better.</p>
          </>
        ) : null}
        {error ? <p className="form-error">{error}</p> : null}
        <button className="primary-button" type="submit" disabled={busy}>{busy ? 'Working...' : isRegister ? 'Create account' : 'Sign in'}</button>
        <button className="link-button" type="button" onClick={() => onNavigate(isRegister ? `/sign-in?returnTo=${returnTo}` : `/register?returnTo=${returnTo}`)}>{isRegister ? 'Already have an account? Sign in' : 'Need an account? Register'}</button>
        <button className="link-button" type="button" onClick={() => onNavigate('/play/ai')}>Continue as guest and play vs AI</button>
      </form>
    </main>
  )
}

function GamePage({ mode, user, gameId }: { mode: 'ai' | 'multiplayer'; user: SessionUser | null; gameId?: string }) {
  const [chess, setChess] = useState(() => new Chess())
  const [selected, setSelected] = useState<Square | null>(null)
  const [orientation, setOrientation] = useState<'white' | 'black'>('white')
  const [lastMove, setLastMove] = useState<Move | null>(null)
  const legalMoves = useMemo(() => (selected ? chess.moves({ square: selected, verbose: true }) : []), [chess, selected])
  const pieces = useMemo(() => boardPieces(chess), [chess])
  const squares = boardSquares(orientation)
  const status = chess.isGameOver() ? (chess.isCheckmate() ? 'Checkmate' : 'Draw') : chess.inCheck() ? `${sideName(chess.turn())} to move, check` : `${sideName(chess.turn())} to move`

  function chooseSquare(square: Square) {
    const piece = chess.get(square)
    if (selected) {
      const move = chess.move({ from: selected, to: square, promotion: 'q' })
      if (move) {
        setLastMove(move)
        setSelected(null)
        setChess(new Chess(chess.fen()))
        return
      }
    }
    setSelected(piece && piece.color === chess.turn() ? square : null)
  }

  function undo() {
    chess.undo()
    setLastMove(null)
    setSelected(null)
    setChess(new Chess(chess.fen()))
  }

  return (
    <main className="game-page">
      <section className="player-column" aria-label="Players and clocks">
        <PlayerCard name={mode === 'ai' ? 'NoChess AI' : 'Opponent'} clock="05:00" captured="p n" />
        <PlayerCard name={user?.email ?? 'Guest'} clock="05:00" captured="P" />
      </section>
      <section className="board-area">
        <p className="game-notice">{mode === 'ai' && !user ? 'Guest game: not saved.' : `Game ${gameId ?? 'local'}`}</p>
        <div className="chessboard" role="grid" aria-label="Chess board">
          {squares.map((square) => {
            const piece = pieces.find((nextPiece) => nextPiece.square === square)
            const isLegal = legalMoves.some((move) => move.to === square)
            const isLast = lastMove?.from === square || lastMove?.to === square
            return (
              <button key={square} type="button" className={`square ${squareColor(square)}${selected === square ? ' selected' : ''}${isLegal ? ' legal' : ''}${isLast ? ' last' : ''}`} aria-label={piece?.label ?? `empty ${square}`} onClick={() => chooseSquare(square)}>
                <span className="coord file">{square[0]}</span>
                <span className="coord rank">{square[1]}</span>
                {piece ? <span className={`piece ${piece.color === 'w' ? 'white-piece' : 'black-piece'}`}>{piece.glyph}</span> : null}
              </button>
            )
          })}
        </div>
      </section>
      <aside className="game-panel">
        <p className="eyebrow">Status</p>
        <h2>{status}</h2>
        <div className="control-grid">
          <button type="button" onClick={undo} disabled={chess.history().length === 0}>Undo</button>
          <button type="button" onClick={() => setOrientation((side) => (side === 'white' ? 'black' : 'white'))}>Flip</button>
          <button type="button" onClick={() => setChess(new Chess())}>Rematch</button>
          <button type="button">Resign</button>
        </div>
        {mode === 'multiplayer' ? <div className="voice-controls"><button type="button">Mic</button><button type="button">Mute</button></div> : null}
        <ol className="move-list" aria-label="Move list">
          {chess.history().map((move, index) => <li key={`${move}-${index}`}>{move}</li>)}
        </ol>
      </aside>
    </main>
  )
}

function PlayerCard({ name, clock, captured }: { name: string; clock: string; captured: string }) {
  return <article className="player-card"><span className="avatar">{initials(name)}</span><div><strong>{name}</strong><span>{captured ? `Captured: ${captured}` : 'No captures'}</span></div><time>{clock}</time></article>
}

function Placeholder({ title, text, empty }: { title: string; text: string; empty?: string }) {
  return <main className="placeholder-page"><p className="eyebrow">{title}</p><h1>{text}</h1>{empty ? <p className="empty-state">{empty}</p> : null}</main>
}

function GateDialog({ gate, onClose, onNavigate }: { gate: Exclude<PendingGate, null>; onClose: () => void; onNavigate: (path: string) => void }) {
  return <div className="dialog-backdrop" role="presentation" onClick={onClose}><section className="gate-dialog" role="dialog" aria-modal="true" aria-labelledby="gate-title" onClick={(event) => event.stopPropagation()}><h2 id="gate-title">{gate.title}</h2><p>{gate.message}</p><div className="dialog-actions"><button className="primary-button" type="button" onClick={() => onNavigate(signInPath(gate.path))}>Sign in</button><button className="secondary-button" type="button" onClick={() => onNavigate(`/register?returnTo=${encodeURIComponent(gate.path)}`)}>Register</button></div></section></div>
}

function boardSquares(orientation: 'white' | 'black'): Square[] {
  const files = orientation === 'white' ? ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'] : ['h', 'g', 'f', 'e', 'd', 'c', 'b', 'a']
  const ranks = orientation === 'white' ? [8, 7, 6, 5, 4, 3, 2, 1] : [1, 2, 3, 4, 5, 6, 7, 8]
  return ranks.flatMap((rank) => files.map((file) => `${file}${rank}` as Square))
}

function squareColor(square: Square) {
  const file = square.charCodeAt(0) - 97
  const rank = Number(square[1])
  return (file + rank) % 2 === 0 ? ' dark' : ' light'
}

function sideName(side: 'w' | 'b') {
  return side === 'w' ? 'White' : 'Black'
}

function initials(value: string) {
  return value.slice(0, 2).toUpperCase()
}

export default App
