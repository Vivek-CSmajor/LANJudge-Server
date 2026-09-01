import { useCallback, useEffect, useState } from 'react'
import { RefreshCw, Check, X } from 'lucide-react'
import { fetchOverrides, respondToOverride } from '../api/overrides'
import { fetchTeams } from '../api/teams'
import './Overrides.css'

function fmtTime(iso) {
  const d = new Date(iso)
  if (!iso || Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString([], { dateStyle: 'short', timeStyle: 'medium' })
}

export default function Overrides() {
  const [overrides, setOverrides] = useState([])
  const [teamsById, setTeamsById] = useState({})
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [actingId, setActingId] = useState(null)

  const load = useCallback(() => {
    // Team names are refetched alongside overrides every time (not just once
    // on mount) so a team registered after this page was opened still shows
    // its real name instead of "Team {id}" until the next full remount.
    const teamsPromise = fetchTeams()
      .then((teams) => {
        const map = {}
        for (const t of teams ?? []) map[t.id] = t.name
        setTeamsById(map)
      })
      .catch(() => {}) // team names are best-effort; falls back to "Team {id}"

    const overridesPromise = fetchOverrides()
      .then((data) => {
        setOverrides(Array.isArray(data) ? data : [])
        setError(null)
      })
      .catch((e) => setError(e.message))

    return Promise.all([teamsPromise, overridesPromise]).finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    load()
  }, [load])

  function handleRefresh() {
    setRefreshing(true)
    load().finally(() => setRefreshing(false))
  }

  function handleRespond(id, action) {
    setActingId(id)
    respondToOverride(id, action)
      .then(() => load())
      .catch((e) => setError(e.message))
      .finally(() => setActingId(null))
  }

  const rows = [...overrides].sort((a, b) => b.id - a.id)

  return (
    <div className="overrides-page">
      <header className="overrides-header">
        <h1>Device Overrides</h1>
        <button className="refresh-btn" onClick={handleRefresh} disabled={refreshing}>
          <RefreshCw size={14} className={refreshing ? 'spin' : ''} />
          {refreshing ? 'Refreshing…' : 'Refresh'}
        </button>
      </header>

      {error && <div className="banner banner--error">{error}</div>}

      {loading ? (
        <div className="empty-state">Loading overrides…</div>
      ) : rows.length === 0 ? (
        <div className="empty-state">No override requests.</div>
      ) : (
        <table className="overrides-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Team</th>
              <th>New Device ID</th>
              <th>Status</th>
              <th>Requested</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {rows.map((o) => (
              <tr key={o.id}>
                <td>#{o.id}</td>
                <td>{teamsById[o.team_id] ?? `Team ${o.team_id}`}</td>
                <td className="device-id">{o.new_device_id}</td>
                <td>
                  <span className={`override-status override-status--${o.status}`}>{o.status}</span>
                </td>
                <td>{fmtTime(o.created_at)}</td>
                <td className="override-actions">
                  {o.status === 'pending' ? (
                    <>
                      <button
                        className="icon-btn icon-btn--approve"
                        disabled={actingId === o.id}
                        onClick={() => handleRespond(o.id, 'approve')}
                      >
                        <Check size={14} /> Approve
                      </button>
                      <button
                        className="icon-btn icon-btn--reject"
                        disabled={actingId === o.id}
                        onClick={() => handleRespond(o.id, 'reject')}
                      >
                        <X size={14} /> Reject
                      </button>
                    </>
                  ) : (
                    '—'
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
