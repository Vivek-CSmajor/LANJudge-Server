import { apiFetch } from './client'

export async function fetchOverrides() {
  const res = await apiFetch('/admin/overrides')
  if (!res.ok) throw new Error(`Failed to load overrides (${res.status})`)
  return res.json()
}

export async function respondToOverride(id, action) {
  const res = await apiFetch(`/admin/overrides/${id}`, {
    method: 'POST',
    body: JSON.stringify({ action }),
  })
  if (!res.ok) throw new Error(`Failed to ${action} override #${id} (${res.status})`)
  return res.json()
}
