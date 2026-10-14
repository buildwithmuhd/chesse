const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? `${window.location.protocol}//${window.location.hostname}:8080`

export type SessionUser = {
  id: number
  email: string
}

export type SessionResponse = {
  accessToken: string
  user: SessionUser
}

export async function apiRequest(path: string, init: RequestInit = {}): Promise<SessionResponse> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...init.headers,
    },
  })

  if (!response.ok) {
    const message = await readError(response)
    throw new Error(message)
  }

  if (response.status === 204) {
    return { accessToken: '', user: { id: 0, email: '' } }
  }
  return (await response.json()) as SessionResponse
}

async function readError(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { error?: string }
    return body.error ?? `Request failed with ${response.status}`
  } catch {
    return `Request failed with ${response.status}`
  }
}
