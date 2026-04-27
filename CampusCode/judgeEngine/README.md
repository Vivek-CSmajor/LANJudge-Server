## SUBMISSION FLOW 
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

Takes source file like sub.cpp and compiles it (Lets assume only C++ submissions for now) via say a command like `g++ submission.cpp -O2 -std=c++17 -o program`


## File System to follow : 
```
server_data/
│
├── contest/                     (READ-ONLY during contest)
│   ├── manifest.json
│   └── problems/
│       ├── 1/
│       │   ├── meta.json
│       │   ├── statement.md
│       │   └── tests/
│       │       ├── 1.in
│       │       ├── 1.out
│       │       ├── 2.in
│       │       ├── 2.out
│       │
│       └── 2/
│           ├── meta.json
│           ├── statement.md
│           └── tests/
│
└── runtime/                     (MUTABLE)
    ├── db.sqlite
    │
    ├── submissions/             (PERSISTENT)
    │   ├── 101/
    │   │   ├── source.cpp
    │   │   ├── compile.log        (only if compile error)
    │   │   └── failure.json       (only if WA/RE/TLE)
    │   │
    │   └── 102/
    │
    └── workspaces/              (EPHEMERAL, PER SUBMISSION)
        ├── 101/
        │   ├── source.cpp
        │   ├── program
        │   ├── input.txt
        │   ├── output.txt
        │   └── compile.log
        │
        └── 102/a
```
manifest.json --> contest configuration

meta.json inside problems --> problem constraints

db.sqlite --> Stores metadata like participants,submissions,verdicts,scoreboard etc. This is the Main and only db for the judgeEngine

workspaces --> easier sandboxing , like if we solving a problem 101 then we copy the binary here,testcase input,run it  etc

DIRECTORY RESPONSIBILITIES

/contest/

* Immutable contest data
* Testcases, problem configs
* Never modified during execution

⸻

/runtime/db.sqlite

Stores:

* users
* submissions metadata
* verdicts
* scoreboard

/runtime/submissions/

Persistent record per submission:

* source code (for admin viewing)
* compile logs (if failure)
* minimal failure info

/runtime/workspaces/

* One folder per submission
* Used for:
    * compilation
    * execution
    * output comparison
* Fully isolated
* Deleted after execution