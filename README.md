# Campus Secure Coding Assessment Platform

### Master Product Reference Document

---

> **Positioning:** Offline-capable, LAN-based secure coding contest and assessment infrastructure for universities and campus placements.
> 
> This is **not** a Codeforces replacement. **Not** an enterprise proctoring system. It is a controlled, plug-and-play campus deployment system built for simplicity, reliability, and offline operation.

---

## Table of Contents

1. [What This Product Is](#1-what-this-product-is)
2. [Target Users](#2-target-users)
3. [System Architecture](#3-system-architecture)
4. [Core Components](#4-core-components)
5. [Judge Subsystem](#5-judge-subsystem)
6. [Deployment Flow](#6-deployment-flow)
7. [MVP Scope](#7-mvp-scope)
8. [MVP+ Differentiators](#8-mvp-differentiators)
9. [Performance Targets](#9-performance-targets)
10. [Security Model](#10-security-model)
11. [Known Technical Risks](#11-known-technical-risks)
12. [Technology Stack](#12-technology-stack)
13. [Long-Term Vision](#13-long-term-vision)
14. [Competitive Advantage](#14-competitive-advantage)

---

## 1. What This Product Is

A **hybrid cloud + LAN-based secure coding contest and assessment platform** designed for universities and campus placement drives.

It allows institutions to conduct the following **entirely without internet access** during the contest:

- Competitive programming contests
- Placement / recruitment coding rounds
- Internal lab exams and evaluations
- Hackathon coding assessments

The system bundles four things into one unified ecosystem:

|Component|Role|
|---|---|
|Local judging engine|Codeforces-style automated evaluation|
|LAN contest server|Hosts everything locally on organizer's machine|
|Participant client|Lightweight kiosk-mode app (like Safe Exam Browser)|
|Admin monitoring dashboard|Real-time oversight of all participants|

**Core differentiator:** Offline capability + integrated monitoring + one-click deployment.

---

## 2. Target Users

|User|Use Case|
|---|---|
|Tier-2 / Tier-3 universities|Internal exams, lab assessments|
|Placement cells|Coding rounds during campus recruitment|
|Department faculty|Controlled programming evaluations|
|Hackathon organizers|On-site coding evaluation|
|Campus recruitment drives|Pre-placement technical screening|

---

## 3. System Architecture

The system operates in two layers that work in sequence — cloud before the contest, local during the contest.

### Layer A — Control Plane (Cloud)

Used exclusively **before** the contest. No judging happens here.

Responsibilities:

- Admin login and account management
- Contest creation and configuration
- Problem and testcase uploads
- Scoring rule configuration
- License management
- Generating a downloadable contest package

### Layer B — Data Plane (Local LAN Server)

Runs **entirely on the organizer's laptop** during the contest. No internet required.

Responsibilities:

- Host the contest and serve the UI over LAN
- Accept and judge all submissions
- Run the live scoreboard
- Monitor participant activity
- Store all submissions and results
- Handle screenshot and heartbeat data from clients

### Architecture Diagram

```
         ┌──────────────────────────┐
         │        YOUR CLOUD        │
         │   Contest Creation UI    │
         └────────────┬─────────────┘
                      │
               Download Package
                      │
                      ▼
      ┌───────────────────────────────┐
      │      ADMIN LAPTOP SERVER      │
      │  ┌─────────────────────────┐  │
      │  │ API Server              │  │
      │  │ Judge Worker            │  │
      │  │ Submission Queue        │  │
      │  │ Database                │  │
      │  │ Scoreboard Engine       │  │
      │  │ Monitoring Service      │  │
      │  └─────────────────────────┘  │
      └──────────────┬────────────────┘
                     │  LAN
       ┌─────────────┼─────────────┐
       │             │             │
  Participant    Participant   Participant
  App (Win/Mac)  App (Win/Mac) App (Win/Mac)
```

---

## 4. Core Components

### 4.1 Organizer App (Local Server)

Distributed as a **single executable**. The organizer launches it and the entire local server stack spins up automatically.

Handles:

- Starting the local API server
- Initializing the database
- Starting judge workers
- Serving the contest UI over LAN
- Accepting and queuing submissions
- Updating the live scoreboard
- Receiving screenshots and heartbeats
- Logging all participant events

---

### 4.2 Participant App

A standard client app downloaded by participants **before** the contest day. Available for Windows and macOS.

Handles:

- LAN server discovery (auto or via join code)
- Secure connection to the local server
- Launching an embedded, restricted browser
- Enforcing fullscreen / kiosk mode
- Logging focus switches
- Capturing and uploading periodic screenshots
- Sending heartbeat signals to the server

> Behaves like Safe Exam Browser but is fully integrated into the contest ecosystem.

---

### 4.3 Admin Dashboard

Web-based panel served locally by the organizer app.

Capabilities:

- Create and configure contests
- Upload problems and testcases
- Start / stop the contest
- View live submissions and verdicts
- Rejudge individual submissions
- View and manage the live scoreboard
- Monitor participant activity (screenshots, focus events, idle status)

---

## 5. Judge Subsystem

This is the most critical internal component.

### Submission Flow

```
Submission received
       ↓
Stored in DB
       ↓
Pushed to queue
       ↓
Worker picks job
       ↓
Compile inside sandbox
       ↓
Run against all testcases
       ↓
Generate verdict
       ↓
Store result
       ↓
Trigger scoreboard update
```

### Sandbox Enforcement

Every submission must be isolated with the following hard limits:

|Constraint|Enforced|
|---|---|
|Time limit|✅|
|Memory limit|✅|
|CPU limit|✅|
|Process count limit|✅|
|Network access|❌ Blocked|
|Filesystem access|❌ Isolated|

### MVP Implementation Strategy

- Container-based sandbox (Docker)
- 4–8 parallel judge workers (configurable)
- Submission queue with overflow handling
- Reject submissions gracefully when server is overloaded

### Supported Languages (Initial)

- C++
- Python
- Java

---

## 6. Deployment Flow

### Before the Contest (Admin)

1. Log in to the cloud web platform
2. Create a contest and configure settings
3. Upload problems and testcases
4. Generate and download the contest package (server bundle)

### On Contest Day (Organizer Setup)

1. Open the organizer app on the laptop
2. Click **"Start Server"**
3. App displays the LAN IP and join code

### Participants Joining

1. Open the participant client app
2. Enter the join code **or** use auto-discovery
3. Contest begins in fullscreen kiosk mode

---

## 7. MVP Scope

The following must be complete and stable before any additional features are added.

### 7.1 Contest Engine

- LAN API server
- Problem viewer UI
- Code submission system
- Automated judge
- ICPC-style scoring (penalty-based)
- Real-time scoreboard via WebSockets

### 7.2 Admin Dashboard

- Contest creation and management
- Problem + testcase upload
- Start / stop contest controls
- Submission viewer with rejudge option
- Scoreboard view

### 7.3 Participant Client

- Fullscreen enforcement
- No URL bar or external navigation
- Focus change detection and logging
- Periodic screenshot upload
- Heartbeat signal

### 7.4 Network

- Fully LAN-based operation
- No internet dependency
- Support for 50–100 concurrent participants
- Auto server discovery or join-code-based connection

> **Out of scope for MVP:** AI-based cheating detection, advanced behavioral analytics, distributed judging, partial scoring, function-based problems.

---

## 8. MVP+ Differentiators

Features that go beyond MVP and set this apart from SEB + Codeforces combinations.

### 8.1 Integrated Monitoring Panel

Admins get a unified real-time view of all participants:

- Focus switch count per participant
- Screenshot grid (all participants visible at once)
- Active / idle status
- Submission frequency
- Live submission stream

> Safe Exam Browser provides none of this centralized visibility.

### 8.2 Flexible Question Types

Beyond standard competitive programming:

- Standard I/O problems
- Function-based problems (LeetCode style)
- Partial scoring
- Manual grading option
- Hackathon-style evaluation submissions

### 8.3 Full Offline Mode

The entire contest — judging, scoreboard, monitoring — runs without any internet connection. This is the primary differentiator for campus environments with unreliable connectivity.

---

## 9. Performance Targets

|Metric|Target|
|---|---|
|Concurrent users|100|
|Peak submissions|~100 / minute|
|Average compile time|< 2 seconds|
|Server uptime|No crashes during contest|

### Strategy to Hit Targets

- Cap concurrent judge workers (prevents CPU exhaustion)
- Submission queue absorbs burst load
- Graceful rejection with feedback if queue is full
- Continuous CPU/memory monitoring during the contest

---

## 10. Security Model

This is a **realistic** security model designed for campus environments, not a military-grade lockdown.

### What Is Secured

|Layer|Mechanism|
|---|---|
|Network|LAN isolation — no internet during contest|
|Browser|Embedded restricted browser, no external navigation|
|Focus|Logged and flagged on every switch|
|Visual|Periodic screenshots sent to admin|
|Human|Admin invigilation via monitoring dashboard|

### What This Is Not

- Not a kernel-level lockdown
- Not bypass-proof
- Not designed to stop a determined, technically sophisticated attacker

The goal is to raise the cost of cheating to the point where it becomes impractical in a supervised exam setting.

---

## 11. Known Technical Risks

|Risk|Mitigation|
|---|---|
|Sandbox escape from malicious code|Strict container resource limits + process isolation|
|Resource exhaustion (CPU/memory bombs)|Hard limits + queue overflow handling|
|Shortcut / OS-level bypass on Windows/macOS|Extensive testing per platform, known limitation documented|
|Network instability on LAN|Heartbeat detection + reconnect logic|
|Database corruption under high load|Write-ahead logging, stress testing, recovery mechanisms|

---

## 12. Technology Stack

> This is the suggested stack. Final decisions are not locked in.

| Layer                         | Technology                                       |
| ----------------------------- | ------------------------------------------------ |
| Backend                       | Go                                               |
| Database                      | SQLite                                           |
| Sandbox                       | Docker-based container                           |
| Frontend (Admin + Contest UI) | React                                            |
| Participant App               | Electron (cross-platform)                        |
| Realtime                      | WebSockets (scoreboard + monitoring)             |
| Native APIs                   | Focus detection, screenshot capture via Electron |

---

## 13. Long-Term Vision

A plug-and-play campus assessment infrastructure with the following properties:

- **Installs in under 10 minutes** on any standard laptop
- **Handles 100 concurrent users** without manual tuning
- **Requires minimal technical expertise** from the organizer
- **Provides structured monitoring** with a real audit trail
- **Operates entirely offline** when needed

The end goal is to make conducting a fair, monitored, offline coding assessment as easy as starting a single application.

---

## 14. Competitive Advantage

### vs. Codeforces + Safe Exam Browser

|Feature|CF + SEB|This Platform|
|---|---|---|
|Works offline|❌|✅|
|Integrated monitoring|❌|✅|
|Single deployment unit|❌|✅|
|No external infra dependency|❌|✅|
|Flexible question formats|❌|✅|

### vs. DOMjudge

|Feature|DOMjudge|This Platform|
|---|---|---|
|Easy installation|❌ Complex|✅ One executable|
|Windows / macOS support|⚠️ Limited|✅ Native|
|Designed for small colleges|❌|✅|
|Integrated participant client|❌|✅|

---

_Last updated: March 2026 — Pre-development reference draft._