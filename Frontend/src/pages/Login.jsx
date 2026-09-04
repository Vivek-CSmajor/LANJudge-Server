import { useState } from 'react'
import { ShieldCheck, User, KeyRound, AlertCircle, Loader2 } from 'lucide-react'
import { adminLogin } from '../api/auth'
import './Login.css'

export default function Login({ onSuccess }) {
  const [username, setUsername] = useState('')
  const [loginCode, setLoginCode] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    if (!username.trim() || !loginCode.trim()) {
      setError('Please fill in both fields.')
      return
    }
    setError('')
    setLoading(true)
    try {
      const data = await adminLogin(username.trim(), loginCode.trim())
      localStorage.setItem('session_token', data.session_token)
      onSuccess()
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page">
      <form className="login-card" onSubmit={handleSubmit}>
        <div className="login-icon"><ShieldCheck size={22} /></div>
        <h1>Admin Login</h1>
        <p className="login-subtitle">Sign in to manage the contest.</p>

        {error && (
          <div className="banner banner--error login-error">
            <AlertCircle size={15} strokeWidth={2} />
            <span>{error}</span>
          </div>
        )}

        <label className="login-field">
          Username
          <div className="login-input-wrap">
            <User size={16} className="login-input-icon" />
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="admin"
              autoFocus
            />
          </div>
        </label>

        <label className="login-field">
          Login Code
          <div className="login-input-wrap">
            <KeyRound size={16} className="login-input-icon" />
            <input
              type="password"
              value={loginCode}
              onChange={(e) => setLoginCode(e.target.value)}
              placeholder="Enter your login code"
            />
          </div>
        </label>

        <button type="submit" className="login-submit-btn" disabled={loading}>
          {loading ? <><Loader2 size={15} className="spin" /> Signing in…</> : 'Sign in'}
        </button>
      </form>
    </div>
  )
}
