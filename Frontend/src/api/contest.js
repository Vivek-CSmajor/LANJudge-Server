import { apiFetch } from './client'

export async function fetchContestState() {
  const res = await apiFetch('/admin/contest')
  if (!res.ok) throw new Error(`Failed to load contest state (${res.status})`)
  return res.json()
}

export async function updateContestState(body) {
  const res = await apiFetch('/admin/contest', {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`Failed to update contest (${res.status})`)
  return res.json()
}
