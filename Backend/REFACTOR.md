# CampusCode Backend — Refactor Analysis

**Scope:** `CampusCode-Server/Backend/` (23 Go files, ~1,867 lines). `testing/` excluded.
**Date:** 2026-09-06
**Nature of this document:** analysis only. No code was modified.

> **Scope boundary.** This document covers **structure, readability, and maintainability only** — how the code is organized, regardless of whether its current behaviour is correct.
>
> Behavioural defects, security findings, and data-integrity issues have been separated out and live in the companion **Correctness Audit**. Sections **2**, **6**, and parts of **3**, **7**, and **8** were relocated there; their numbers are deliberately left vacant below so cross-references stay stable and it is obvious that nothing was silently dropped.
>
> Where a structural finding here has a behavioural consequence, this document describes only the structure. The consequence is the audit's business.

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. *(relocated — Correctness Audit)*
3. [Shared State & Ownership](#3-shared-state--ownership)
4. [Architecture: The Missing Service Layer](#4-architecture-the-missing-service-layer)
5. [Layering Violations & Dependency Direction](#5-layering-violations--dependency-direction)
6. *(relocated — Correctness Audit)*
7. [The Judge Subsystem: Structural Findings](#7-the-judge-subsystem-structural-findings)
8. [Persistence Layer: Structural Findings](#8-persistence-layer-structural-findings)
9. [HTTP Layer: Routing, Contracts, Consistency](#9-http-layer-routing-contracts-consistency)
10. [Types, Naming & Domain Modelling](#10-types-naming--domain-modelling)
11. [Configuration & Magic Numbers](#11-configuration--magic-numbers)
12. [Testability](#12-testability)
13. [Hygiene: Formatting, Dead Code, Dependencies, Comments](#13-hygiene-formatting-dead-code-dependencies-comments)
14. [Proposed Target Structure](#14-proposed-target-structure)
15. [Prioritized Roadmap](#15-prioritized-roadmap)

---

## 1. Executive Summary

The codebase is small, readable in an afternoon, and the core mechanics are genuinely implemented rather than stubbed. That matters for a refactor assessment: everything below is *relocation and renaming*, not rewriting. The logic largely exists and is largely in one piece — it is simply in the wrong packages, reachable from too many places, and impossible to isolate.

The central structural problem is that **there is no service layer**. HTTP handlers parse requests, enforce business rules, run raw GORM queries, write files to disk, and push to the judge queue in the same function body. `db.DB` is a package-level global reached for directly from `api/handlers`, which welds the HTTP layer to the persistence layer with no seam between them. The immediate cost is that a rule like "may this team submit?" has no single home, so it exists in three places with three slightly different definitions. The downstream cost is that nothing can be exercised without a real SQLite file and a real Docker daemon — which is why there are **zero test files** in the repository.

The second structural problem is **package-level mutable state with no owner**. `models.ActiveContest`, `db.DB`, `judge.SubmissionQueue`, `judge.ActiveManifest`, and `ws.Clients` are all exported globals that any package may read or mutate. This is what makes the dependency graph hard to reason about and what blocks every testability improvement in §12.

The third is **dependency direction**: `judge` — the core domain — imports `api/ws`, the transport layer. The graph currently reads backwards.

None of the three requires a rewrite. §14 sketches the destination; §15 sequences the path.

### Scorecard

| Dimension | State | Notes |
| --- | --- | --- |
| Layering | 🔴 | No service layer; `judge` imports `api/ws`; handlers do file I/O |
| Encapsulation | 🔴 | Five exported mutable globals; `db.DB` reachable from anywhere |
| Testability | 🔴 | 0 tests; globals everywhere; `http.DefaultServeMux` |
| Rule locality | 🔴 | Team-status rule defined 3×; penalty formula inside a repository function |
| Formatting | 🔴 | 23/23 files fail `gofmt` |
| Naming & modelling | ⚠️ | Plural type names; stringly-typed statuses; 3 overlapping `Problem` types |
| Configuration | ⚠️ | ~18 tunables hardcoded across 8 files |
| Dead code | ⚠️ | ~19 commented-out blocks, 1 empty file, 1 stub struct |
| Dependencies | ⚠️ | ~25 unused modules (gin, mongo-driver, quic-go) |
| API consistency | ⚠️ | Two response formats; mixed singular/plural routes; ID type varies by endpoint |

### Verification performed

```
$ gofmt -l . | grep -v testing     → 23 files (all of them)
$ find . -name "*_test.go"         → 0
$ go mod tidy -diff                → removes gin, mongo-driver, quic-go, sonic, validator, …
```

---

## 3. Shared State & Ownership

*(Race-condition findings from the original §3 are in the Correctness Audit. What follows is the structural finding underneath them: state with no owner.)*

### 3.1 🔴 Five exported mutable globals, none with an owner

| Global | Declared | Mutated by | Read by |
| --- | --- | --- | --- |
| `models.ActiveContest` | [models.go:6](models/models.go#L6) | `main`, `AdminChangeContestState` | middleware, ws ticker, judge workers |
| `db.DB` | [db.go:11](db/db.go#L11) | `InitDB` | ~15 sites across 4 packages |
| `judge.SubmissionQueue` | [queue.go:10](judge/queue.go#L10) | `InitializeQueue` | handlers, workers |
| `judge.ActiveManifest` | [fileSystem.go:64](judge/fileSystem.go#L64) | `main` | `AdminLoginHandler` |
| `ws.Clients` + `ws.Mu` | [ws.go:13-14](api/ws/ws.go#L13-L14) | handlers, broadcast | broadcast |

The common shape: a package declares a variable, another package's `main` initializes it, and a third package mutates it. No type owns any of this state, so there is nowhere to put an invariant and nowhere to look to find out who is allowed to change what.

`models.ActiveContest` is the sharpest case. It is a bare `*Contest` — a **pointer** handed to every reader, so `RequireContestOpen`, `BroadcastTime`, and the judge workers all hold a mutable reference to the same struct. Any of them could legally write through it. `AdminChangeContestState` does exactly that ([admin.go:202-209](api/handlers/admin.go#L202-L209)), mutating three fields in place.

**Structural fix — a type that owns the state and hands out copies:**

```go
// service/contest.go
type State struct {
    mu      sync.RWMutex
    contest models.Contest
}

func (s *State) Snapshot() models.Contest {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.contest        // value copy — callers cannot mutate through it
}

func (s *State) Apply(fn func(*models.Contest)) models.Contest { … }
func (s *State) IsOpen(now time.Time) bool { … }
```

Returning a value rather than a pointer is the load-bearing part. It makes "who may mutate the contest?" answerable — exactly one method on one type — instead of "anyone holding the pointer."

`IsOpen` is also where §4.3's homeless contest-openness rule belongs.

### 3.2 🟠 `ScoreBoard` duplicates state that is derivable

[models.go:73-80](models/models.go#L73-L80)

`models.ScoreBoard` persists `TotalSolved`, `TotalPenalty`, and `LastAcTime` — every one of which is an aggregate over `TeamProblemStats`, which is *also* persisted. Two tables model the same facts at two levels of granularity, and a 45-line function ([repository.go:48-92](db/repository.go#L48-L92)) exists solely to keep them agreeing with each other.

It also persists `Rank`, which is not derivable-but-cached so much as simply misplaced: `GetScoreboard` computes ranks in a Go loop after the query ([repository.go:103-105](db/repository.go#L103-L105)), while `DB.Save(&team)` writes `Rank = 0` back to the row on every submission. A value computed at read time does not belong in a table at all.

For a 100-team contest, a single `GROUP BY` over `team_problem_stats` produces the whole scoreboard in well under a millisecond. **Deleting the `ScoreBoard` table would remove a table, a model, the `Rank` incoherence, and the most delicate function in the repository** — roughly 60 lines net, replaced by one query. This is the largest single simplification available anywhere in the codebase.

If the cache stays, drop the `Rank` column at minimum.

### 3.3 🟠 The WebSocket hub has no per-connection type

[ws.go:13-39](api/ws/ws.go#L13-L39), [scoreboard.go:44-71](api/handlers/scoreboard.go#L44-L71)

Connection state is a bare `map[*websocket.Conn]bool` plus a package-level mutex, and the lifecycle is split across two packages: `api/handlers` registers and deregisters ([scoreboard.go:44-46](api/handlers/scoreboard.go#L44-L46), [scoreboard.go:66-68](api/handlers/scoreboard.go#L66-L68)), `api/ws` iterates and evicts ([ws.go:31-38](api/ws/ws.go#L31-L38)). Two packages share responsibility for one object's lifetime, with the mutex exported so both can reach it.

The conventional gorilla structure gives each connection a type that owns its own writes:

```go
type client struct {
    conn *websocket.Conn
    send chan any        // buffered
}

func (c *client) writePump() {
    defer c.conn.Close()
    for msg := range c.send {
        c.conn.SetWriteDeadline(time.Now().Add(writeDeadline))
        if err := c.conn.WriteJSON(msg); err != nil {
            return
        }
    }
}

type Hub struct {
    mu      sync.RWMutex          // unexported
    clients map[*client]struct{}
}

func (h *Hub) Register(conn *websocket.Conn) { … }
func (h *Hub) Broadcast(msg any)             { … }   // non-blocking sends
```

`ws.Mu` and `ws.Clients` stop being exported, the handler's only job becomes `hub.Register(conn)`, and broadcast becomes a channel send rather than a network write under a shared lock.

### 3.4 🟡 `main` is doing dependency wiring through globals

[main.go:11-28](main.go#L11-L28) is, in effect, already a composition root — it initializes the runtime, the DB, the contest, the manifest, the problems, the queue, and the workers, in dependency order. That instinct is right. The implementation assigns to package globals rather than constructing and passing values:

```go
db.InitDB()                                    // assigns db.DB
models.ActiveContest = db.LoadOrInitContest()  // assigns a global
judge.ActiveManifest = manifest                // assigns a global
judge.InitializeQueue()                        // assigns judge.SubmissionQueue
judge.StartWorkers(3)                          // reads all of the above implicitly
api.Server()                                   // reads all of the above implicitly
```

The dependency *order* is explicit; the dependency *edges* are invisible. `StartWorkers(3)` and `Server()` take no arguments, so nothing in their signatures reveals that both need the DB, the queue, and the contest. Converting these to constructor calls that return values and take dependencies as parameters makes the graph legible at the one place it should be legible — and is a prerequisite for §12.

---

## 4. Architecture: The Missing Service Layer

### 4.1 The current shape

```
HTTP request
     ↓
handlers/*.go  ── parse JSON, check method, enforce rules,
                  run GORM queries, write files, push to queue,
                  format response
     ↓
db.DB (global *gorm.DB)
```

Two layers where there should be four. `api/handlers` imports `db`, `judge`, `models`, `utils`, `os`, `filepath`, and `encoding/json` — the handler package touches everything in the tree.

`SubmitHandler` ([submission.go:14-58](api/handlers/submission.go#L14-L58)) is the clearest illustration. In 44 lines it: enforces the HTTP method, decodes JSON, reads auth context, inserts a DB row, computes a filesystem path, creates a directory, looks up a language spec, writes a source file, nil-checks a package-level channel, pushes to it, and encodes a response. Six distinct responsibilities, none of them separable.

### 4.2 The target shape

Following the entity/service separation this repo's `/refactor` skill describes:

```
handlers/     HTTP only — decode, auth, call service, encode
     ↓
service/      orchestration, enforcement, transactions, side effects
     ↓
store/        data access + business rules as predicates
     ↓
gorm / filesystem / docker
```

The key discipline: **entities define rules as predicates that return values; services enforce them by returning errors.**

```go
// store/team.go — defines the rule, never errors on it
func (t Team) CanSubmit() (bool, string) {
    if t.Status == nil { return true, "" }
    switch *t.Status {
    case StatusDisqualified: return false, "team is disqualified"
    case StatusDisabled:     return false, "team account is disabled"
    }
    return true, ""
}

// service/submission.go — enforces it
func (s *SubmissionService) Submit(ctx context.Context, team store.Team, req SubmitRequest) (int, error) {
    if ok, reason := team.CanSubmit(); !ok {
        return 0, apierr.Forbidden(reason)
    }
    …
}
```

### 4.3 🔴 Business rules have no single home, so they are defined more than once

This is the finding that most justifies the whole refactor. Because there is no layer whose job is to state a rule, rules end up restated wherever they are needed.

The team-status rule exists in three copies, and the three do not agree:

| Location | Blocks | Effect |
| --- | --- | --- |
| [middleware.go:41-44](api/handlers/middleware.go#L41-L44) | `disabled` **and** `disqualified` | 403 on authenticated requests |
| [auth.go:35-38](api/handlers/auth.go#L35-L38) | `disqualified` only | 403 at login |
| [repository.go:98](db/repository.go#L98) | `disqualified` only | excluded from scoreboard |

A reader cannot tell whether the divergence is intentional, because the rule is never stated anywhere as a rule — only implemented three times as a condition. That is the structural defect, independent of which behaviour is the desired one.

Other rules with no home:

- **Penalty formula.** `minutes + 20 × wrongAttempts` is a bare expression at [repository.go:75](db/repository.go#L75), *inside a persistence function*. ICPC scoring is domain logic; the `20` is a contest configuration parameter (§11). Neither belongs in the data-access layer.
- **Contest openness.** `Status == Active && now ∈ [start, end]` is expressed only as HTTP middleware ([middleware.go:56-69](api/handlers/middleware.go#L56-L69)). Anything that is not an HTTP request — the judge workers, for instance — has no way to ask the question.
- **Valid team statuses.** The set `{enabled, disabled, disqualified}` is enforced by an inline three-way `&&` at [admin.go:172](api/handlers/admin.go#L172) and compared against string literals in four other files (§10.2).
- **Override state machine.** `pending → approved | rejected` is an `if / else if` chain over raw strings at [admin.go:256-262](api/handlers/admin.go#L256-L262), with no type expressing the state set and no function expressing the legal transitions.

Each of these is one predicate on one domain type. Extracting them is the highest-leverage readability change in this document: it converts "grep for the string `disqualified` and hope you found all of them" into "read `TeamStatus`."

### 4.4 🟠 The repository pattern is applied to about a fifth of the operations

`db/repository.go` has seven functions. Handlers bypass them routinely:

| Operation | Repository function? | What handlers actually do |
| --- | --- | --- |
| Save submission | ✅ | uses it |
| Get submission | ✅ | [submission.go:76](api/handlers/submission.go#L76) uses raw `db.DB.First` anyway |
| Register team | ✅ | uses it |
| List teams | ❌ | [admin.go:157](api/handlers/admin.go#L157) raw `db.DB.Find` |
| Update team status | ❌ | [admin.go:177](api/handlers/admin.go#L177) raw |
| List submissions | ❌ | [admin.go:79](api/handlers/admin.go#L79), [submission.go:104-110](api/handlers/submission.go#L104-L110) raw |
| Session lookup | ❌ | [middleware.go:33](api/handlers/middleware.go#L33), [middleware.go:82](api/handlers/middleware.go#L82) raw |
| Overrides | ❌ | [admin.go:229](api/handlers/admin.go#L229), [admin.go:248](api/handlers/admin.go#L248) raw |
| Problem by ID | ❌ | [problems.go:51](api/handlers/problems.go#L51) raw |
| Update submission verdict/status | ❌ | [queue.go:27-52](judge/queue.go#L27-L52) raw — *from the judge package* |

The comment at [repository.go:34](db/repository.go#L34) — *"i write like this as my first language was c# and it follows this repository thing very strictly"* — has the right instinct. The problem is not the pattern; it is that partial application leaves `db.DB` globally reachable, so the pattern provides no actual seam. A repository that can be bypassed is documentation, not architecture.

**Recommendation:** make `DB` unexported. The compiler then enumerates every violation, and each one becomes a store method. This single change is the gateway to §12 — nothing else in the testability section is reachable without it.

### 4.5 🟡 "Repository" as a function suffix

`RegisterTeamRepository` reads oddly because *Repository* is a **type** name in the C# convention it borrows from, not a function suffix. In Go the package or receiver supplies that namespace:

```go
store.Teams.Register(team)     // not  db.RegisterTeamRepository(res)
store.Submissions.Save(sub)    // not  db.SaveSubmissionToDB(req)
store.Submissions.ByID(id)     // not  db.GetSubmissionByID(id)
```

`SaveSubmissionToDB`'s `ToDB` suffix is likewise redundant — a store method saving somewhere *other* than the DB would be the case worth naming.

---

## 5. Layering Violations & Dependency Direction

### 5.1 🔴 `judge` imports `api/ws` — the domain depends on the transport

[queue.go:4](judge/queue.go#L4) imports `campuscode/api/ws`; [queue.go:49](judge/queue.go#L49) calls `ws.BroadcastScoreboard()`.

The judge is the core domain and should be the *most* independent package in the tree. Instead it reaches upward into the WebSocket transport. Consequences: the judge cannot be compiled or exercised without the WebSocket hub; it cannot be reused headlessly; and the dependency graph reads backwards to anyone new to the codebase.

**Fix — invert with an interface owned by the consumer:**

```go
// judge/worker.go
type ScoreboardNotifier interface {
    ScoreboardChanged()
}

type Worker struct {
    store    SubmissionStore
    notifier ScoreboardNotifier
}
```

`main` wires `ws.Hub` in as the notifier. `judge` then depends on nothing beyond `models` and the standard library, and tests pass a no-op notifier.

### 5.2 🟠 `judge` writes to the database

[fileSystem.go:99-132](judge/fileSystem.go#L99-L132) — `LoadOrInitProblems` imports `campuscode/db` and issues `db.DB.First` and `db.DB.Create` from inside the judge package. Problem seeding is a bootstrap concern, not a judging one.

Split it: `judge.LoadManifest()` parses the file and returns data; `store.Problems.SyncFromManifest(m)` persists it. Each half becomes independently testable, and `judge` sheds its dependency on `db`.

Note that the file this lives in is named `fileSystem.go` — a file about paths that also contains manifest parsing, two DTO types, and database writes (§13.2).

### 5.3 🟠 `api/handlers` reaches into the filesystem directly

Three handlers do their own file I/O against judge-owned paths:

- [submission.go:34-45](api/handlers/submission.go#L34-L45) — `MkdirAll` + `WriteFile` of source into `judge.SubmissionPathByID(...)`
- [submission.go:87-92](api/handlers/submission.go#L87-L92) — reads it back
- [admin.go:94-110](api/handlers/admin.go#L94-L110) — reads source and `compile.log`

The HTTP layer therefore knows the judge's on-disk layout, the source-file naming convention, and the `compile.log` filename. Three places to update if the layout ever changes, and three places that must independently agree on a convention nothing enforces.

`judge` should own its storage and expose an artifact API:

```go
func (s *Store) SaveSource(subID int, lang, code string) error
func (s *Store) LoadSource(subID int) (string, error)
func (s *Store) LoadCompileLog(subID int) (string, error)
```

### 5.4 🟠 `problems.go` reconstructs a path the judge already computes

[problems.go:45](api/handlers/problems.go#L45):

```go
problemPath := filepath.Join("server_data", "contest", "problems", id)
```

`judge/fileSystem.go` already defines `basePath`, `contestPath`, and `TestsDir(id)`. This handler rebuilds the same path from string literals, one directory shallower. Two definitions of the layout, neither aware of the other. Add `judge.ProblemDir(id)` and call it.

### 5.5 🟡 The `FileSystem` struct is a stub with an undecided design memo attached

[fileSystem.go:18-48](judge/fileSystem.go#L18-L48) constructs a `FileSystem{}` inside `InitializeRunTime`, uses it to build three paths, and discards it. The three path helpers below it are plain functions closing over package-level constants, ignoring the struct entirely. A 10-line commented block ([fileSystem.go:38-48](judge/fileSystem.go#L38-L48)) lays out both options and reaches no conclusion.

Pick one. The comment already identifies the deciding factor correctly: *"you can build a `FileSystem{Runtime: "/tmp/test_data/runtime"}` in a test and get isolated paths, no shared-state test pollution."* Given §12, methods on the struct are the right call — but a stub plus an unresolved memo is the one option that costs without paying.

---

## 7. The Judge Subsystem: Structural Findings

*(Judge behavioural defects are in the Correctness Audit. What follows concerns the shape of the code.)*

### 7.1 🟠 `compile` and `runAdHocCompile` are near-duplicates

[judge.go:100-140](judge/judge.go#L100-L140) and [run.go:63-93](judge/run.go#L63-L93) construct the same 11-argument `docker run` invocation with the same timeout and the same buffer handling. They differ only in whether the captured log is written to `compile.log` or returned as a string. The comment at [run.go:60-62](judge/run.go#L60-L62) honestly documents the duplication rather than resolving it.

One function returning `(log string, err error)` serves both; the submission path additionally persists the log. That removes ~30 duplicated lines — and, more importantly, means any future change to the sandbox invocation happens in one place rather than two that must be kept in sync by hand. The two copies have **already drifted**: the flag sets are not identical between them.

### 7.2 🟠 `runDocker`'s signature does not scale

[docker.go:12](judge/docker.go#L12):

```go
func runDocker(submissionID int, workspacePath string, command string,
               timeoout time.Duration, spec LanguageSpec) (string, string, error)
```

Five positional parameters, of which `submissionID` is never read in the body — call sites pass `0` for it ([run.go:46](judge/run.go#L46)). `timeoout` is misspelled. Resource limits (`--memory`, `--cpus`, `--pids-limit`) are hardcoded in the body ([docker.go:23-25](judge/docker.go#L23-L25)) rather than parameterized, so there is no path through this signature for per-problem configuration to reach the sandbox — even though `models.Problem` models exactly that (`TimeLimit`, `MemoryLimit`) and the comment at [docker.go:13](judge/docker.go#L13) says that is the intent.

An options struct fixes the shape and opens the path:

```go
type runOpts struct {
    Workspace string
    Command   string
    Timeout   time.Duration
    MemoryMB  int
    CPUs      float64
    PidsLimit int
    Spec      LanguageSpec
}

func runDocker(ctx context.Context, o runOpts) (stdout, stderr string, err error)
```

Adding `ctx` as the first parameter also matches convention — `runDocker` currently constructs its own `context.Background()` internally, which means callers cannot cancel it (relevant to graceful shutdown, and to §12.2's fake runner).

### 7.3 🟠 Verdicts are `string` despite a perfectly good enum existing

`models.SubmissionVerdict` is defined at [enums.go:3-11](models/enums.go#L3-L11) with all six values. The judge then round-trips through `string` at every boundary:

```go
// judge.go — returns string
return string(models.InternalError), fmt.Errorf(…)
return string(models.Accepted), nil

// queue.go:42 — converts back
sub.Verdict = models.SubmissionVerdict(result)
```

`processSubmission`, `runAllTestCases`, and `runTestCase` all have signature `(string, error)`. Changing all three to return `models.SubmissionVerdict` deletes about a dozen conversions and makes it a compile error to return something outside the set. As written, `return "AC", nil` and `return "Accepted", nil` are equally valid Go — the type exists but is not doing any work.

### 7.4 🟠 `runTestCase` discards structured information to return a string

[judge.go:145-176](judge/judge.go#L145-L176)

At the moment it decides a verdict, `runTestCase` holds the test name, the expected output, the actual output, the stderr, and (implicitly) the elapsed time. It returns a single string, so all of it is dropped at the boundary.

The comment at [judge.go:91](judge/judge.go#L91) shows this was noticed — *"store failure.json in submission workspace with details of the failed test case for frontend to display"* — but the return type makes it awkward, so it was never done. The type is the obstacle:

```go
type TestResult struct {
    Verdict  models.SubmissionVerdict
    TestName string
    Expected string        // truncated; retained only for a failing case
    Actual   string
    Stderr   string
    Duration time.Duration
}
```

This is the natural home for the per-submission diagnostics the admin panel and the participant UI both want, and for the execution timing that nothing currently measures.

### 7.5 🟡 Shell commands are assembled by string concatenation at call sites

[judge.go:156](judge/judge.go#L156):

```go
runDocker(submissionID, workspacePath, spec.RunCmd+"< /workspace/input.txt", …)
```

The call site is doing string surgery on a shell command — and doing it slightly differently from [run.go:46](judge/run.go#L46), which is the same expression written out a second time. The redirect is an invariant of how this judge feeds input to a program; it belongs in `LanguageSpec` or inside `runDocker`, stated once, not appended by each caller.

(There is a missing space before `<` in both copies. Noted in the Correctness Audit.)

### 7.6 🟡 `judge/fileSystem.go` is four unrelated files in one

143 lines containing: runtime directory initialization, three path helpers, a package-level `*Manifest` global, manifest JSON parsing, two manifest DTO types, a `ProblemMeta` DTO, database-writing problem seeding (§5.2), and an unused `FileSystem` struct (§5.5).

Split along the seams that are already visible: `paths.go`, `manifest.go`, `storage.go`.

### 7.7 🟡 `NormalizeOutput` lives in `utils`

[util.go:73-80](utils/util.go#L73-L80) defines what "matching output" means for the entire platform — CRLF normalized, per-line trailing whitespace stripped, leading/trailing blank lines ignored. That is the single most consequential rule in a judge, and it is filed under `utils` next to a file-copy helper.

It belongs in `judge`, adjacent to the comparison it serves, where someone auditing judging fairness will find it. See §10.4.

---

## 8. Persistence Layer: Structural Findings

*(Data-integrity and transaction findings are in the Correctness Audit.)*

### 8.1 🟠 `InitDB` panics instead of returning an error

[db.go:13-22](db/db.go#L13-L22):

```go
func InitDB() {
    db, err := gorm.Open(...)
    if err != nil { panic(err) }
    log.Println("DB connection established")
    db.AutoMigrate(&models.Submission{}, ...)   // error discarded
    DB = db
}
```

Two structural problems. First, `panic` in a library function removes `main`'s ability to decide how to fail — and `main` already handles errors properly for the other two initialization steps ([main.go:13-25](main.go#L13-L25)), so the codebase is inconsistent with itself. Second, the function returns nothing and assigns to a global, so it cannot be used to construct a second database (a test fixture, say) even in principle.

```go
func Open(cfg config.Config) (*Store, error)
```

Returning a `*Store` rather than assigning `DB` is the same change §4.4 asks for, arriving from the other direction.

### 8.2 🟠 `UpdateTeamProblemStatsAndScoreboard` has no error return, so errors became log lines

[repository.go:48](db/repository.go#L48) — signature is `func(sub, startTime)`, returning nothing. Consequently:

```go
res := DB.Save(&stats)
log.Println(res.Error)          // prints "<nil>" on every submission
log.Println(res.RowsAffected)
result := DB.Save(&team)
log.Println(result.Error)
log.Println(result.RowsAffected)
```

These are error checks that had nowhere to go, so they were written as prints. Four lines of noise per submission in a log an organizer may be watching live, with `<nil>` rendered as a bare line.

The fix is the signature, not the prints: add `error`, propagate, and let the caller in `judge` decide. The function is also doing three separate jobs — loading stats, applying ICPC scoring rules, and persisting two tables — which is why it is the longest function in `db/`. §3.2 proposes removing most of it outright; failing that, the scoring calculation should be a pure function in `domain`, testable without a database.

### 8.3 🟡 `FirstOrCreate` relies on implicit GORM condition-building

[repository.go:50-57](db/repository.go#L50-L57):

```go
DB.FirstOrCreate(&stats, models.TeamProblemStats{
    TeamID:            sub.TeamID,
    ProblemID:         sub.ProblemID,
    WrongAttemptCount: 0,
    Solved:            false,
    FirstAcTime:       nil,
    Penalty:           nil,
})
```

GORM builds the `WHERE` clause from the struct's **non-zero** fields, so the last four are silently dropped and the lookup is `team_id = ? AND problem_id = ?`.

A reader has to know that rule to understand what this queries — and the four zero-valued fields actively suggest they are part of the condition when they are not. Whether or not the behaviour is what was wanted, the code does not say what it does. Pass only the key fields as conditions, and express defaults with `Attrs()`, where they read as defaults.

### 8.4 🟡 Missing indexes on routinely-queried columns

`Sessions.TeamID` is indexed and `SessionToken` is the primary key. Not indexed:

- `Submission.TeamID` and `Submission.ProblemID` — both filtered on every `/submissions/me` ([submission.go:104-107](api/handlers/submission.go#L104-L107))
- `Teams.Status` — joined and filtered on every scoreboard read ([repository.go:97-98](db/repository.go#L97-L98))
- `DeviceOverride.Status` — filtered by the admin panel

At 100 teams these are trivial scans. The argument for adding them is consistency and intent-documentation — a `gorm:"index"` tag records which columns are query keys, which is information the next reader otherwise has to reconstruct by grepping.

### 8.5 🟡 `models` is a persistence-schema package and a DTO package at once

`models/models.go` holds GORM entities; `models/dto.go` holds HTTP request/response shapes; `models/enums.go` holds domain enums used by both. `models.ActiveContest` — a runtime global — is declared at the top of the entity file ([models.go:6](models/models.go#L6)).

The result is that `models` is imported by every package in the tree, so every package transitively depends on both the database schema and the HTTP wire format. Separating `domain/` (pure types and rules), `store/` (persistence), and `transport/http/dto.go` (wire shapes) breaks that — and lets the wire format change without touching the schema, which is the point of having both.

---

## 9. HTTP Layer: Routing, Contracts, Consistency

### 9.1 🟠 Twenty-one hand-written method checks the router could do

Every handler opens with a variant of:

```go
if r.Method != http.MethodGet {
    http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
    return
}
```

There are 21, using **six different message strings**: `"Wrong method is called"`, `"Method Not Allowed"`, `"Invalid method called"`, `"Wrong method called in login Handler"`, `"Wrong method is called , must call POST"`, `"invalid method is called for scoreboard"`.

Go 1.22's `ServeMux` does this — and [server.go:57-58](api/server.go#L57-L58) already uses the feature, with a comment recording the discovery:

```go
http.HandleFunc("GET /admin/contest", …)
http.HandleFunc("PATCH /admin/contest", …)
```

Applying it everywhere deletes ~84 lines, yields one consistent 405, and turns the route table into documentation:

```go
mux.HandleFunc("POST /submit",          requireAuth(requireContestOpen(h.Submit)))
mux.HandleFunc("GET /submission/{id}",  requireAuth(h.GetSubmission))
mux.HandleFunc("GET /problems",         requireAuth(requireContestOpen(h.ListProblems)))
```

### 9.2 🟠 `http.DefaultServeMux` prevents isolation

[server.go:26-58](api/server.go#L26-L58) registers on the package-global default mux. Route registration is therefore a global side effect: `Server()` can only be called once per process (a second call panics on duplicate patterns), and no test can construct an isolated router.

```go
func NewRouter(deps Deps) *http.ServeMux
```

`httptest.NewServer(NewRouter(deps))` then works, and §12 becomes reachable.

Related: [server.go:63](api/server.go#L63) uses `http.ListenAndServe`, which gives no handle on the server. An explicit `&http.Server{...}` is needed for timeouts (§11) and for any shutdown handling.

### 9.3 🟠 Two response formats for one API

Success is JSON via `utils.WriteJSON`. Errors are `text/plain` via `http.Error`. Every frontend call has to branch on status and parse differently, and error text is unstructured — so a client cannot distinguish error *kinds* except by matching English sentences.

That distinction is load-bearing today: [auth.go:79](api/handlers/auth.go#L79) returns `"Device mismatch. Request Override validated by Admin"` with a 401, which the client must handle completely differently from a credentials failure that returns the same status. The only discriminator on the wire is the prose.

```go
type APIError struct {
    Code    string `json:"code"`     // "device_mismatch", "contest_closed", …
    Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, status int, code, msg string) {
    WriteJSON(w, status, APIError{Code: code, Message: msg})
}
```

A stable `code` is also what lets error strings be reworded (§13.4) without breaking the client.

### 9.4 🟠 Handlers are functions over globals rather than methods on a type

Every handler is a bare `func(w, r)` reaching for `db.DB`, `judge.SubmissionQueue`, `judge.LanguageSpecs`, and `models.ActiveContest` at the point of use. There is no struct holding what a handler needs, so there is no signature anywhere that documents a handler's dependencies and no way to substitute them.

```go
type Handlers struct {
    submissions *service.SubmissionService
    contest     *service.ContestState
    scoreboard  *service.ScoreboardService
}

func (h *Handlers) Submit(w http.ResponseWriter, r *http.Request) { … }
```

This is the same change as §3.4 and §4.4 seen from the transport side; all three are one refactor.

### 9.5 🟡 Inconsistent route naming and verb conventions

```
/submission/{id}      singular
/submissions/me       plural
/admin/submissions    plural
/admin/teams/{id}     plural
/admin/overrides/{id} plural
```

`/submission/{id}` is the outlier. Separately, two endpoints express "change one field on one resource" two different ways: `PATCH /admin/teams/{id}` takes `{"status": …}`, while `POST /admin/overrides/{id}` takes `{"action": "approve"|"reject"}`.

### 9.6 🟡 Middleware coverage is inconsistent and undocumented

| Route | Auth | Contest-open |
| --- | --- | --- |
| `/submit` | ✅ | ✅ |
| `/problems` | ✅ | ✅ |
| `/problems/{id}` | ✅ | ✅ |
| `/run` | ✅ | ❌ |
| `/submission/{id}` | ✅ | ❌ |
| `/submissions/me` | ✅ | ❌ |
| `/contest` | ✅ | ❌ |
| `/scoreboard` | ❌ | ❌ |

Some of these gaps are clearly deliberate — `/contest` must be readable before the contest opens. Others are not self-evidently either way, and nothing in the code says which is which. A route table with an explicit middleware column (§9.1's method-scoped registration makes this natural) turns the policy into something reviewable at a glance.

### 9.7 🟡 The two auth middlewares are near-duplicates

[middleware.go:23-49](api/handlers/middleware.go#L23-L49) and [middleware.go:72-93](api/handlers/middleware.go#L72-L93) share ~20 lines — header parsing, `Bearer` trimming, empty-token check, session lookup, team load, context injection — differing only in the `WHERE` clause and a status check. Two copies of an auth path is two places to update, and divergence between them is invisible at a glance.

One `authenticate(r *http.Request, requireAdmin bool) (Team, error)` helper, with the two middlewares as thin wrappers.

### 9.8 🟡 `LoginHandler` uses a closure that writes HTTP responses

[auth.go:45-66](api/handlers/auth.go#L45-L66) defines a `login := func() { … }` closure that calls `http.Error` and `return`s — but that `return` exits the *closure*, not the handler. Control flow at the call sites then depends on the caller also returning immediately.

It happens to be written correctly at both call sites, but a helper whose HTTP-writing control flow does not propagate is a trap for the next edit. Make it a method returning `(token string, err error)` and let the handler own all response writing.

The surrounding `if / else if / else if` chain ([auth.go:68-85](api/handlers/auth.go#L68-L85)) also re-tests `team.BoundDeviceID != nil` in branches 2 and 3, which branch 1 already established, and has no `else`. A `switch` with a `default` states the same logic without the redundancy:

```go
switch {
case team.BoundDeviceID == nil:        // first login — bind
case *team.BoundDeviceID == deviceID:  // same device
default:                                // different device — request override
}
```

### 9.9 🟡 `AdminGetScoreboard` delegates by calling another handler

[admin.go:220](api/handlers/admin.go#L220) is `ScoreboardHandler(w, r)`. Handler-calls-handler means the method check runs twice and any handler-specific behaviour is inherited implicitly — `ScoreboardHandler` already does WebSocket upgrade negotiation, which the admin route silently acquires. Both should call a shared service function instead.

### 9.10 🟡 `LoggingMiddleware` logs on the way in only

[middleware.go:13-18](api/handlers/middleware.go#L13-L18) logs `--> METHOD path` and nothing more. No status, no duration, no request ID. For a server whose primary operational interface is the organizer's terminal, capturing the outcome is where the value is:

```go
sw := &statusWriter{ResponseWriter: w, status: 200}
start := time.Now()
next.ServeHTTP(sw, r)
log.Printf("%s %s → %d (%s)", r.Method, r.URL.Path, sw.status, time.Since(start))
```

---

## 10. Types, Naming & Domain Modelling

### 10.1 🟠 Plural type names for singular entities

```go
type Teams struct { … }      // one team
type Sessions struct { … }   // one session
```

Every use site reads wrong: `var team models.Teams`, `func TeamFromContext(r) models.Teams`. `Submission`, `Contest`, `Problem`, and `DeviceOverride` are already singular, so the codebase is inconsistent with itself as well as with convention.

GORM derives table names by pluralizing, so `Team` → `teams` produces the **same table name** you have now. This rename is free of migration cost.

### 10.2 🟠 Three status concepts, all stringly-typed

| Field | Values | Compared in |
| --- | --- | --- |
| `Teams.Status *string` | `enabled` / `disabled` / `disqualified` | middleware.go:41, auth.go:36, admin.go:172, repository.go:98 |
| `Submission.Status string` | `Queued` / `Running` / `Done` | repository.go:15, queue.go:29, queue.go:52, admin.go:138 |
| `DeviceOverride.Status string` | `pending` / `approved` / `rejected` | auth.go:78, admin.go:252, admin.go:257, admin.go:261 |

`SubmissionVerdict` and `ContestStatus` in [enums.go](models/enums.go) show the right pattern already exists here — it simply was not applied to these three. Costs of that:

- `Teams.Status` is a `*string`, so "no status" and `"enabled"` are two encodings of one state, and every read site must nil-check *and* compare.
- A typo like `"Disqualified"` compiles and silently never matches.
- `Submission.Status` uses `Title Case` while every other status in the system is lowercase — an inconsistency with no reason behind it.
- The valid set is not discoverable from the type; you find it by grepping string literals across four files.

```go
type TeamStatus string

const (
    TeamEnabled      TeamStatus = "enabled"
    TeamDisabled     TeamStatus = "disabled"
    TeamDisqualified TeamStatus = "disqualified"
)

func (s TeamStatus) Valid() bool     { … }
func (s TeamStatus) CanSubmit() bool { return s == TeamEnabled }
```

Making it non-pointer with `enabled` as the zero value also deletes a category of nil-check — and `CanSubmit()` is precisely the missing home from §4.3.

### 10.3 🟠 Three overlapping `Problem` types with inconsistent ID types

| Type | ID type | Fields | Role |
| --- | --- | --- | --- |
| `models.Problem` | `int` | Title, TimeLimit, MemoryLimit, SolvedCount | DB row |
| `judge.Problem` | `string` | ID, Path, Title | manifest entry |
| `models.ProblemResponse` | `string` | Title, TimeLimit, MemoryLimit, Description | API response |

Two are named `Problem` in different packages with **incompatible ID types**, forcing a `strconv.Atoi` at the boundary ([fileSystem.go:102](judge/fileSystem.go#L102)). And because `ProblemResponse.ID` is a string while `Problem.ID` is an int, `GET /problems` serializes `"id": 3` while `GET /problems/3` serializes `"id": "3"` — the same field, two JSON types, depending on the endpoint.

Rename `judge.Problem` → `judge.ManifestProblem` and make all three IDs `int`. If the frontend is being built against these endpoints now, pull this one forward.

Related: `models.Problem.SolvedCount` is declared, migrated, and serialized to the frontend, but nothing anywhere writes to it.

### 10.4 🟠 `utils` is a grab-bag

[utils/util.go](utils/util.go) contains a crypto token generator, a file copier, a directory lister, an HTTP JSON writer, a string normalizer, and a markdown reader — six unrelated concerns in one 86-line file named `util`.

A package named `utils` tells a reader nothing and acts as a magnet: every subsequent "where does this go?" resolves to `utils`, and it grows monotonically. Split by concern:

```
internal/httpx/     WriteJSON, WriteError, DecodeJSON
internal/token/     Generate
internal/fsx/       CopyFile, ListInputFiles, ReadProblemStatement
judge/              NormalizeOutput   ← domain logic, not a utility (§7.7)
```

### 10.5 🟡 DTOs carry fields the handlers deliberately ignore

`SubmitRequest.TeamID` ([dto.go:6](models/dto.go#L6)) is decoded from the request body and then immediately overwritten from the auth context at [submission.go:27](api/handlers/submission.go#L27). `LoginRequest.BoundDeviceID` ([dto.go:31](models/dto.go#L31)) is declared in the body type but never read — the handler takes the device from the `X-Device-ID` header instead ([auth.go:40](api/handlers/auth.go#L40)).

In both cases the wire type advertises a field that is not part of the contract. Remove them: a DTO should describe what the endpoint actually accepts, and a field that exists only to be discarded invites a future reader to use it.

`RegisterRequest.Members` ([dto.go:20](models/dto.go#L20)) is decoded and never used either — the corresponding `Teams.Members` model field is commented out (§13.2).

### 10.6 🟡 `WriteJSON` discards its encoder error

[util.go:67-71](utils/util.go#L67-L71) ignores the return of `json.NewEncoder(w).Encode(a)`. Since headers are already written there is nothing to *send* on failure — but there is something to *log*, and a silently truncated response is unpleasant to diagnose from the client end.

### 10.7 🟡 `SubmissionDetail` — a good pattern, in the wrong file

[submission.go:63-66](api/handlers/submission.go#L63-L66) embeds `models.Submission` to flatten the JSON, with a clear comment explaining why. The technique is right. It is noted only as evidence for a dedicated `transport/http/dto.go`: response shapes assembled from domain models belong together, not scattered through handler files where they are found by accident.

---

## 11. Configuration & Magic Numbers

Not one tunable is configurable. Every value below is a literal in source:

| Value | Location | Should be |
| --- | --- | --- |
| `3` workers | [main.go:27](main.go#L27) | `JUDGE_WORKERS`, default `runtime.NumCPU()` |
| `100` queue depth | [queue.go:13](judge/queue.go#L13) | config |
| `3` concurrent ad-hoc runs | [run.go:11](api/handlers/run.go#L11) | config |
| `:8080` | [server.go:63](api/server.go#L63) | `PORT` |
| `6 * time.Hour` session TTL | [auth.go:57](api/handlers/auth.go#L57), [admin.go:64](api/handlers/admin.go#L64) | config — currently duplicated |
| `20` minute ICPC penalty | [repository.go:75](db/repository.go#L75) | contest config (manifest) |
| `10s` run timeout | [judge.go:156](judge/judge.go#L156), [run.go:46](judge/run.go#L46) | per-problem (§7.2) |
| `30s` compile timeout | [judge.go:110](judge/judge.go#L110), [run.go:68](judge/run.go#L68) | config — duplicated |
| `--memory=256m` (run) | [docker.go:23](judge/docker.go#L23) | per-problem (§7.2) |
| `--cpus=0.5` (run) | [docker.go:24](judge/docker.go#L24) | config |
| `--pids-limit=64` | [docker.go:25](judge/docker.go#L25) | config |
| `--memory=512m --cpus=1` (compile) | [judge.go:115-116](judge/judge.go#L115-L116), [run.go:74-75](judge/run.go#L74-L75) | config — duplicated |
| `60s` time-sync interval | [ws.go:60](api/ws/ws.go#L60) | config |
| `5s` WS write deadline | [ws.go:25](api/ws/ws.go#L25) | config |
| `1024` WS buffer sizes | [scoreboard.go:17-18](api/handlers/scoreboard.go#L17-L18) | fine as literals |
| `server_data/runtime/db.sqlite` | [db.go:14](db/db.go#L14) | config |
| `server_data` tree | [fileSystem.go:14-16](judge/fileSystem.go#L14-L16) | config |
| Docker image tags | [judge.go:24-42](judge/judge.go#L24-L42) | config; `gcc:latest` is unpinned |

Three points worth drawing out:

**The 20-minute penalty is a contest rule, not a constant.** Different organizers score differently. It belongs in `manifest.json` alongside `duration_minutes`, which is already there — and in `domain`, not in `db/repository.go` (§4.3).

**`gcc:latest` is not pinned**, while `python:3.12-slim` and `eclipse-temurin:21-jdk` are. For an offline-first product distributed as a package, the C++ toolchain therefore varies by whenever the organizer last pulled. This is an inconsistency within the same map literal.

**Several values are duplicated** rather than merely hardcoded — the session TTL, the compile timeout, and the compile resource limits each appear in two files. Those will drift; the compile flags already have (§7.1).

**Suggested shape** — one struct, populated from env with documented defaults, passed down from `main`:

```go
type Config struct {
    Port           string
    DataDir        string
    JudgeWorkers   int
    QueueDepth     int
    SessionTTL     time.Duration
    CompileTimeout time.Duration
    PenaltyMinutes int
}

func Load() Config   // env → struct
```

Passing `cfg` explicitly rather than reading env at point-of-use is what lets a test run against a temp directory with a one-second timeout.

---

## 12. Testability

**There are zero test files.** That is a consequence of the structure, not of neglect — most of this code cannot be tested as written.

### 12.1 What blocks testing today

| Blocker | Effect | Ref |
| --- | --- | --- |
| `db.DB` package global | Any handler test needs a real SQLite file on disk | §4.4 |
| `models.ActiveContest` global | Tests share state; cannot run in parallel | §3.1 |
| `judge.SubmissionQueue` global | Cannot inject a fake queue | §3.1 |
| `judge.ActiveManifest` global | Admin login untestable without a manifest file | §3.1 |
| `http.DefaultServeMux` | Cannot build an isolated router | §9.2 |
| Handlers call `db.DB` directly | No seam to substitute a fake store | §4.4 |
| Handlers are bare funcs | Nowhere to inject dependencies | §9.4 |
| `judge` imports `api/ws` | Judge tests drag in the WebSocket hub | §5.1 |
| Hardcoded `server_data` paths | Tests write into the real data directory | §11 |
| Docker invoked directly | Every judge test needs a Docker daemon | §7.2 |
| Scoring logic inside a repository function | Cannot test ICPC rules without a DB | §8.2 |

Note that every row is a structural finding from an earlier section. Testability is not a separate workstream — it is the payoff from §§3–5.

### 12.2 The unlock sequence

Each step makes the next cheap:

1. **Unexport `db.DB`.** The compiler enumerates all ~15 violations; each becomes a store method. Now there is a seam.
2. **Define narrow interfaces at the consumer.** Not one fat `Store` — small interfaces where they are used, per Go convention:
   ```go
   // in judge
   type SubmissionStore interface {
       ByID(id int) (models.Submission, error)
       SetStatus(id int, status models.SubmissionStatus) error
       SetVerdict(id int, v models.SubmissionVerdict) error
   }
   ```
   A fake is then ~15 lines, and `judge` stops importing `db` and `gorm` entirely.
3. **Wire dependencies in `main`** (§3.4). Handlers become methods on a struct (§9.4).
4. **`NewRouter() http.Handler`** (§9.2). Enables `httptest.NewServer`.
5. **Extract a `Runner` interface for Docker** (§7.2):
   ```go
   type Runner interface {
       Run(ctx context.Context, o runOpts) (stdout, stderr string, err error)
   }
   ```
   Verdict-aggregation logic becomes testable with a fake runner and no daemon.
6. **Extract pure functions** — penalty calculation out of `UpdateTeamProblemStatsAndScoreboard` (§8.2), contest-openness out of middleware (§4.3), output normalization into `judge` (§7.7). Each is then testable with no infrastructure at all.

### 12.3 Where the first tests should go

Chosen for coverage-per-line-written, favouring things that are pure after modest extraction:

1. **`NormalizeOutput`** — CRLF, trailing newline, per-line trailing spaces, empty input, whitespace-only. Pure today; five minutes to write; defines contest fairness.
2. **Penalty calculation** — first-try AC, N wrong attempts then AC, submissions after AC. Pure once extracted per §8.2.
3. **Contest-open predicate** — before start, during, after end, across each status value. Pure once extracted per §3.1.
4. **`TeamStatus.CanSubmit()`** — pure once §10.2 lands; this is the test that would keep §4.3's three definitions from re-diverging.
5. **Verdict aggregation** (`runAllTestCases`) with a fake runner — table-driven over test-result sequences. Needs step 5 of §12.2.
6. **Scoreboard ordering** — solved DESC, penalty ASC, lastAC ASC, with ties on each field. Needs a store seam.
7. **Device-binding state machine** — first login binds; same device proceeds; different device produces a pending override; approved override rebinds. Needs steps 1–4.

Items 1–4 need no infrastructure whatsoever and can be written before any restructuring lands.

### 12.4 Enable `-race` alongside the first test

Once any test exercises a worker and a handler together, add `-race` to the test command and to whatever CI exists. Even ahead of that, running the dev server as `go run -race main.go` during `dev_reset.sh` costs nothing and exercises the real concurrency paths.

---

## 13. Hygiene: Formatting, Dead Code, Dependencies, Comments

### 13.1 🔴 Every file fails `gofmt`

`gofmt -l .` lists **23 of 23** files. Concretely: missing spaces after commas in argument lists, C-style `if(cond)` parentheses ([repository.go:66](db/repository.go#L66), [repository.go:71](db/repository.go#L71), [auth.go:68](api/handlers/auth.go#L68)), inconsistent struct-tag alignment, mixed indentation — [judge.go:24-42](judge/judge.go#L24-L42) uses spaces while the rest of the tree uses tabs — and orphaned blank lines inside import blocks ([util.go:3-18](utils/util.go#L3-L18)).

In Go this is not a preference. `gofmt` is canonical, every editor runs it on save, and every reviewer expects it. One command:

```bash
gofmt -w .
```

Worth a pre-commit hook or a CI check so it stays fixed. This is the single cheapest readability win available and should be done before any other diff, so that later diffs are reviewable.

### 13.2 🟡 Dead code and dead files

| Item | Location |
| --- | --- |
| `api/handlers/contest.go` — contains only `package handlers` | [contest.go](api/handlers/contest.go) |
| `ContestHandler` lives in `problems.go`, not `contest.go` | [problems.go:25](api/handlers/problems.go#L25) |
| `LogoutHandler` route commented out | [server.go:36](api/server.go#L36) |
| `//"campuscode/api/middleware"` — package does not exist | [server.go:6](api/server.go#L6) |
| `//"campuscode/utils"` — duplicate of a live import | [auth.go:9](api/handlers/auth.go#L9) |
| `Submission.Code` commented out | [models.go:12](models/models.go#L12) |
| `Teams.Members`, `Teams.CreatedAt` commented out | [models.go:22-24](models/models.go#L22-L24) |
| `TeamProblemStats` — 3 lines of superseded field definitions | [models.go:62-64](models/models.go#L62-L64) |
| `Sessions.ID` commented out | [models.go:30](models/models.go#L30) |
| `compile()` — 5 lines of commented-out path building | [judge.go:101-102](judge/judge.go#L101-L102), [judge.go:132-136](judge/judge.go#L132-L136) |
| `runDocker` — commented-out buffer explanation | [docker.go:34-35](judge/docker.go#L34-L35) |
| `FileSystem` struct — constructed once, then ignored | [fileSystem.go:139-143](judge/fileSystem.go#L139-L143) |
| `ScoreboardHandler` — 3 lines of superseded broadcast notes | [scoreboard.go:54-55](api/handlers/scoreboard.go#L54-L55), [scoreboard.go:72](api/handlers/scoreboard.go#L72) |
| `TokenGenerator` — commented-out alternative | [util.go:30](utils/util.go#L30) |
| `models.Problem.SolvedCount` — declared, never written | [models.go:87](models/models.go#L87) |
| `RegisterRequest.Members` — decoded, never used | [dto.go:20](models/dto.go#L20) |
| `LoginRequest.BoundDeviceID` — decoded, never used (§10.5) | [dto.go:31](models/dto.go#L31) |
| `runDocker(submissionID …)` — parameter never read (§7.2) | [docker.go:12](judge/docker.go#L12) |
| `TestsDir` — `//asdasdasd` | [fileSystem.go:59](judge/fileSystem.go#L59) |

Version control holds the history. Deleting commented-out code costs nothing and removes the ambiguity of whether a reader is looking at something intentional.

### 13.3 🟠 ~25 unused dependencies

`go mod tidy -diff` confirms the module graph carries **gin**, **mongo-driver**, **quic-go**, **sonic**, **go-playground/validator**, **goccy/go-json**, **go-yaml**, **json-iterator**, and more — none of which appear anywhere in the source. The only real dependencies are `gorilla/websocket`, `gorm.io/gorm`, and `gorm.io/driver/sqlite`.

Every entry is also marked `// indirect`, including the three that *are* direct — so `go mod tidy` has never been run against the current source. Costs: slower builds, a much larger dependency surface for a product shipped as a single binary, and a `go.sum` that tells a reader nothing about what the project actually uses.

```bash
go mod tidy    # verified: removes ~25 modules, correctly marks the 3 real ones direct
```

### 13.4 🟡 Comments: two kinds, to be treated differently

**The good ones — keep these.** Several capture reasoning that would otherwise be lost, and are exactly what comments are for:

- [queue.go:23-25](judge/queue.go#L23-L25) — explains *why* the contest is read per-iteration rather than cached, and names the problem that motivated the change.
- [judge.go:105-106](judge/judge.go#L105-L106) — explains why `filepath.Abs` is required (the Docker daemon does not share the server's working directory). Non-obvious and correct.
- [submission.go:60-62](api/handlers/submission.go#L60-L62) — explains the struct-embedding trick and the frontend contract it serves.
- [fileSystem.go:94-98](judge/fileSystem.go#L94-L98) — explains why `LoadOrInitProblems` must exist given that `AutoMigrate` only creates schema.
- [run.go:25-27](api/handlers/run.go#L25-L27) — explains the semaphore and why the `defer` must be adjacent to the acquire.
- [admin.go:209-210](api/handlers/admin.go#L209-L210) — explains why saving a pointer updates the intended row.

**The rest.** Development-diary asides and self-directed commentary: [models.go:42](models/models.go#L42) `//why tf am i still doing this project`, [models.go:34](models/models.go#L34) `//…since i m way too dumb`, [auth.go:40](api/handlers/auth.go#L40) `//gpt made this bs`, [admin.go:185](api/handlers/admin.go#L185) `// should be path because its a partial update like duhhh bitch`, [run.go:2-3](judge/run.go#L2-L3).

The same register reaches end users through error strings — `"Bad request body bitch"` ([auth.go:24](api/handlers/auth.go#L24), [admin.go:49](api/handlers/admin.go#L49)), `"Server busy try again later babe"` ([run.go:33](api/handlers/run.go#L33)) — and through placeholder codes that correspond to nothing: `"internal server error 10200"` ([problems.go:53](api/handlers/problems.go#L53)), `"internal 123124"` ([admin.go:112](api/handlers/admin.go#L112)).

No judgement on a solo project. But this repository is described as a product for universities and placement cells; the moment a co-maintainer, a customer, or an evaluator reads it, these are the lines they remember. A single pass — keeping every comment in the first list — is maybe fifteen minutes. §9.3's structured error codes are what make the user-facing strings safe to reword without breaking the client.

### 13.5 🟡 Unmarked TODOs

Several pending items are written as prose rather than marked, so they are invisible to tooling and to `grep`:

- [admin.go:179-181](api/handlers/admin.go#L179-L181) — three unimplemented behaviours for disabled/disqualified teams
- [judge.go:91](judge/judge.go#L91) — the unimplemented `failure.json` (§7.4)
- [judge.go:136](judge/judge.go#L136) — unimplemented compile.log behaviour
- [auth.go:39](api/handlers/auth.go#L39) — *"guess its future me problem"*

Prefix with `// TODO:` or move to an issue tracker.

### 13.6 🟡 `dev_reset.sh` has a sequencing shortcut

[dev_reset.sh:22-26](dev_reset.sh#L22-L26) waits for the DB file to appear, then sleeps a fixed second with the comment `# let migrations finish` before seeding. A fixed sleep standing in for a readiness signal will eventually seed against a partially-migrated schema on a slower machine.

A `/health` endpoint already exists ([health.go](api/handlers/health.go)) — polling that instead is both more reliable and self-documenting. Also worth guarding the `rm -rf server_data` at [line 11](dev_reset.sh#L11) behind a confirmation or `--force`, given what running it on the wrong machine would do.

---

## 14. Proposed Target Structure

Not a rewrite — the same code, relocated so that dependencies point one direction:

```
Backend/
├── main.go                       # wire config → store → services → router; run
├── internal/
│   ├── config/
│   │   └── config.go             # Config struct, Load() from env with defaults
│   ├── domain/                   # (was models/) pure types + rules, zero imports
│   │   ├── team.go               # Team, TeamStatus, CanSubmit()
│   │   ├── submission.go         # Submission, SubmissionStatus, SubmissionVerdict
│   │   ├── contest.go            # Contest, ContestStatus, IsOpen(now)
│   │   ├── problem.go            # Problem
│   │   └── scoring.go            # Penalty(minutes, wrongAttempts) — pure, testable
│   ├── store/                    # (was db/) all GORM; DB unexported
│   │   ├── store.go              # Open(cfg), migrations
│   │   ├── teams.go
│   │   ├── submissions.go
│   │   ├── sessions.go
│   │   ├── problems.go
│   │   ├── contest.go
│   │   └── scoreboard.go
│   ├── service/                  # orchestration + enforcement
│   │   ├── auth.go               # login, device binding, override state machine
│   │   ├── submission.go         # validate → persist source → enqueue
│   │   ├── contest.go            # ContestState — owns the contest (§3.1)
│   │   ├── scoring.go            # stats + scoreboard update
│   │   └── admin.go
│   ├── judge/                    # depends on domain only
│   │   ├── worker.go             # worker loop + consumer-side interfaces
│   │   ├── judge.go              # processSubmission, runAllTestCases
│   │   ├── sandbox.go            # (was docker.go) Runner interface + docker impl
│   │   ├── compile.go            # single compile path (§7.1)
│   │   ├── language.go           # LanguageSpecs
│   │   ├── storage.go            # artifacts: source, compile.log, failure.json
│   │   ├── paths.go              # path helpers (§7.6)
│   │   ├── manifest.go           # parse only — no DB (§5.2)
│   │   └── compare.go            # NormalizeOutput (§7.7)
│   ├── transport/
│   │   ├── http/
│   │   │   ├── router.go         # NewRouter(deps) — method-scoped patterns
│   │   │   ├── middleware.go     # auth, admin, contest-open, logging, CORS
│   │   │   ├── response.go       # WriteJSON, WriteError, APIError
│   │   │   ├── dto.go            # request/response shapes only
│   │   │   ├── auth.go           # handlers, as methods on a Handlers struct
│   │   │   ├── submission.go
│   │   │   ├── problem.go
│   │   │   ├── scoreboard.go
│   │   │   ├── run.go
│   │   │   └── admin.go
│   │   └── ws/
│   │       ├── hub.go            # Hub — unexported map + mutex (§3.3)
│   │       └── client.go         # per-conn writePump + send channel
│   └── platform/                 # (was utils/) narrow, named packages
│       ├── token/
│       └── fsx/
├── server_data/                  # runtime data (gitignored)
└── testing/                      # unchanged
```

`internal/` prevents anything outside the module importing these — correct for a single-binary product, and it documents the intent.

**Dependency rule:** `domain` imports nothing. `store` and `judge` import `domain`. `service` imports `domain`, `store`, `judge`. `transport` imports `service` and `domain`. `main` imports everything and wires it. **No package imports anything to its right.**

That one rule, once it holds, is what makes §5's violations impossible to reintroduce by accident.

---

## 15. Prioritized Roadmap

Sequenced so each phase is independently shippable and reviewable. Estimates assume familiarity with the code.

*(Phases addressing behavioural defects are in the Correctness Audit and are not duplicated here. If both documents are being worked, the audit's fixes are best landed **after** Phase 0 — so they apply to formatted code — and **before** Phase 1, so restructuring does not carry known defects across package boundaries.)*

### Phase 0 — Mechanical, zero behaviour change (~45 min)

| # | Task | Ref |
| --- | --- | --- |
| 0.1 | `gofmt -w .` | §13.1 |
| 0.2 | `go mod tidy` | §13.3 |
| 0.3 | Delete commented-out code and the empty `contest.go` | §13.2 |
| 0.4 | Strip development-diary comments; keep the six good ones | §13.4 |
| 0.5 | Mark prose TODOs as `// TODO:` | §13.5 |
| 0.6 | Delete the four debug `log.Println` in `UpdateTeamProblemStats…` | §8.2 |
| 0.7 | Remove unused DTO fields and the unread `runDocker` parameter | §10.5, §7.2 |

Do this first so every later diff is readable. Nothing here changes behaviour, so it needs no testing beyond a build.

### Phase 1 — Ownership and seams (~1 day)

The prerequisite for everything after it.

| # | Task | Ref |
| --- | --- | --- |
| 1.1 | **Unexport `db.DB`; convert each compiler error into a store method** | §4.4 |
| 1.2 | `InitDB` → `Open(cfg) (*Store, error)`; no panic, no global | §8.1 |
| 1.3 | Wrap the contest in a `ContestState` type that hands out copies | §3.1 |
| 1.4 | Wire dependencies in `main` as values, not global assignment | §3.4 |
| 1.5 | Handlers become methods on a `Handlers` struct | §9.4 |
| 1.6 | `NewServeMux` + `NewRouter(deps)`; drop `DefaultServeMux` | §9.2 |

1.1 is the keystone: the compiler produces the work list, and every subsequent item gets easier once the seam exists.

### Phase 2 — Layering (~1 day)

| # | Task | Ref |
| --- | --- | --- |
| 2.1 | Introduce `service/`; move business logic out of handlers | §4.1, §4.2 |
| 2.2 | **Consolidate the team-status, contest-open, and penalty rules** | §4.3 |
| 2.3 | Invert `judge → ws` with a `ScoreboardNotifier` interface | §5.1 |
| 2.4 | Move `LoadOrInitProblems`'s DB writes out of `judge` | §5.2 |
| 2.5 | Give `judge` an artifact API; remove file I/O from handlers | §5.3 |
| 2.6 | `judge.ProblemDir(id)`; delete the duplicated path literal | §5.4 |
| 2.7 | Resolve the `FileSystem` struct question either way | §5.5 |
| 2.8 | Split `judge/fileSystem.go` into `paths` / `manifest` / `storage` | §7.6 |
| 2.9 | Extract per-connection `client` type; unexport `ws.Clients` / `ws.Mu` | §3.3 |

2.2 is the highest-leverage readability change in this document.

### Phase 3 — Modelling and API consistency (~1 day)

| # | Task | Ref |
| --- | --- | --- |
| 3.1 | **Unify problem ID types** — fixes the int/string endpoint inconsistency | §10.3 |
| 3.2 | Typed enums for the three status fields | §10.2 |
| 3.3 | `Teams`→`Team`, `Sessions`→`Session` (no migration needed) | §10.1 |
| 3.4 | Verdicts as `SubmissionVerdict` end-to-end | §7.3 |
| 3.5 | Split `models` into `domain` / `store` / `dto` | §8.5 |
| 3.6 | Split `utils`; move `NormalizeOutput` into `judge` | §10.4, §7.7 |
| 3.7 | Method-scoped routes; delete the 21 method checks | §9.1 |
| 3.8 | Structured `APIError` responses; reword user-facing strings | §9.3, §13.4 |
| 3.9 | Merge the two auth middlewares | §9.7 |
| 3.10 | `Config` struct from env; de-duplicate the repeated constants | §11 |

3.1 is worth pulling to the front of this phase if the frontend is being built against these endpoints now — it is a live contract inconsistency.

### Phase 4 — Judge internals (~half day)

| # | Task | Ref |
| --- | --- | --- |
| 4.1 | Deduplicate `compile` / `runAdHocCompile` | §7.1 |
| 4.2 | `runOpts` struct + `ctx` parameter for `runDocker` | §7.2 |
| 4.3 | `TestResult` struct from `runTestCase`; implement `failure.json` | §7.4 |
| 4.4 | Move the input redirect out of call-site string concatenation | §7.5 |
| 4.5 | Consider deleting the `ScoreBoard` table and deriving it | §3.2 |
| 4.6 | Pure `Penalty()` function out of the repository layer | §8.2 |

4.5 is the largest single simplification available — evaluate it seriously before doing 4.6, since it may remove the caller entirely.

### Phase 5 — Tests (~1 day)

| # | Task | Ref |
| --- | --- | --- |
| 5.1 | The four infrastructure-free tests (items 1–4) | §12.3 |
| 5.2 | `Runner` interface + fake; verdict-aggregation tests | §12.2, §12.3 |
| 5.3 | Store seam tests: scoreboard ordering, device binding | §12.3 |
| 5.4 | `go test -race ./...` in CI | §12.4 |
| 5.5 | `gofmt` check in CI so Phase 0 stays done | §13.1 |
| 5.6 | Add the missing column indexes | §8.4 |

Items 5.1's tests can be written **before** any restructuring lands, against the current code, provided §12.2's extraction steps 6 are done first. They are the cheapest insurance that Phases 1–4 preserve behaviour.

---

## Appendix: Findings by Leverage

Ranked by maintainability impact per unit of effort, not by risk.

**🔴 High leverage (7)** — each unblocks several others

§3.1 unowned globals · §4.3 rules defined more than once · §4.4 partial repository / `db.DB` global · §5.1 `judge → ws` inversion · §13.1 gofmt · §12.1 the testability blockers as a set · §14's one-direction dependency rule

**🟠 Medium leverage (21)** — meaningful cleanups, mostly independent

§3.2 derivable scoreboard · §3.3 WS connection type · §4.5 repository naming · §5.2–5.4 layering leaks · §7.1 duplicated compile · §7.2 `runDocker` signature · §7.3 stringly-typed verdicts · §7.4 discarded test detail · §8.1 `InitDB` shape · §8.2 no error return · §8.5 `models` overloading · §9.1 method checks · §9.2 `DefaultServeMux` · §9.3 response formats · §9.4 handler dependencies · §10.1–10.4 naming and modelling · §11 config · §13.3 unused deps

**🟡 Low leverage (18)** — worth doing, low urgency

§3.4 · §5.5 · §7.5–7.7 · §8.3–8.4 · §9.5–9.10 · §10.5–10.7 · §13.2, §13.4–13.6

---

*Analysis produced against the working tree at `/Users/snoopy/CampusCode/CampusCode-Server/Backend`. No files were modified. Behavioural, security, and data-integrity findings are maintained separately in the Correctness Audit.*
