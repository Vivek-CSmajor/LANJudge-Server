import { useCallback, useEffect, useState } from 'react'
import { RefreshCw } from 'lucide-react'
import { fetchTeams, registerTeam, setTeamStatus } from '../api/teams'
import './Teams.css'

export default function Teams() {
  const [teams, setTeams] = useState([])
  const [loading, setLoading] = useState(true)
  const [listError, setListError] = useState(null)

  const [name, setName] = useState('')
  const [members, setMembers] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState(null)
  const [lastCreated, setLastCreated] = useState(null)

  const [togglingId, setTogglingId] = useState(null)

  const load = useCallback(() => {
    return fetchTeams()
      .then((data) => {
        setTeams(Array.isArray(data) ? data : [])
        setListError(null)
      })
      .catch((e) => setListError(e.message))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    load()
  }, [load])

  function handleRegister(e) {
    e.preventDefault()
    if (!name.trim()) return
    setSubmitting(true)
    setFormError(null)
    registerTeam(name.trim(), members.trim())
      .then((res) => {
        setLastCreated(res)
        setName('')
        setMembers('')
        return load()
      })
      .catch((e) => setFormError(e.message))
      .finally(() => setSubmitting(false))
  }

  function handleToggle(team) {
    const next = team.status === 'disabled' ? 'enabled' : 'disabled'
    setTogglingId(team.id)
    setTeamStatus(team.id, next)
      .then(() => load())
      .catch((e) => setListError(e.message))
      .finally(() => setTogglingId(null))
  }

  return (
    <div className="teams-page">
      <h1>Teams</h1>

      <form className="team-form" onSubmit={handleRegister}>
        <h2>Register New Team</h2>
        <div className="form-row">
          <label>
            Team Name
            <input value={name} onChange={(e) => setName(e.target.value)} required />
          </label>
          <label>
            Members
            <input
              value={members}
              onChange={(e) => setMembers(e.target.value)}
              placeholder="comma-separated names"
            />
          </label>
          <button type="submit" disabled={submitting}>
            {submitting ? 'Registering…' : 'Register Team'}
          </button>
        </div>
        {formError && <div className="banner banner--error">{formError}</div>}
        {lastCreated && (
          <div className="banner banner--success">
            Team "{lastCreated.name}" created — login code: <code>{lastCreated.login_code}</code>
          </div>
        )}
      </form>

      <div className="teams-list-header">
        <h2>All Teams</h2>
        <button className="refresh-btn" onClick={load}>
          <RefreshCw size={14} /> Refresh
        </button>
      </div>

      {listError && <div className="banner banner--error">{listError}</div>}

      {loading ? (
        <div className="empty-state">Loading teams…</div>
      ) : teams.length === 0 ? (
        <div className="empty-state">No teams yet.</div>
      ) : (
        <table className="teams-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {teams.map((t) => (
              <tr key={t.id}>
                <td>#{t.id}</td>
                <td>{t.name}</td>
                <td>
                  <span className={`team-status team-status--${t.status ?? 'enabled'}`}>
                    {t.status ?? 'enabled'}
                  </span>
                </td>
                <td>
                  <button
                    className="toggle-btn"
                    disabled={togglingId === t.id}
                    onClick={() => handleToggle(t)}
                  >
                    {togglingId === t.id ? '…' : t.status === 'disabled' ? 'Enable' : 'Disable'}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
