import { apiFetch } from './client'

export async function adminLogin(username, loginCode) {
  const res = await apiFetch('/admin/login', {
    method: 'POST',
    body: JSON.stringify({ user_name: username, login_code: loginCode }),
  })
  if (!res.ok) {
    if (res.status === 401) throw new Error('Invalid username or login code.')
    throw new Error(`Login failed (${res.status}).`)
  }
  return res.json() // { session_token }
}
