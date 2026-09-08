import { useEffect, useRef, useState } from 'react'
import { Play, Pause, Square, RotateCcw } from 'lucide-react'
import { fetchContestState, updateContestState } from '../api/contest'
import './Contest.css'

// GET /admin/contest is the real source of truth — this localStorage copy is
// just a fast first paint (avoids a flash of "PENDING" while the first poll
// is in flight) and a fallback for the rare case a poll itself fails.
const STORAGE_KEY = 'cc_admin_contest_state'
const POLL_INTERVAL_MS = 5000

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

const STATUS_LABEL = { active: 'live', paused: 'paused', ended: 'ended', pending: 'pending' }

export default function Contest() {
  const persisted = loadPersisted()
  const [status, setStatus] = useState(persisted?.status ?? 'pending')
  const [durationMins, setDurationMins] = useState(180)
  const [error, setError] = useState('')
  const [syncError, setSyncError] = useState('')
  const [successMsg, setSuccessMsg] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [previewNow] = useState(() => Date.now())
  const pollingPausedRef = useRef(false) // true while an action is in flight, so its own poll tick can't race the response

  // start_time/end_time aren't rendered anywhere on this screen and, since
  // every action now sends only the fields it actually intends to change
  // (see `act` below), nothing here needs to hold onto them between renders
  // either — status is the only piece of server state this page tracks.
  function applyServerState(data, fallbackStatus) {
    const next = { status: data.status ?? fallbackStatus ?? 'pending' }
    setStatus(next.status)
    savePersisted(next)
    return next
  }

  // Reconciles with the server's real state on an interval — not just once on
  // mount — so this screen can't silently keep showing a stale status while
  // the contest actually changed via another session, device, or auto-action.
  useEffect(() => {
    let cancelled = false

    function poll() {
      if (pollingPausedRef.current) return
      fetchContestState()
        .then((data) => {
          if (cancelled) return
          setSyncError('')
          applyServerState(data)
        })
        .catch((e) => {
          if (cancelled) return
          // Couldn't confirm real state this tick — keep showing the last
          // confirmed state rather than blanking the screen, but say so.
          setSyncError(`Couldn't confirm live contest state (${e.message}) — showing the last state this browser saw.`)
        })
    }

    poll()
    const id = setInterval(poll, POLL_INTERVAL_MS)
    return () => { cancelled = true; clearInterval(id) }
  }, [])

  // Only ever sends what this action actually intends to change. Earlier
  // this always resent this screen's local start_time/end_time on every
  // action (a workaround for the backend once clobbering those fields on
  // any PATCH) — now that /admin/contest preserves omitted fields, resending
  // them here only risked overwriting the real value with stale local state
  // (e.g. rewriting the recorded start time to "now" from a pause/end click
  // if this browser's copy was ever out of date). Only Start supplies times,
  // because Start is the one action actually setting them.
  async function act(newStatus, overrides = {}, successText) {
    setError('')
    setSuccessMsg('')
    setSubmitting(true)
    pollingPausedRef.current = true
    try {
      const data = await updateContestState({ status: newStatus, ...overrides })
      applyServerState(data, newStatus)
      setSuccessMsg(successText ?? `Contest is now ${STATUS_LABEL[data.status ?? newStatus]}.`)
    } catch (e) {
      setError(e.message)
    } finally {
      setSubmitting(false)
      pollingPausedRef.current = false
    }
  }

  function handleStart() {
    const now = new Date()
    const end = new Date(now.getTime() + durationMins * 60 * 1000)
    act('active', { start_time: now.toISOString(), end_time: end.toISOString() }, 'Contest started.')
  }

  // NEEDS BACKEND: /admin/contest's start_time/end_time are *time.Time —
  // omitting a field and sending it as explicit `null` both decode to a nil
  // pointer, so there's currently no way for this PATCH to actually clear a
  // previously-set start_time/end_time back to empty. Reset only changes
  // status; the old times remain stored (and will be overwritten for real
  // the next time Start runs) until that's a real endpoint capability.
  async function handleReset() {
    await act('pending', {}, 'Contest reset to pending.')
  }

  return (
    <div className="contest-page">
      <h1>Contest Control</h1>

      <div className={`contest-status-bar status-${status}`}>
        {status.toUpperCase()}
      </div>

      {error && <div className="banner banner--error">{error}</div>}
      {syncError && <div className="banner banner--warning">{syncError}</div>}
      {successMsg && <div className="banner banner--success">{successMsg}</div>}

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
                disabled={submitting}
              />
              <span className="duration-unit">minutes</span>
            </div>
            <span className="duration-hint">
              {Math.floor(durationMins / 60)}h {durationMins % 60}m — ends ~{
                new Date(previewNow + durationMins * 60000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
              } if started now
            </span>
          </label>
          <button className="action-btn btn-start" onClick={handleStart} disabled={submitting}>
            <Play size={16} /> {submitting ? 'Starting…' : 'Start Contest'}
          </button>
        </div>
      )}

      {status === 'active' && (
        <div className="contest-card action-row">
          <button className="action-btn btn-pause" onClick={() => act('paused', {}, 'Contest paused.')} disabled={submitting}>
            <Pause size={16} /> Pause
          </button>
          <button className="action-btn btn-end" onClick={() => act('ended', {}, 'Contest ended.')} disabled={submitting}>
            <Square size={16} /> End Contest
          </button>
        </div>
      )}

      {status === 'paused' && (
        <div className="contest-card action-row">
          <button className="action-btn btn-start" onClick={() => act('active', {}, 'Contest resumed.')} disabled={submitting}>
            <Play size={16} /> Resume
          </button>
          <button className="action-btn btn-end" onClick={() => act('ended', {}, 'Contest ended.')} disabled={submitting}>
            <Square size={16} /> End Contest
          </button>
        </div>
      )}

      {status === 'ended' && (
        <div className="contest-card">
          <p className="ended-msg">Contest has ended.</p>
          <button className="action-btn btn-reset" onClick={handleReset} disabled={submitting}>
            <RotateCcw size={16} /> Reset to Pending
          </button>
        </div>
      )}
    </div>
  )
}
