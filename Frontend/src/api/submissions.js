import { apiFetch } from './client'

export async function fetchSubmissions() {
  const res = await apiFetch('/admin/submissions')
  if (!res.ok) throw new Error(`Failed to load submissions (${res.status})`)
  return res.json()
}

export async function fetchSubmissionDetail(id) {
  const res = await apiFetch(`/admin/submissions/${id}`)
  if (!res.ok) throw new Error(`Failed to load submission #${id} (${res.status})`)
  return res.json()
}

export async function rejudgeSubmission(id) {
  const res = await apiFetch(`/admin/submissions/${id}/rejudge`, { method: 'POST' })
  if (!res.ok) throw new Error(`Failed to rejudge submission #${id} (${res.status})`)
  return res.json()
}
