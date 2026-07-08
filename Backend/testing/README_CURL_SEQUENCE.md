# CampusCode Sequential Curl Test Flow

Use this when you want deterministic testing **after** runtime is created by app startup.

## 0) One-time prep per reset

```bash
cd CampusCode
./testing/generate_server_data.sh
```

Start server in another terminal (this creates `server_data/runtime` + DB):

```bash
go run .
```

Optional baseline seed (teams/contest/history rows):

```bash
sqlite3 server_data/runtime/db.sqlite < testing/seed_runtime.sql
```

Mandatory minimum if you are skipping full seed:

```bash
sqlite3 server_data/runtime/db.sqlite < testing/seed_core.sql
```

## 1) Configure host

```bash
HOST="http://localhost:8080"
```

## 2) Register fresh teams

```bash
RUN_TAG=$(date +%s)
TEAM_A="team_alpha_${RUN_TAG}"
TEAM_B="team_beta_${RUN_TAG}"
TEAM_C="team_gamma_${RUN_TAG}"

REG1=$(curl -s -X POST "$HOST/register" -H "Content-Type: application/json" -d "{\"name\":\"$TEAM_A\",\"members\":\"a1,a2,a3\"}")
REG2=$(curl -s -X POST "$HOST/register" -H "Content-Type: application/json" -d "{\"name\":\"$TEAM_B\",\"members\":\"b1,b2,b3\"}")
REG3=$(curl -s -X POST "$HOST/register" -H "Content-Type: application/json" -d "{\"name\":\"$TEAM_C\",\"members\":\"c1,c2,c3\"}")

echo "$REG1"
echo "$REG2"
echo "$REG3"
```

## 3) Login checks using generated login codes

```bash
CODE1=$(printf "%s" "$REG1" | python3 -c 'import json,sys; print(json.loads(sys.stdin.read()).get("login_code",""))')
CODE2=$(printf "%s" "$REG2" | python3 -c 'import json,sys; print(json.loads(sys.stdin.read()).get("login_code",""))')
CODE3=$(printf "%s" "$REG3" | python3 -c 'import json,sys; print(json.loads(sys.stdin.read()).get("login_code",""))')

curl -i -s -X POST "$HOST/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: device-alpha-01" \
  -d "{\"name\":\"$TEAM_A\",\"login_code\":\"$CODE1\",\"bound_device_id\":\"device-alpha-01\"}"

curl -i -s -X POST "$HOST/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: device-beta-01" \
  -d "{\"name\":\"$TEAM_B\",\"login_code\":\"$CODE2\",\"bound_device_id\":\"device-beta-01\"}"

curl -i -s -X POST "$HOST/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: device-gamma-01" \
  -d "{\"name\":\"$TEAM_C\",\"login_code\":\"$CODE3\",\"bound_device_id\":\"device-gamma-01\"}"
```

Resolve team IDs (submit endpoint uses numeric team IDs):

```bash
TEAM_A_ID=$(sqlite3 server_data/runtime/db.sqlite "select id from teams where name='${TEAM_A}' order by id desc limit 1;")
TEAM_B_ID=$(sqlite3 server_data/runtime/db.sqlite "select id from teams where name='${TEAM_B}' order by id desc limit 1;")
TEAM_C_ID=$(sqlite3 server_data/runtime/db.sqlite "select id from teams where name='${TEAM_C}' order by id desc limit 1;")
```

## 4) Submission sequence (practical contest style)

### 4.1 Team A solves problem 1 (AC)

```bash
curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "$(cat <<EOF
{
  "team_id": $TEAM_A_ID,
  "problem_id": 1,
  "code": "#include <bits/stdc++.h>\nusing namespace std;\nint main(){ long long a,b; if(!(cin>>a>>b)) return 0; cout<<a+b<<\"\\n\"; }"
}
EOF
)"
```

### 4.2 Team B wrong attempt on problem 2 (WA)

```bash
curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "$(cat <<EOF
{
  "team_id": $TEAM_B_ID,
  "problem_id": 2,
  "code": "#include <bits/stdc++.h>\nusing namespace std;\nint main(){ long long a,b; if(!(cin>>a>>b)) return 0; cout<<a+b<<\"\\n\"; }"
}
EOF
)"
```

