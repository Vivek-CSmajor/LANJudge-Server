const SERVER_URL = import.meta.env.VITE_SERVER_URL ?? 'http://localhost:8080'

export function apiFetch(path, options = {}) {
  const token = localStorage.getItem('session_token') ?? ''
  return fetch(`${SERVER_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      // RequireAdminAuth reads Authorization, not X-Session-Token — send
      // both so this keeps working regardless of which the backend checks.
      'Authorization': `Bearer ${token}`,
      'X-Session-Token': token,
      ...options.headers,
    },
  }).then((res) => {
    if (res.status === 401) {
      window.dispatchEvent(new Event('auth:unauthorized'))
    }
    return res
  })
}
