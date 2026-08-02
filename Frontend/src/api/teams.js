import { apiFetch } from './client'

export async function fetchTeams() {
  const res = await apiFetch('/admin/teams')
  if (!res.ok) throw new Error(`Failed to load teams (${res.status})`)
  return res.json()
}

export async function registerTeam(name, members) {
  const res = await apiFetch('/register', {
    method: 'POST',
    body: JSON.stringify({ name, members }),
  })
  if (!res.ok) throw new Error(`Failed to register team (${res.status})`)
  return res.json() // { name, login_code }
}

export async function setTeamStatus(id, status) {
  const res = await apiFetch(`/admin/teams/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  })
  if (!res.ok) throw new Error(`Failed to update team #${id} (${res.status})`)
  return res.json()
}
