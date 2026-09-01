import { useEffect, useState } from 'react'
import { Play, Pause, Square, RotateCcw } from 'lucide-react'
import { fetchContestState, updateContestState } from '../api/contest'
import './Contest.css'

// GET /admin/contest now exists, so that's the real source of truth on
// mount — this localStorage copy is just a fast first paint (avoids a
// flash of "PENDING" while the fetch is in flight) and a fallback for the
// rare case the fetch itself fails (server hiccup, etc).
const STORAGE_KEY = 'cc_admin_contest_state'

function loadPersisted() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

function savePersisted(state) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
  } catch {
    // storage full/unavailable — not fatal, just means this mitigation is a no-op
  }
}

export default function Contest() {
  const persisted = loadPersisted()
  const [status, setStatus] = useState(persisted?.status ?? 'pending')
  const [startTime, setStartTime] = useState(persisted?.startTime ?? null)
  const [endTime, setEndTime] = useState(persisted?.endTime ?? null)
  const [durationMins, setDurationMins] = useState(180)
  const [error, setError] = useState('')
  const [syncError, setSyncError] = useState('')
  const [previewNow] = useState(() => Date.now())

  // Reconcile with the server's real state on mount — this is what actually
  // fixes "a different browser/refresh doesn't know the contest is already
  // running," not just the localStorage copy above.
  useEffect(() => {
    fetchContestState()
      .then((data) => {
        const next = {
          status: data.status ?? 'pending',
          startTime: data.start_time ?? null,
          endTime: data.end_time ?? null,
        }
        setStatus(next.status)
        setStartTime(next.startTime)
        setEndTime(next.endTime)
        savePersisted(next)
      })
      .catch((e) => {
        // Couldn't confirm real state — keep whatever localStorage had
        // (already in state from the initial useState above) and say so,
        // rather than silently trusting a possibly-stale local copy.
        setSyncError(`Couldn't confirm live contest state (${e.message}) — showing the last state this browser saw.`)
      })
  }, [])

  // The backend's PATCH /admin/contest overwrites start_time/end_time
  // unconditionally, even on a status-only change — so every call here
  // re-sends the last-known-good timing instead of leaving it out and
  // letting the server zero it out from under every connected participant.
  async function act(newStatus, overrides = {}) {
    setError('')
    const now = new Date()
    const body = {
      status: newStatus,
      // Guard against the (should-be-impossible) case of acting on a
      // paused/active contest with no timing on hand yet — send something
      // sane rather than letting `undefined` reach JSON.stringify as a
      // dropped key, which is exactly the bug this function exists to avoid.
      start_time: overrides.start_time ?? startTime ?? now.toISOString(),
      end_time: overrides.end_time ?? endTime ?? new Date(now.getTime() + durationMins * 60000).toISOString(),
    }
    try {
      const data = await updateContestState(body)
      const next = {
        status: data.status ?? newStatus,
        startTime: data.start_time ?? body.start_time,
        endTime: data.end_time ?? body.end_time,
      }
      setStatus(next.status)
      setStartTime(next.startTime)
      setEndTime(next.endTime)
      savePersisted(next)
    } catch (e) {
      setError(e.message)
    }
  }

  function handleStart() {
    const now = new Date()
    const end = new Date(now.getTime() + durationMins * 60 * 1000)
    act('active', { start_time: now.toISOString(), end_time: end.toISOString() })
  }

  // Reset is the one case where wiping start/end is actually correct — a
  // "pending" contest with stale leftover timing is more confusing than a
  // clean slate, so this deliberately omits them rather than resending them.
  async function handleReset() {
    setError('')
    try {
      const data = await updateContestState({ status: 'pending' })
      setStatus(data.status ?? 'pending')
      setStartTime(null)
      setEndTime(null)
      savePersisted({ status: data.status ?? 'pending', startTime: null, endTime: null })
    } catch (e) {
      setError(e.message)
    }
  }

  return (
    <div className="contest-page">
      <h1>Contest Control</h1>

      <div className={`contest-status-bar status-${status}`}>
        {status.toUpperCase()}
      </div>

      {error && <div className="banner banner--error">{error}</div>}
      {syncError && <div className="banner banner--warning">{syncError}</div>}

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
          <button className="action-btn btn-reset" onClick={handleReset}>
            <RotateCcw size={16} /> Reset to Pending
          </button>
        </div>
      )}
    </div>
  )
}
