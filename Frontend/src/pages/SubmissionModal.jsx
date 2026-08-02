import { useEffect, useState } from 'react'
import { X } from 'lucide-react'
import { fetchSubmissionDetail } from '../api/submissions'
import './SubmissionModal.css'

export default function SubmissionModal({ submissionId, onClose }) {
  const [detail, setDetail] = useState(null)
  const [error, setError] = useState(null)

  useEffect(() => {
    let cancelled = false
    fetchSubmissionDetail(submissionId)
      .then((data) => { if (!cancelled) setDetail(data) })
      .catch((e) => { if (!cancelled) setError(e.message) })
    return () => { cancelled = true }
  }, [submissionId])

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-panel" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Submission #{submissionId}</h2>
          <button className="modal-close" onClick={onClose} aria-label="Close">
            <X size={18} />
          </button>
        </div>

        {error && <div className="banner banner--error">{error}</div>}
        {!error && !detail && <div className="empty-state">Loading…</div>}

        {detail && (
          <div className="modal-body">
            <div className="modal-meta">
              <span>Team {detail.team_id}</span>
              <span>Problem {detail.problem_id}</span>
              <span className={`verdict verdict--${detail.verdict || 'pending'}`}>
                {detail.verdict || detail.status || 'Pending'}
              </span>
            </div>

            <pre className="code-block"><code>{detail.code}</code></pre>

            {detail.compile_log && (
              <>
                <h3>Compile Log</h3>
                <pre className="code-block">{detail.compile_log}</pre>
              </>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
