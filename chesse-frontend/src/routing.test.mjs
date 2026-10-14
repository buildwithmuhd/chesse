import assert from 'node:assert/strict'
import test from 'node:test'

const protectedNames = new Set(['game', 'challenge', 'spectate', 'tournaments', 'history'])

function parseRoute(pathname, search = '') {
  const returnTo = new URLSearchParams(search).get('returnTo') ?? '/'
  if (pathname === '/') return { name: 'home', path: pathname }
  if (pathname === '/play/ai') return { name: 'ai', path: pathname }
  if (pathname.startsWith('/play/')) return { name: 'game', path: pathname, gameId: pathname.slice('/play/'.length) }
  if (pathname === '/challenge') return { name: 'challenge', path: pathname }
  if (pathname === '/spectate') return { name: 'spectate', path: pathname }
  if (pathname === '/tournaments') return { name: 'tournaments', path: pathname }
  if (pathname.startsWith('/tournaments/')) return { name: 'tournaments', path: pathname, tournamentId: pathname.slice('/tournaments/'.length) }
  if (pathname === '/history') return { name: 'history', path: pathname }
  if (pathname.startsWith('/history/')) return { name: 'history', path: pathname, gameId: pathname.slice('/history/'.length) }
  if (pathname === '/sign-in') return { name: 'sign-in', path: pathname, returnTo }
  if (pathname === '/register') return { name: 'register', path: pathname, returnTo }
  return { name: 'not-found', path: pathname }
}

function needsAuth(route) {
  return protectedNames.has(route.name)
}

function signInPath(returnTo) {
  return `/sign-in?returnTo=${encodeURIComponent(returnTo)}`
}

function safeReturnTo(value) {
  return value.startsWith('/') && !value.startsWith('//') ? value : '/'
}

test('route guards protect multiplayer and history but not AI', () => {
  assert.equal(needsAuth(parseRoute('/play/ai')), false)
  assert.equal(needsAuth(parseRoute('/play/game-1')), true)
  assert.equal(needsAuth(parseRoute('/challenge')), true)
  assert.equal(needsAuth(parseRoute('/history')), true)
})

test('returnTo redirects are encoded and local-only', () => {
  assert.equal(signInPath('/challenge'), '/sign-in?returnTo=%2Fchallenge')
  assert.equal(safeReturnTo('/play/game-1'), '/play/game-1')
  assert.equal(safeReturnTo('//evil.example'), '/')
})
