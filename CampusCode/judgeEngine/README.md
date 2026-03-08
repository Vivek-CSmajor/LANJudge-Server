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
├── contest/
│   ├── manifest.json
│   └── problems/
│       ├── A/
│       │   ├── meta.json
│       │   └── tests/
│       │       ├── in
│       │       ├── out
│       │       
│       │       
│       │
│       └── B/
│           ├── meta.json
│           └── tests/
│
└── runtime/
    ├── db.sqlite
    ├── submissions/
    │   ├── 101/
    │   │   ├── source.cpp
    │   │   ├── binary
    │   │   └── compile.log
    │   │
    │   └── 102/
    │
    ├── runs/
    │   ├── 101/
    │   │   ├── 1.out
    │   │   └── 2.out
    │
    └── workspaces/
        ├── 101/
        └── 102/
```
manifest.json --> contest configuration

meta.json inside problems --> problem constraints

db.sqlite --> Stores metadata like participants,submissions,verdicts,scoreboard etc. This is the Main and only db for the judgeEngine

runs --> for outputs during execution of tests

workspaces --> easier sandboxing , like if we solving a problem 101 then we copy the binary here,testcase input,run it  etc

