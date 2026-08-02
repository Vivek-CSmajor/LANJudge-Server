import { useCallback, useEffect, useState } from 'react'
import { RefreshCw } from 'lucide-react'
import { fetchSubmissions } from '../api/submissions'
import { fetchTeams } from '../api/teams'
import SubmissionModal from './SubmissionModal'
import './Submissions.css'

const VERDICT_LABEL = {
  AC: 'Accepted',
  WA: 'Wrong Answer',
  TLE: 'Time Limit',
  RE: 'Runtime Error',
  CE: 'Compile Error',
  IE: 'Internal Error',
}

function fmtTime(iso) {
  const d = new Date(iso)
  if (!iso || Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString([], { dateStyle: 'short', timeStyle: 'medium' })
}

export default function Submissions() {
  const [submissions, setSubmissions] = useState([])
  const [teamsById, setTeamsById] = useState({})
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [selectedId, setSelectedId] = useState(null)

  const load = useCallback(() => {
    return fetchSubmissions()
      .then((data) => {
        setSubmissions(Array.isArray(data) ? data : [])
        setError(null)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    fetchTeams()
      .then((teams) => {
        const map = {}
        for (const t of teams ?? []) map[t.id] = t.name
        setTeamsById(map)
      })
      .catch(() => {}) // team names are best-effort; falls back to "Team {id}"
  }, [])

  useEffect(() => {
    load()
  }, [load])

  function handleRefresh() {
    setRefreshing(true)
    load().finally(() => setRefreshing(false))
  }

  const rows = [...submissions].sort((a, b) => b.id - a.id)

  return (
    <div className="submissions-page">
      <header className="submissions-header">
        <h1>Submissions</h1>
        <button className="refresh-btn" onClick={handleRefresh} disabled={refreshing}>
          <RefreshCw size={14} className={refreshing ? 'spin' : ''} />
          {refreshing ? 'Refreshing…' : 'Refresh'}
        </button>
      </header>

      {error && <div className="banner banner--error">{error}</div>}

      {loading ? (
        <div className="empty-state">Loading submissions…</div>
      ) : rows.length === 0 ? (
        <div className="empty-state">No submissions yet.</div>
      ) : (
        <table className="submissions-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Team</th>
              <th>Problem</th>
              <th>Language</th>
              <th>Verdict</th>
              <th>Submitted</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((s) => (
              <tr key={s.id} className="submissions-row" onClick={() => setSelectedId(s.id)}>
                <td>#{s.id}</td>
                <td>{teamsById[s.team_id] ?? `Team ${s.team_id}`}</td>
                <td>{s.problem_id}</td>
                <td>C++</td>
                <td>
                  <span className={`verdict verdict--${s.verdict || 'pending'}`}>
                    {VERDICT_LABEL[s.verdict] ?? s.status ?? 'Pending'}
                  </span>
                </td>
                <td>{fmtTime(s.submitted_at)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {selectedId != null && (
        <SubmissionModal key={selectedId} submissionId={selectedId} onClose={() => setSelectedId(null)} />
      )}
    </div>
  )
}
