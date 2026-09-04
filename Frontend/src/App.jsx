import { useEffect, useState } from 'react'
import { LogOut } from 'lucide-react'
import Login from './pages/Login'
import Submissions from './pages/Submissions'
import Teams from './pages/Teams'
import Scoreboard from './pages/Scoreboard'
import Contest from './pages/Contest'
import Overrides from './pages/Overrides'
import './App.css'

const TABS = [
  { id: 'submissions', label: 'Submissions', Component: Submissions },
  { id: 'teams', label: 'Teams', Component: Teams },
  { id: 'scoreboard', label: 'Scoreboard', Component: Scoreboard },
  { id: 'contest', label: 'Contest', Component: Contest },
  { id: 'overrides', label: 'Overrides', Component: Overrides },
]

export default function App() {
  const [tab, setTab] = useState('submissions')
  const [authed, setAuthed] = useState(() => !!localStorage.getItem('session_token'))
  const ActivePage = TABS.find((t) => t.id === tab).Component

  function handleLogout() {
    localStorage.removeItem('session_token')
    setAuthed(false)
  }

  useEffect(() => {
    window.addEventListener('auth:unauthorized', handleLogout)
    return () => window.removeEventListener('auth:unauthorized', handleLogout)
  }, [])

  if (!authed) {
    return <Login onSuccess={() => setAuthed(true)} />
  }

  return (
    <div className="app-shell">
      <nav className="app-nav">
        {TABS.map((t) => (
          <button
            key={t.id}
            className={`app-nav-btn ${tab === t.id ? 'app-nav-btn--active' : ''}`}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
        <button className="app-nav-btn app-nav-btn--logout" onClick={handleLogout}>
          <LogOut size={14} /> Logout
        </button>
      </nav>

      <ActivePage />
    </div>
  )
}
