import { useEffect, useState } from 'react'

const SERVER_WS = (import.meta.env.VITE_SERVER_URL ?? 'http://localhost:8080')
  .replace(/^http/, 'ws')

export function useAdminScoreboard() {
  const [rows, setRows] = useState([])
  const [connected, setConnected] = useState(false)

  useEffect(() => {
    // A native WebSocket can't set an Authorization header, so this is now
    // gated by RequireAdminAuth's query-param fallback (?token=) instead —
    // the only way a browser can actually hand it credentials.
    const token = localStorage.getItem('session_token') ?? ''
    const ws = new WebSocket(`${SERVER_WS}/admin/scoreboard?token=${encodeURIComponent(token)}`)

    ws.onopen = () => setConnected(true)
    ws.onclose = () => setConnected(false)
    ws.onerror = () => setConnected(false)

    ws.onmessage = (e) => {
      let msg
      try { msg = JSON.parse(e.data) } catch { return }

      if (Array.isArray(msg)) {
        setRows(msg)
      } else if (msg.type === 'scoreboard') {
        setRows(msg.rows ?? [])
      }
    }

    return () => ws.close()
  }, [])

  return { rows, connected }
}
