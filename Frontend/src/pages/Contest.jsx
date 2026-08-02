import { useState } from 'react'
import { Play, Pause, Square, RotateCcw } from 'lucide-react'
import { updateContestState } from '../api/contest'
import './Contest.css'

export default function Contest() {
  const [status, setStatus] = useState('pending')
  const [durationMins, setDurationMins] = useState(180)
  const [error, setError] = useState('')
  const [previewNow] = useState(() => Date.now())

  async function act(newStatus, extraFields = {}) {
    setError('')
    try {
      const data = await updateContestState({ status: newStatus, ...extraFields })
      setStatus(data.status)
    } catch (e) {
      setError(e.message)
    }
  }

  function handleStart() {
    const now = new Date()
    const end = new Date(now.getTime() + durationMins * 60 * 1000)
    act('active', { start_time: now.toISOString(), end_time: end.toISOString() })
  }

  return (
    <div className="contest-page">
      <h1>Contest Control</h1>

      <div className={`contest-status-bar status-${status}`}>
        {status.toUpperCase()}
      </div>

      {error && <div className="banner banner--error">{error}</div>}

      {status === 'pending' && (
        <div className="contest-card">
          <label className="duration-label">
            Contest Duration
            <div className="duration-input-row">
              <input
                type="number"
                min={1}
                value={durationMins}
                onChange={e => setDurationMins(Number(e.target.value))}
                className="duration-input"
              />
              <span className="duration-unit">minutes</span>
            </div>
            <span className="duration-hint">
              {Math.floor(durationMins / 60)}h {durationMins % 60}m — ends ~{
                new Date(previewNow + durationMins * 60000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
              } if started now
            </span>
          </label>
          <button className="action-btn btn-start" onClick={handleStart}>
            <Play size={16} /> Start Contest
          </button>
        </div>
      )}

      {status === 'active' && (
        <div className="contest-card action-row">
          <button className="action-btn btn-pause" onClick={() => act('paused')}>
            <Pause size={16} /> Pause
          </button>
          <button className="action-btn btn-end" onClick={() => act('ended')}>
            <Square size={16} /> End Contest
          </button>
        </div>
      )}

      {status === 'paused' && (
        <div className="contest-card action-row">
          <button className="action-btn btn-start" onClick={() => act('active')}>
            <Play size={16} /> Resume
          </button>
          <button className="action-btn btn-end" onClick={() => act('ended')}>
            <Square size={16} /> End Contest
          </button>
        </div>
      )}

      {status === 'ended' && (
        <div className="contest-card">
          <p className="ended-msg">Contest has ended.</p>
          <button className="action-btn btn-reset" onClick={() => act('pending')}>
            <RotateCcw size={16} /> Reset to Pending
          </button>
        </div>
      )}
    </div>
  )
}
