# CampusCode — Developer Reference Documentation

> **Internal reference. Pre-release v0.1.0. Last updated May 2026.** This document covers everything: architecture, file structure, DB schema, auth flow, API endpoints, judge engine, build order, and design decisions. Refer here before asking questions.

---

## Table of Contents

1. [Product Overview](#1-product-overview)
2. [System Architecture](#2-system-architecture)
3. [Go Project File Structure](#3-go-project-file-structure)
4. [Database Schema](#4-database-schema)
5. [Authentication & Session System](#5-authentication--session-system)
6. [Full API Reference](#6-full-api-reference)
7. [Judge Engine](#7-judge-engine)
8. [Scoreboard](#8-scoreboard)
9. [Participant App (Electron)](#9-participant-app-electron)
10. [Build Order & Current Status](#10-build-order--current-status)
11. [Performance Targets & Security Model](#11-performance-targets--security-model)
12. [Key Design Decisions & Quick Reference](#12-key-design-decisions--quick-reference)

---

## 1. Product Overview

CampusCode is an **offline-capable, LAN-based secure coding contest and assessment platform** for universities and campus placement drives.

It is **not** a Codeforces clone. It is **not** an enterprise proctoring tool. It is a plug-and-play, single-executable deployment system built for simplicity, reliability, and full offline operation.

### Core Value Proposition

| Property              | Description                                                                                                        |
| --------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Offline First         | The entire contest — judging, scoreboard, monitoring — runs with zero internet. One admin laptop hosts everything. |
| Single Executable     | The organiser runs one binary. DB, judge workers, API server, scoreboard all spin up automatically.                |
| Integrated Monitoring | Admin sees all participants in real time: screenshots, focus switches, idle status, submission frequency.          |
| LAN Discovery         | Participants connect via auto-discovery or a short join code. No manual IP configuration needed.                   |
| Controlled Client     | Electron app enforces kiosk/fullscreen mode. No external browsing, no URL bar, focus switches logged.              |

### Four Bundled Components

|Component|Role|Tech|Status|
|---|---|---|---|
|Judge Engine|Compile, run, verdict|Go|✅ Core working|
|LAN API Server|REST endpoints over LAN|Go net/http|🔨 In progress|
|Admin Dashboard|Web UI, real-time monitoring|React|📋 Planned|
|Participant App|Kiosk client, SEB-style|Electron|📋 Planned|

### Target Users

|User|Use Case|
|---|---|
|Tier-2 / Tier-3 Universities|Internal exams, lab assessments, semester evaluations|
|Placement Cells|Coding rounds during campus recruitment drives|
|Department Faculty|Controlled programming evaluations for courses|
|Hackathon Organizers|On-site coding evaluation with real-time monitoring|
|Campus Recruitment Drives|Pre-placement technical screening rounds|

---

## 2. System Architecture

### Two-Layer Model

The system operates in two layers that work in sequence.

**Layer A — Control Plane (Cloud)** Used exclusively BEFORE the contest. Admin logs in, creates contest, uploads problems and testcases, generates and downloads the contest package. No judging happens here.

**Layer B — Data Plane (Local LAN)** Runs entirely on the organiser's laptop DURING the contest. No internet required. Hosts all judging, scoreboard, monitoring, and participant connections over LAN.

### Architecture Diagram

```
         ┌──────────────────────────┐
         │        YOUR CLOUD        │
         │   Contest Creation UI    │
         └────────────┬─────────────┘
                      │  Download Package
                      ▼
      ┌───────────────────────────────┐
      │      ADMIN LAPTOP SERVER      │
      │  ┌─────────────────────────┐  │
      │  │ API Server  (net/http)  │  │
      │  │ Judge Worker (Go)       │  │
      │  │ Submission Queue (chan) │  │
      │  │ Database (SQLite/GORM)  │  │
      │  │ Scoreboard Engine      │  │
      │  │ Monitoring Service     │  │
      │  └─────────────────────────┘  │
      └──────────────┬────────────────┘
                     │  LAN
       ┌─────────────┼─────────────┐
  Participant    Participant   Participant
  App (Electron) App (Electron) App (Electron)
```

### Package Dependency Graph

One-way dependencies only. Lower packages never import higher ones. No cycles.

```
models/          → imports nothing (pure structs + enums)
db/              → imports models
judgeEngine/     → imports models, db
api/handlers/    → imports models, db, judgeEngine
api/server.go    → imports all handlers, middleware
main.go          → wires everything, starts server
```

### File System Layout

```
server_data/
├── contest/                        ← READ-ONLY during contest
│   ├── manifest.json               ← contest config (name, times, join code, admin creds)
│   └── problems/
│       ├── 1/
│       │   ├── meta.json           ← time limit, memory limit, scoring config
│       │   ├── statement.md        ← problem statement rendered in participant app
│       │   └── tests/
│       │       ├── 1.in / 1.out
│       │       └── 2.in / 2.out
│       └── 2/ ...
└── runtime/                        ← MUTABLE during contest
    ├── db.sqlite                   ← single DB for everything
    ├── submissions/                ← PERSISTENT per submission
    │   └── {id}/
    │       ├── source.cpp
    │       ├── compile.log         ← only written on CE
    │       └── failure.json        ← only written on WA/RE/TLE
    └── workspaces/                 ← EPHEMERAL, deleted after execution
        └── {id}/
            ├── source.cpp
            ├── program             ← compiled binary
            ├── input.txt
            └── output.txt
```

---

## 3. Go Project File Structure

### Directory Layout

```
CampusCode/
├── main.go                         ← wiring only, zero logic
├── go.mod / go.sum
│
├── models/
│   ├── models.go                   ← all GORM structs
│   └── enums.go                    ← verdict constants, status values
│
├── db/
│   └── db.go                       ← InitDB(), GORM connection, AutoMigrate
│
├── judgeEngine/
│   ├── judge.go                    ← Submit(), public entry point
│   ├── queue.go                    ← job channel, overflow handling
│   ├── executor.go                 ← compile + run logic (was handlers.go)
│   ├── fileSystem.go               ← workspace create/copy/cleanup
│   └── helpers.go                  ← output comparison, utilities
│
├── api/
│   ├── server.go                   ← mux creation, route registration
│   ├── middleware.go               ← auth check, logging, CORS
│   └── handlers/
│       ├── health.go
│       ├── auth.go
│       ├── problems.go
│       ├── submissions.go
│       ├── contest.go
│       └── admin.go
│
└── server_data/                    ← data directory, NOT a Go package
```

### File Responsibilities

|File|Responsibility|
|---|---|
|`main.go`|Init DB → init judge → start workers → create API server → listen on port. No logic.|
|`models/models.go`|All GORM structs: Team, Session, DeviceOverride, Submission. Pure data shapes.|
|`models/enums.go`|Verdict constants (AC, WA, TLE, RE, CE, MLE), submission status values, language IDs.|
|`db/db.go`|Opens SQLite connection, runs AutoMigrate on all models. Knows path to runtime/db.sqlite.|
|`judgeEngine/judge.go`|Public API of the judge. Exposes `Submit(req)`. Orchestrates the full flow.|
|`judgeEngine/queue.go`|Buffered Go channel holding pending jobs. Feeds workers. Handles overflow gracefully — rejects submissions with a friendly error when full.|
|`judgeEngine/executor.go`|Compile via g++, run binary against each testcase, capture stdout, enforce time limit. Returns raw results to determine verdict.|
|`judgeEngine/fileSystem.go`|Create workspaces, copy source files, write input/output files, clean up after execution. No compilation logic here.|
|`judgeEngine/helpers.go`|Output comparison (trim, normalize whitespace), file read utilities, small reusable functions.|
|`api/server.go`|Creates the mux. Calls `handlers.Register()` for each domain. One line per route group. No logic.|
|`api/middleware.go`|Auth token validation, request logging, CORS headers. Wraps all protected routes.|
|`api/handlers/health.go`|`GET /api/health` → 200. Used by participant app for LAN server discovery.|
|`api/handlers/auth.go`|RegisterHandler, LoginHandler, OverrideHandler. All auth logic lives here.|
|`api/handlers/problems.go`|List problems, get single problem. Reads from filesystem and DB.|
|`api/handlers/submissions.go`|POST submit (calls judgeEngine.Submit), GET verdict polling, GET my submissions.|
|`api/handlers/contest.go`|GET contest metadata from manifest.json. In-memory start/stop state.|
|`api/handlers/admin.go`|Admin-only: all submissions, rejudge, team management, override approvals.|

### How main.go Wires Everything

```go
// Pseudocode — actual implementation will vary
func main() {
    db.InitDB()
    judgeEngine.Init(db.DB)
    judgeEngine.StartWorkers(4)
    server := api.NewServer(db.DB)
    server.Start(":8080")
}
```

---

## 4. Database Schema

**Single SQLite file at `server_data/runtime/db.sqlite`. One DB. Always.**

Managed by GORM. Shared across all packages via the `db/` package. Neither `judgeEngine` nor `api/handlers` own the DB — they both import `db/`.

### Team

```go
type Team struct {
    ID        int       `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" gorm:"unique"`
    Members   string    `json:"members" gorm:"type:text"`  // JSON array string
    LoginCode string    `json:"login_code" gorm:"unique"`
    IsActive  bool      `json:"is_active" gorm:"default:true"`
    CreatedAt time.Time `json:"created_at"`
}
```

|Field|Notes|
|---|---|
|ID|GORM auto-increments. Never set manually.|
|Name|Unique per contest.|
|Members|JSON-encoded string. e.g. `["John","Jane"]`. Never use `[]string` — GORM won't serialize it correctly with SQLite.|
|LoginCode|Generated by server. Format: `WOLF-4829`. Admin hands this to team physically.|
|IsActive|Admin can disable a team without deleting. Disabled teams cannot log in.|

### Session

```go
type Session struct {
    ID           int       `json:"id" gorm:"primaryKey"`
    TeamID       int       `json:"team_id" gorm:"not null;index"`
    DeviceID     string    `json:"device_id"`
    SessionToken string    `json:"session_token" gorm:"unique"`
    IsActive     bool      `json:"is_active" gorm:"default:true"`
    CreatedAt    time.Time `json:"created_at"`
    LastSeenAt   time.Time `json:"last_seen_at"`
    ExpiresAt    time.Time `json:"expires_at"`
}
```

| Field        | Notes                                                                                                          |
| ------------ | -------------------------------------------------------------------------------------------------------------- |
| TeamID       | Indexed for fast lookup. Foreign key to Team.ID.                                                               |
| DeviceID     | UUID generated by Electron on first launch, stored in electron-store. Never changes unless app is reinstalled. |
| SessionToken | `crypto/rand` 32-byte hex string. Sent as `Authorization: Bearer {token}` on every request.                    |
| IsActive     | Set to false on logout or when admin approves a device override.                                               |
| LastSeenAt   | Updated on every authenticated request. Used for idle detection in monitoring panel — free data.               |

### DeviceOverride

```go
type DeviceOverride struct {
    ID         int       `json:"id" gorm:"primaryKey"`
    TeamID     int       `json:"team_id" gorm:"not null;index"`
    Status     string    `json:"status"`   // "pending" / "approved" / "rejected"
    CreatedAt  time.Time `json:"created_at"`
    ResolvedAt time.Time `json:"resolved_at"`
}
```

|Field|Notes|
|---|---|
|TeamID|Which team is requesting the override.|
|Status|pending on creation. Admin sets to approved or rejected.|
|ResolvedAt|Null until admin acts. Audit trail.|

> Admin does NOT need OldDeviceID or NewDeviceID. He confirms with the team physically and approves. The system then deletes the old session and creates a new one with the new device_id. The device IDs are irrelevant to the admin.

### Submission

```go
type Submission struct {
    ID          int       `json:"id" gorm:"primaryKey"`
    TeamID      int       `json:"team_id" gorm:"not null;index"`
    ProblemID   int       `json:"problem_id" gorm:"not null;index"`
    Language    string    `json:"language"`           // cpp / python / java
    Status      string    `json:"status"`             // queued / compiling / running / done / error
    Verdict     string    `json:"verdict"`            // AC / WA / TLE / RE / CE / MLE — null until done
    SubmittedAt time.Time `json:"submitted_at"`
    JudgedAt    time.Time `json:"judged_at"`          // null until verdict written
}
```

### What Does NOT Have a DB Table

|Thing|Where it lives|
|---|---|
|Contest metadata|`server_data/contest/manifest.json` — read-only, never duplicated in DB|
|Problem data & testcases|`server_data/contest/problems/{id}/` — filesystem, not DB|
|Contest running state|In-memory variable on the server. No persistence needed.|
|Admin credentials|In the contest package config/manifest. Downloaded with the package. No DB table.|

---

## 5. Authentication & Session System

Session-based auth with server-side session store in SQLite. **Not JWT** — JWT is stateless and cannot support device binding or forced invalidation. Sessions are stored in DB, checked on every request, and revocable at any time.

### Auth Flow

**Registration (Admin does this before contest)**

```
Admin enters team details in dashboard
        ↓
Server creates Team record in DB
        ↓
Server generates LoginCode (e.g. WOLF-4829)
        ↓
Returns { name, login_code } to admin
        ↓
Admin notes it down, hands to team physically
```

**First Login**

```
Team submits { team_name, login_code, device_id }
        ↓
Server validates credentials
        ↓
No existing active session → this is first login
        ↓
Creates Session { team_id, device_id, session_token, is_active: true }
        ↓
Returns { session_token } to client
        ↓
Client stores token, sends as Bearer on every request
```

**Re-login Same Device**

```
Team submits { team_name, login_code, device_id }
        ↓
Server finds existing session
        ↓
device_id matches stored DeviceID → allowed
        ↓
Returns fresh session_token
```

**Re-login Different Device**

```
Team submits { team_name, login_code, device_id }
        ↓
Server finds existing session
        ↓
device_id MISMATCH → rejected (409 Conflict)
        ↓
Server creates DeviceOverride { team_id, status: "pending" }
        ↓
Admin sees override request in dashboard
        ↓
Admin physically confirms with team → approves
        ↓
Old session: IsActive = false
        ↓
New session created with new device_id
```

**Admin Login** Single credential from `manifest.json`. Returns a session_token the same way. No registration flow. No DB table for admin.

### device_id

UUID generated by Electron on first launch using `crypto.randomUUID()`. Stored in `electron-store` (persistent local storage). Sent as a header on every auth request. Does not change for that machine unless the app is reinstalled. Not cryptographically binding — it is a persistent identifier that makes re-login on a different device detectable.

### Session Token

Generated using `crypto/rand` — 32 random bytes encoded as hex. Stored in Sessions table. Every protected route validates the token against the DB. Revocable instantly by setting `IsActive = false`.

### LoginCode Generation

Uses `math/rand/v2`. Human-readable format like `WOLF-4829`. Short, uppercase, easy to type. Not a cryptographic secret — handed physically by admin to team. `math/rand/v2` is sufficient.

---

## 6. Full API Reference

All endpoints prefixed with `/api`. Protected routes require `Authorization: Bearer {session_token}` header.

### Health

|Method|Endpoint|Auth|Request|Response|
|---|---|---|---|---|
|GET|`/api/health`|None|—|`{ status: "ok", time }` — used for LAN discovery|

### Auth

|Method|Endpoint|Auth|Request|Response|
|---|---|---|---|---|
|POST|`/api/auth/register`|Admin|`{ team_name, members }`|`{ name, login_code }`|
|POST|`/api/auth/login`|None|`{ team_name, login_code, device_id }`|`{ session_token }` or 409 + override created|
|POST|`/api/admin/auth/login`|None|`{ username, password }`|`{ session_token }`|
|POST|`/api/admin/override/:id`|Admin|`{ action: "approve" or "reject" }`|`{ status }`|

### Problems

|Method|Endpoint|Auth|Request|Response|
|---|---|---|---|---|
|GET|`/api/problems`|Participant|—|Array of `{ id, title, time_limit, memory_limit, test_count }`|
|GET|`/api/problems/:id`|Participant|—|`{ id, title, statement_md, time_limit, memory_limit }`|

### Submissions

|Method|Endpoint|Auth|Request|Response|
|---|---|---|---|---|
|POST|`/api/submissions`|Participant|`{ problem_id, language, code }`|`{ submission_id, status: "queued" }`|
|GET|`/api/submissions/:id`|Participant|—|`{ id, verdict, status, judged_at }` — poll until status = done|
|GET|`/api/submissions/me`|Participant|—|Array of team's own submissions|
|GET|`/api/submissions`|Admin|`?problem_id, ?team_id, ?verdict`|All submissions, filterable|
|POST|`/api/submissions/:id/rejudge`|Admin|—|`{ status: "queued" }` — re-queues the job|

### Contest

|Method|Endpoint|Auth|Request|Response|
|---|---|---|---|---|
|GET|`/api/contest`|Participant|—|`{ name, start_time, end_time, status, problem_ids }` from manifest.json|
|POST|`/api/admin/contest/start`|Admin|—|Sets in-memory contest status to running|
|POST|`/api/admin/contest/stop`|Admin|—|Sets in-memory contest status to ended|

### Scoreboard

|Method|Endpoint|Auth|Request|Response|
|---|---|---|---|---|
|GET|`/api/scoreboard`|Participant|—|Array of `{ team_id, team_name, solved, penalty, problems[] }`|

### Admin — Teams & Monitoring

|Method|Endpoint|Auth|Request|Response|
|---|---|---|---|---|
|GET|`/api/admin/teams`|Admin|—|All registered teams with active session status|
|POST|`/api/admin/teams/:id/disable`|Admin|—|Sets `Team.IsActive = false`|
|GET|`/api/admin/monitor`|Admin|—|All participants: last_seen_at, submission count, focus switches|

### Participant Monitoring (Planned — build when Electron app is being built)

|Method|Endpoint|Auth|Request|Response|
|---|---|---|---|---|
|POST|`/api/monitor/heartbeat`|Participant|`{ timestamp }`|Updates `LastSeenAt`. Called every 30s by Electron.|
|POST|`/api/monitor/screenshot`|Participant|multipart image|Stores screenshot with timestamp. Admin views in panel.|
|POST|`/api/monitor/focusevent`|Participant|`{ type: "blur" or "focus" }`|Logs focus switch. Increments counter visible to admin.|

### WebSocket (Planned — do not build yet)

```
WS /ws/scoreboard    ← push scoreboard updates to browser after each verdict
WS /ws/monitor       ← push participant activity events to admin dashboard
```

---

## 7. Judge Engine

The most critical internal component. Fully contained in `judgeEngine/`. Has **zero HTTP knowledge** — exposes plain Go functions. The API layer calls into it.

### Submission Flow

```
POST /api/submissions
        ↓
api/handlers/submissions.go   decode + validate HTTP request
        ↓
judgeEngine.Submit(req)        public entry point
        ↓
Save submission to DB          status: queued
        ↓
Write source.cpp to runtime/submissions/{id}/
        ↓
Push job ID to queue channel
        ↓
Worker picks job from channel
        ↓
Create workspace at runtime/workspaces/{id}/
        ↓
Copy source.cpp to workspace
        ↓
Compile: g++ source.cpp -O2 -std=c++17 -o program
        ↓ (if CE → write compile.log, set verdict CE, cleanup, done)
For each testcase in problems/{problem_id}/tests/:
        ↓
    Copy testcase input → workspace/input.txt
        ↓
    Run ./program < input.txt > output.txt  (with timeout)
        ↓ (timeout → TLE) (non-zero exit → RE)
    Compare output.txt with expected .out
        ↓ (mismatch → WA, stop running more testcases)
All testcases passed → AC
        ↓
Write verdict + judged_at to DB
        ↓
Delete workspace (ephemeral)
        ↓
Signal scoreboard via Go channel
```

### Verdicts

|Verdict|Meaning|Trigger|
|---|---|---|
|AC|Accepted|All testcases passed, output matches.|
|WA|Wrong Answer|Output does not match expected for at least one testcase.|
|TLE|Time Limit Exceeded|Process did not finish within problem's time limit. Process killed.|
|RE|Runtime Error|Program exited with non-zero exit code (segfault, exception, etc).|
|CE|Compile Error|g++ returned non-zero. compile.log written. Program never runs.|
|MLE|Memory Limit Exceeded|Process exceeded memory limit. Add later with Docker or /proc monitoring.|

### Worker Pool

- 4–8 parallel workers, configurable. Each worker is a goroutine reading from the shared job channel.
- Buffered Go channel absorbs burst submissions. If full, new submissions are rejected gracefully with a clear error.
- Each submission gets its own workspace directory. Workers never share filesystem state.
- Docker sandboxing wraps the same executor logic — add it after the demo is working end-to-end.

### Sandbox Enforcement (Target State)

|Constraint|Current|Target|
|---|---|---|
|Time limit|✅ Process timeout|✅|
|Memory limit|⚠️ Not yet|✅ via Docker or /proc|
|CPU limit|⚠️ Not yet|✅ via Docker|
|Process count|⚠️ Not yet|✅ via Docker|
|Network access|⚠️ Not blocked|✅ Docker network: none|
|Filesystem isolation|⚠️ Workspace only|✅ Docker volume isolation|

### Scoreboard Notification

After writing a verdict, the judge signals the scoreboard service via a **Go channel** — not HTTP, not WebSocket. Both live in the same process. The scoreboard listens on that channel and recomputes standings when a new verdict arrives.

```go
// In judgeEngine — after writing verdict
verdictCh <- VerdictEvent{ SubmissionID: id, TeamID: teamID, ProblemID: problemID, Verdict: "AC" }

// In scoreboard service — listening
for event := range verdictCh {
    scoreboard.Recompute(event)
}
```

---

## 8. Scoreboard

Separate package from the judge. The judge writes verdicts; the scoreboard reads them. They communicate via a Go channel. The scoreboard never calls into the judge.

### ICPC Scoring

- Each solved problem contributes 1 to the team's solved count.
- Penalty = time of AC submission (in minutes from contest start) + 20 minutes per wrong attempt before AC.
- Teams ranked by: solved count descending, then total penalty ascending.
- Problems never solved have no penalty regardless of wrong attempts before giving up.

### Data Flow

```
judgeEngine writes AC verdict to DB
        ↓
Fires event on verdictCh Go channel
        ↓
scoreboard service receives event
        ↓
Reads all verdicts from DB for the contest
        ↓
Recomputes standings in memory
        ↓
Serves via GET /api/scoreboard (polling now)
        ↓
Later: pushes via WS /ws/scoreboard
```

**Build the scoreboard LAST.** It is a DB query + sort. Trivial once verdicts exist. Do not block API or participant app work on it.

---

## 9. Participant App (Electron)

Cross-platform desktop app for Windows and macOS. Behaves like Safe Exam Browser but fully integrated into the contest ecosystem. Downloaded by participants before contest day.

### Core Responsibilities

|Feature|Description|
|---|---|
|LAN Discovery|Auto-discover server via mDNS/broadcast, or manually enter the join code. Calls `GET /api/health` to verify.|
|Kiosk Mode|Launch in fullscreen. No URL bar, no external navigation, no devtools. Electron `BrowserWindow` with appropriate flags.|
|Focus Detection|Detect window blur/focus events. Log every switch. POST to `/api/monitor/focusevent`. Admin sees count in dashboard.|
|Screenshot Upload|Capture screenshot every N seconds (configurable). POST to `/api/monitor/screenshot`. Admin sees grid in dashboard.|
|Heartbeat|POST to `/api/monitor/heartbeat` every 30 seconds. Server updates `LastSeenAt`. Admin sees idle status.|
|Problem Viewer|Embedded browser renders problem statements (Markdown). Clean read-only UI.|
|Code Editor|Embedded code editor (Monaco or CodeMirror). Language selection. Submit button.|
|Verdict Polling|After submit, poll `GET /api/submissions/:id` every 2 seconds until status = done. Show verdict inline.|

### device_id

On first launch, Electron generates a UUID using `crypto.randomUUID()` and stores it in `electron-store` (persistent local storage). This is the `device_id` sent with every auth request. It never changes for that machine unless the app is reinstalled.

### Security Model

This is a campus exam tool, not a military lockdown. Goal: raise the cost of cheating to impractical in a physically supervised setting — not to make it technically impossible.

- Fullscreen enforced — leaving fullscreen logs a focus event visible to admin immediately
- No external navigation — all links intercepted and blocked in the embedded browser
- DevTools disabled in production build
- Focus switches logged and visible to admin in real time
- Screenshots give admin visual confirmation of activity
- Determined technical attacker can still bypass — this is a **known and documented limitation**

### Build Order

Build the participant app AFTER the API layer is complete and the admin dashboard is working. You need the API to exist before the client has anything to talk to, and you need the dashboard so you can observe what the app is doing during testing.

---

## 10. Build Order & Current Status

### What Is Done

- ✅ Judge engine — compile, run against testcases, AC/WA/CE/RE verdict, DB write
- ✅ File system management — workspace create/copy/cleanup
- ✅ Submission queue — buffered Go channel, worker pool
- ✅ DB schema — Team, Session, DeviceOverride, Submission models
- ✅ Auth — team registration, login, device binding, override flow

### Remaining Build Order

|#|Task|Effort|Notes|
|---|---|---|---|
|1|`GET /api/health`|1 hour|First endpoint. Confirms HTTP setup. Used for LAN discovery.|
|2|`GET /api/problems` + `/api/problems/:id`|1 day|Read from filesystem. No judge interaction. Easy win.|
|3|`GET /api/submissions/:id`|2 hours|Poll verdict. Judge already writes to DB — just expose it.|
|4|`GET /api/submissions/me`|2 hours|Filter by team_id from session. Simple DB query.|
|5|`GET /api/contest`|2 hours|Read manifest.json. Return contest metadata.|
|6|Admin: start/stop contest|Half day|In-memory state toggle. No persistence needed.|
|7|Admin: all submissions + rejudge|1 day|Read from DB. Rejudge re-queues the job.|
|8|Admin: team management|Half day|List teams, disable team, view override requests.|
|9|React admin dashboard — skeleton|3 days|Submissions view, scoreboard tab, participant list.|
|10|Electron participant app — skeleton|3 days|Problem viewer, code editor, submit, verdict display.|
|11|Kiosk mode + focus detection|2 days|Fullscreen enforcement, focus event logging, screenshot upload.|
|12|Scoreboard logic + endpoint|1 day|ICPC penalty calculation. Polling version first.|
|13|WebSocket — scoreboard|1 day|Push updates to browser after each verdict.|
|14|WebSocket — monitoring panel|1 day|Push focus events, heartbeat status to admin dashboard.|
|15|TLE + memory limits|1 day|Process timeout (TLE already planned). Memory via /proc or Docker.|
|16|Docker sandboxing|2–3 days|Wrap executor in Docker container. Network isolation. Very last step.|

---

## 11. Performance Targets & Security Model

### Performance Targets

|Metric|Target|Strategy|
|---|---|---|
|Concurrent users|100|LAN only, SQLite handles this trivially|
|Peak submissions|~100/min|Buffered queue absorbs bursts, workers process at steady rate|
|Avg compile time|< 2 seconds|Direct g++ call, no container overhead until Docker added|
|Server uptime|No crashes|Graceful queue overflow rejection, WAL on SQLite|
|DB query time|< 10ms|Indexed TeamID + ProblemID, SQLite on local SSD|

### Security Layers

|Layer|Mechanism|Limitation|
|---|---|---|
|Network|LAN isolation — no internet during contest|Admin must control the router/network|
|Browser|Embedded restricted browser, no external navigation|OS-level browser accessible if app is killed|
|Focus|Every window blur logged and flagged|Physical shoulder-surfing not detectable|
|Visual|Periodic screenshots sent to admin|Interval-based — activity between screenshots not captured|
|Device|device_id binding — new device triggers admin override|UUID can technically be copied|
|Judge|Isolated workspace per submission, process killed on timeout|No Docker yet — resource bombs possible|
|Human|Admin physically invigilates using monitoring dashboard|**This is the strongest layer. Technology assists, human decides.**|

---

## 12. Key Design Decisions & Quick Reference

### Core Design Decisions

|Decision|Reasoning|
|---|---|
|One DB|Single SQLite file. No Postgres, no Redis, no second DB. Everything in db.sqlite.|
|No JWT|Session-based auth. Sessions stored in DB. Revocable, supports device binding, supports override flow.|
|No service layer|Handlers call DB and judgeEngine directly. No intermediate service struct. Project is too small for that abstraction.|
|No global DB var|DB passed as dependency via function args or handler struct. No `init()` magic.|
|In-memory contest state|Contest running/stopped is an in-memory variable. Does not need to survive server restart — admin manually restarts if needed.|
|Filesystem over DB|Problem data, testcases, statements live on filesystem. DB stores metadata and runtime state only.|
|Go channels not WebSocket|Judge → scoreboard communication is a Go channel. WebSocket is only ever for server → browser.|
|`math/rand/v2` for login codes|Not a cryptographic secret. Fast, sufficient, human-readable output.|
|Singular model names|`Team` not `Teams`, `Session` not `Sessions`. GORM convention.|
|Members as JSON string|Store `[]string` as JSON string in DB. Never use `[]string` directly with GORM + SQLite — it won't serialize correctly.|

### What NOT to Build Yet

- **Docker sandboxing** — add after the demo works end to end. It's hardening, not functionality.
- **WebSocket endpoints** — polling works fine until the dashboard UI is actually built.
- **Scoreboard** — build last. It's trivial once verdicts exist. Don't block other work on it.
- **AI cheating detection** — explicitly out of MVP scope.
- **Partial scoring** — out of MVP scope.
- **Function-based problems (LeetCode style)** — out of MVP scope.
- **Rate limiting** — not needed at LAN campus scale (100 users).
- **API versioning** — single deployment, no external consumers.
- **Pagination** — 100 users, return everything.

### Common Questions

|Question|Answer|
|---|---|
|How many DBs?|One. `server_data/runtime/db.sqlite`. Always.|
|Who owns the DB?|Nobody. `db/` package is shared. Both `judgeEngine` and `api/handlers` import `db/`.|
|Where does contest config live?|`server_data/contest/manifest.json`. Read-only. Not in DB.|
|Where do problem testcases live?|`server_data/contest/problems/{id}/tests/`. Filesystem. Not in DB.|
|What triggers scoreboard update?|Go channel signal from `judgeEngine` after verdict is written to DB.|
|How does admin log in?|Credential from manifest/package config. No DB table. Returns session token same way as teams.|
|Where does device_id come from?|Electron generates UUID on first launch, stores in `electron-store`, sends as header on every auth request.|
|When to add Docker?|After the full demo works end to end. It is security hardening, not a functional requirement.|
|Does `judgeEngine` import `api`?|Never. One-way dependency only. `judgeEngine` imports `db` and `models`. Nothing imports `judgeEngine` except `api/handlers`.|
|Is there a cycle between `judgeEngine` and `db`?|No. `db` imports nothing except the standard library and `models`. `judgeEngine` imports `db`. One direction only.|

---

_CampusCode — Internal Developer Reference — Pre-release v0.1.0 — May 2026_ 