### 4.3 Team B retries problem 2 and solves (AC)

```bash
curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "$(cat <<EOF
{
  "team_id": $TEAM_B_ID,
  "problem_id": 2,
  "code": "#include <bits/stdc++.h>\nusing namespace std;\nint main(){ long long a,b; if(!(cin>>a>>b)) return 0; cout<<a*b<<\"\\n\"; }"
}
EOF
)"
```

### 4.4 Team C solves problem 7 (AC)

```bash
curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "$(cat <<EOF
{
  "team_id": $TEAM_C_ID,
  "problem_id": 7,
  "code": "#include <bits/stdc++.h>\nusing namespace std;\nint main(){ string s; if(!(cin>>s)) return 0; int bal=0; for(char c:s){ if(c=='(') bal++; else bal--; if(bal<0){ cout<<\"NO\\n\"; return 0; } } cout<<(bal==0?\"YES\":\"NO\")<<\"\\n\"; }\n"
}
EOF
)"
```

### 4.5 Team C solves problem 8 (AC)

```bash
curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "$(cat <<EOF
{
  "team_id": $TEAM_C_ID,
  "problem_id": 8,
  "code": "#include <bits/stdc++.h>\nusing namespace std;\nbool isPrime(long long x){ if(x<2) return false; if(x%2==0) return x==2; for(long long i=3;i*i<=x;i+=2) if(x%i==0) return false; return true; }\nint main(){ ios::sync_with_stdio(false); cin.tie(nullptr); int q; if(!(cin>>q)) return 0; while(q--){ long long x; cin>>x; cout<<(isPrime(x)?\"YES\":\"NO\")<<\"\\n\"; } }\n"
}
EOF
)"
```

## 5) Additional edge cases

```bash
curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "{\"team_id\":$TEAM_A_ID,\"problem_id\":3,\"code\":\"#include <bits/stdc++.h>\\nusing namespace std;\\nint main(){int n; if(!(cin>>n)) return 0; long long s=0,x; for(int i=0;i<n;i++){cin>>x; s+=x;} cout<<s<<\\\"\\\\n\\\"; }\"}"

curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "{\"team_id\":$TEAM_B_ID,\"problem_id\":4,\"code\":\"#include <bits/stdc++.h>\\nusing namespace std;\\nint main(){ this_does_not_compile }\\n\"}"

curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "{\"team_id\":$TEAM_A_ID,\"problem_id\":6,\"code\":\"#include <bits/stdc++.h>\\nusing namespace std;\\nint main(){long long n,m; if(!(cin>>n>>m)) return 0; cout<<(n*m)%1000000007LL<<\\\"\\\\n\\\"; }\"}"

curl -s -X POST "$HOST/submit" \
  -H "Content-Type: application/json" \
  -d "{\"team_id\":$TEAM_C_ID,\"problem_id\":1,\"code\":\"#include <bits/stdc++.h>\\nusing namespace std;\\nint main(){ vector<int> v; cout<<v.at(1)<<\\\"\\\\n\\\"; }\\n\"}"
```

## 6) Verify scoreboard/API state

```bash
curl -s "$HOST/scoreboard"
curl -s "$HOST/scoreboard" | jq
```

## 7) Optional DB checks

```bash
sqlite3 server_data/runtime/db.sqlite "select id,team_id,problem_id,status,verdict,submitted_at from submissions order by id desc limit 20;"
sqlite3 server_data/runtime/db.sqlite "select team_id,total_solved,total_penalty,last_ac_time,rank from score_boards order by total_solved desc,total_penalty asc;"
```

## Fastest workflow

If you want all these requests in one shot, run:

```bash
bash testing/replay_submissions.sh
```

Notes:

- The replay script now registers teams and captures their generated `login_code` values automatically for login tests.
- Replay now runs a realistic 3-team contest path (register -> login -> submissions -> retries).
- If `/scoreboard` returns `null`, script also prints DB snapshots so you can still validate submission flow and stat persistence.
