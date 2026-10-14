export type Route =
  | { name: 'home'; path: string }
  | { name: 'ai'; path: string }
  | { name: 'game'; path: string; gameId: string }
  | { name: 'challenge'; path: string }
  | { name: 'spectate'; path: string }
  | { name: 'tournaments'; path: string; tournamentId?: string }
  | { name: 'history'; path: string; gameId?: string }
  | { name: 'sign-in'; path: string; returnTo: string }
  | { name: 'register'; path: string; returnTo: string }
  | { name: 'not-found'; path: string }

export const protectedPaths = ['/play/', '/challenge', '/spectate', '/tournaments', '/history']

export function parseRoute(pathname: string, search = ''): Route {
  const returnTo = new URLSearchParams(search).get('returnTo') ?? '/'
  if (pathname === '/') return { name: 'home', path: pathname }
  if (pathname === '/play/ai') return { name: 'ai', path: pathname }
  if (pathname.startsWith('/play/')) return { name: 'game', path: pathname, gameId: pathname.slice('/play/'.length) }
  if (pathname === '/challenge') return { name: 'challenge', path: pathname }
  if (pathname === '/spectate') return { name: 'spectate', path: pathname }
  if (pathname === '/tournaments') return { name: 'tournaments', path: pathname }
  if (pathname.startsWith('/tournaments/')) {
    return { name: 'tournaments', path: pathname, tournamentId: pathname.slice('/tournaments/'.length) }
  }
  if (pathname === '/history') return { name: 'history', path: pathname }
  if (pathname.startsWith('/history/')) return { name: 'history', path: pathname, gameId: pathname.slice('/history/'.length) }
  if (pathname === '/sign-in') return { name: 'sign-in', path: pathname, returnTo }
  if (pathname === '/register') return { name: 'register', path: pathname, returnTo }
  return { name: 'not-found', path: pathname }
}

export function needsAuth(route: Route): boolean {
  return route.name === 'game' || route.name === 'challenge' || route.name === 'spectate' || route.name === 'tournaments' || route.name === 'history'
}

export function signInPath(returnTo: string): string {
  return `/sign-in?returnTo=${encodeURIComponent(returnTo)}`
}

export function safeReturnTo(value: string): string {
  return value.startsWith('/') && !value.startsWith('//') ? value : '/'
}
