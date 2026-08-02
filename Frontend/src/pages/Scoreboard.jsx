import { Wifi, WifiOff } from 'lucide-react'
import { useAdminScoreboard } from '../hooks/useAdminScoreboard'
import './Scoreboard.css'

export default function Scoreboard() {
  const { rows, connected } = useAdminScoreboard()

  return (
    <div className="scoreboard-page">
      <header className="scoreboard-header">
        <h1>Scoreboard</h1>
        <div className={`status-pill ${connected ? 'status-pill--live' : 'status-pill--offline'}`}>
          {connected ? <Wifi size={13} /> : <WifiOff size={13} />}
          {connected ? 'Live' : 'Disconnected'}
        </div>
      </header>

      {rows.length === 0 ? (
        <div className="empty-state">
          {connected ? 'No submissions yet.' : 'Connecting…'}
        </div>
      ) : (
        <table className="scoreboard-table">
          <thead>
            <tr>
              <th>#</th>
              <th>Team</th>
              <th>Solved</th>
              <th>Penalty</th>
              <th>Last AC</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row, i) => (
              <tr key={row.team_id}>
                <td>{i + 1}</td>
                <td>{row.team_name}</td>
                <td>{row.total_solved}</td>
                <td>{row.total_penalty ?? '—'}</td>
                <td>{row.last_ac_time ? new Date(row.last_ac_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
