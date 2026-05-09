# CampusCode Testing Dataset Kit

This folder gives you repeatable seed data and a larger `server_data` filesystem for heavy local testing.

## What this kit creates

- A regenerated `server_data/contest` with 8 problems and multiple testcases each.
- Runtime (`server_data/runtime`) is intentionally **not** created here.
- Your app remains the source of truth for creating runtime + `db.sqlite`.

## Files in this folder

- `testing/generate_server_data.sh` - generates contest data only.
- `testing/seed_core.sql` - minimal required DB bootstrap (`contests` row).
- `testing/seed_runtime.sql` - SQL seed for contest/team/submission/scoreboard related tables.
- `testing/README_CURL_SEQUENCE.md` - full ordered curl workflow.
- `testing/replay_submissions.sh` - one-shot practical contest replay (3 teams register/login + mixed real submissions).

## How to use

1) Generate contest fixtures

```bash
cd CampusCode
chmod +x testing/generate_server_data.sh
./testing/generate_server_data.sh
```

2) Start your server once so your filesystem bootstrap creates `server_data/runtime` and GORM creates `server_data/runtime/db.sqlite`

3) Seed mandatory core DB data (`contests`)

```bash
sqlite3 server_data/runtime/db.sqlite < testing/seed_core.sql
```

4) Optional full-history seed (pre-filled teams, old submissions, scoreboard)

```bash
sqlite3 server_data/runtime/db.sqlite < testing/seed_runtime.sql
```

5) Run curl tests (register/login/submit/scoreboard)

Set host once:

```bash
HOST="http://localhost:8080"
```

### A) Register teams quickly

```bash
curl -s -X POST "$HOST/register" -H "Content-Type: application/json" -d '{"name":"team_alpha","members":"a,b,c"}'
curl -s -X POST "$HOST/register" -H "Content-Type: application/json" -d '{"name":"team_beta","members":"d,e,f"}'
curl -s -X POST "$HOST/register" -H "Content-Type: application/json" -d '{"name":"team_gamma","members":"g,h,i"}'
```

### B) Login samples (device binding/session coverage)

Use known seeded users from `testing/seed_runtime.sql`:

```bash
curl -i -s -X POST "$HOST/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: device-a-01" \
  -d '{"name":"alpha_ninjas","login_code":"ALPHA123","bound_device_id":"device-a-01"}'

curl -i -s -X POST "$HOST/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: device-c-02" \
  -d '{"name":"cache_cowboys","login_code":"CACHE333","bound_device_id":"device-c-02"}'
```

### C) Submit commands by verdict type

#### 1) AC (problem 1: addition)

```bash
curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "$(cat <<'EOF'
{
  "team_id": 1,
  "problem_id": 1,
  "code": "#include <bits/stdc++.h>\nusing namespace std;\nint main(){ long long a,b; if(!(cin>>a>>b)) return 0; cout << a+b << \"\\n\"; }"
}
EOF
)"
```

#### 2) WA (submit multiplication wrong logic on problem 2)

```bash
curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "$(cat <<'EOF'
{
  "team_id": 2,
  "problem_id": 2,
  "code": "#include <bits/stdc++.h>\nusing namespace std;\nint main(){ long long a,b; if(!(cin>>a>>b)) return 0; cout << a+b << \"\\n\"; }"
}
EOF
)"
```

#### 3) CE (compilation error)

```bash
curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "$(cat <<'EOF'
{
  "team_id": 3,
  "problem_id": 1,
  "code": "#include <bits/stdc++.h>\nusing namespace std;\nint main(){ this_will_not_compile }\n"
}
EOF
)"
```

#### 4) TLE (infinite loop)

```bash
curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "$(cat <<'EOF'
{
  "team_id": 4,
  "problem_id": 1,
  "code": "#include <bits/stdc++.h>\nusing namespace std;\nint main(){ while(true){} }\n"
}
EOF
)"
```

#### 5) RE (runtime crash)

```bash
curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "$(cat <<'EOF'
{
  "team_id": 5,
  "problem_id": 1,
  "code": "#include <bits/stdc++.h>\nusing namespace std;\nint main(){ int x=0; cout << 10/x << \"\\n\"; }\n"
}
EOF
)"
```

### D) Stress loop to create many submissions fast

```bash
for i in $(seq 1 25); do
  curl -s -X POST "$HOST/submit" \
    -H "Content-Type: application/json" \
    -d "{\"team_id\":1,\"problem_id\":1,\"code\":\"#include <bits/stdc++.h>\\nusing namespace std;int main(){long long a,b;cin>>a>>b;cout<<a+b<<'\\\\n';}\"}" >/dev/null
done
echo "25 submissions pushed"
```

### E) Scoreboard checks

```bash
curl -s "$HOST/scoreboard"
```

Pretty print with jq if installed:

```bash
curl -s "$HOST/scoreboard" | jq
```

## No-manual workaround (recommended)

After server is running, execute:

```bash
chmod +x testing/replay_submissions.sh
bash testing/replay_submissions.sh
```

This gives you:

- auth/register traffic
- automatic login using freshly generated login codes
- 3 teams with realistic contest progression (wrong attempts + retries + solved)
- cross-problem submissions across P1/P2/P3/P4/P6/P7/P8
- explicit CE and RE cases for failure-path testing
- immediate scoreboard snapshot
- sqlite DB fallback snapshots (`submissions`, `team_problem_stats`, `score_boards`)

## Notes

- This kit does not modify your Go source code.
- `generate_server_data.sh` intentionally overwrites `server_data/contest` so you can reset quickly.
- Runtime ownership stays with your app bootstrap.
- If you want clean reruns, remove only runtime artifacts between runs:

```bash
rm -rf server_data/runtime/submissions server_data/runtime/workspaces
mkdir -p server_data/runtime/submissions server_data/runtime/workspaces
```
