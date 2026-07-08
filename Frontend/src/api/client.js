const SERVER_URL = import.meta.env.VITE_SERVER_URL ?? 'http://localhost:8080'

export function apiFetch(path, options = {}) {
  return fetch(`${SERVER_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'X-Session-Token': localStorage.getItem('session_token') ?? '',
      ...options.headers,
    },
  })
}
