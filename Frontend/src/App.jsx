import { useState } from 'react'
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
  const ActivePage = TABS.find((t) => t.id === tab).Component

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
      </nav>

      <ActivePage />
    </div>
  )
}
