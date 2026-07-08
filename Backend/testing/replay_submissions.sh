#!/usr/bin/env bash

set -euo pipefail

HOST="${HOST:-http://localhost:8080}"
DB_PATH="${DB_PATH:-server_data/runtime/db.sqlite}"
RUN_TAG="${RUN_TAG:-$(date +%s)}"

TEAM_A="team_alpha_${RUN_TAG}"
TEAM_B="team_beta_${RUN_TAG}"
TEAM_C="team_gamma_${RUN_TAG}"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing required command: $1"; exit 1; }
}

require_cmd curl
require_cmd python3
require_cmd sqlite3

json_get_login_code() {
  python3 -c 'import json,sys; print(json.loads(sys.stdin.read()).get("login_code",""))'
}

json_payload() {
  # Usage: json_payload <team_id:int> <problem_id:int> <code:string>
  local team_id="$1"
  local problem_id="$2"
  local code="$3"
  python3 - <<'PY' "$team_id" "$problem_id" "$code"
import json,sys
team_id = int(sys.argv[1])
problem_id = int(sys.argv[2])
code = sys.argv[3]
print(json.dumps({"team_id": team_id, "problem_id": problem_id, "code": code}))
PY
}

get_team_id() {
  local team_name="$1"
  sqlite3 "$DB_PATH" "select id from teams where name='${team_name}' order by id desc limit 1;"
}

post_submit_code() {
  local label="$1"
  local team_id="$2"
  local problem_id="$3"
  local code="$4"
  local payload
  payload="$(json_payload "$team_id" "$problem_id" "$code")"
  echo "==> $label"
  curl -s -X POST "$HOST/submit" \
    -H "Content-Type: application/json" \
    -d "$payload"
  echo
  sleep 1
}

echo "[1/7] Register 3 contest teams and capture login codes"
REG1="$(curl -s -X POST "$HOST/register" -H "Content-Type: application/json" -d "{\"name\":\"$TEAM_A\",\"members\":\"a1,a2,a3\"}")"
REG2="$(curl -s -X POST "$HOST/register" -H "Content-Type: application/json" -d "{\"name\":\"$TEAM_B\",\"members\":\"b1,b2,b3\"}")"
REG3="$(curl -s -X POST "$HOST/register" -H "Content-Type: application/json" -d "{\"name\":\"$TEAM_C\",\"members\":\"c1,c2,c3\"}")"

echo "$REG1"
echo "$REG2"
echo "$REG3"

CODE1="$(printf "%s" "$REG1" | json_get_login_code)"
CODE2="$(printf "%s" "$REG2" | json_get_login_code)"
CODE3="$(printf "%s" "$REG3" | json_get_login_code)"

echo "Captured login codes:"
echo "$TEAM_A=$CODE1"
echo "$TEAM_B=$CODE2"
echo "$TEAM_C=$CODE3"
echo

echo "[2/7] Login (bind devices) + negative login"
curl -i -s -X POST "$HOST/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: device-alpha-01" \
  -d "{\"name\":\"$TEAM_A\",\"login_code\":\"$CODE1\",\"bound_device_id\":\"device-alpha-01\"}"
echo

curl -i -s -X POST "$HOST/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: device-beta-01" \
  -d "{\"name\":\"$TEAM_B\",\"login_code\":\"$CODE2\",\"bound_device_id\":\"device-beta-01\"}"
echo

curl -i -s -X POST "$HOST/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: device-gamma-01" \
  -d "{\"name\":\"$TEAM_C\",\"login_code\":\"$CODE3\",\"bound_device_id\":\"device-gamma-01\"}"
echo

echo "Negative login check (expected 401):"
curl -i -s -X POST "$HOST/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: wrong-device" \
  -d "{\"name\":\"$TEAM_A\",\"login_code\":\"BADCODE\",\"bound_device_id\":\"wrong-device\"}" || true
echo

echo "[3/7] Resolve numeric team IDs for submissions"
if [[ ! -f "$DB_PATH" ]]; then
  echo "DB not found at $DB_PATH"
  exit 1
fi

TEAM_A_ID="$(get_team_id "$TEAM_A")"
TEAM_B_ID="$(get_team_id "$TEAM_B")"
TEAM_C_ID="$(get_team_id "$TEAM_C")"

echo "$TEAM_A -> id=$TEAM_A_ID"
echo "$TEAM_B -> id=$TEAM_B_ID"
echo "$TEAM_C -> id=$TEAM_C_ID"
echo

if [[ -z "$TEAM_A_ID" || -z "$TEAM_B_ID" || -z "$TEAM_C_ID" ]]; then
  echo "Failed to resolve one or more team IDs from DB."
  exit 1
fi

echo "[4/7] Practical contest submission flow (3 teams)"
CODE_P1_AC=$'#include <bits/stdc++.h>\nusing namespace std;\nint main(){ long long a,b; if(!(cin>>a>>b)) return 0; cout<<a+b<<"\\n"; }\n'
CODE_P2_WA=$'#include <bits/stdc++.h>\nusing namespace std;\nint main(){ long long a,b; if(!(cin>>a>>b)) return 0; cout<<a+b<<"\\n"; }\n'
CODE_P2_AC=$'#include <bits/stdc++.h>\nusing namespace std;\nint main(){ long long a,b; if(!(cin>>a>>b)) return 0; cout<<a*b<<"\\n"; }\n'
CODE_P7_AC=$'#include <bits/stdc++.h>\nusing namespace std;\nint main(){ string s; if(!(cin>>s)) return 0; int bal=0; for(char c:s){ if(c==\'(\') bal++; else bal--; if(bal<0){ cout<<"NO\\n"; return 0; } } cout<<(bal==0?"YES":"NO")<<"\\n"; }\n'
CODE_P8_AC=$'#include <bits/stdc++.h>\nusing namespace std;\nbool isPrime(long long x){ if(x<2) return false; if(x%2==0) return x==2; for(long long i=3;i*i<=x;i+=2) if(x%i==0) return false; return true; }\nint main(){ ios::sync_with_stdio(false); cin.tie(nullptr); int q; if(!(cin>>q)) return 0; while(q--){ long long x; cin>>x; cout<<(isPrime(x)?"YES":"NO")<<"\\n"; } }\n'
CODE_P3_AC=$'#include <bits/stdc++.h>\nusing namespace std;\nint main(){ ios::sync_with_stdio(false); cin.tie(nullptr); int n; if(!(cin>>n)) return 0; long long s=0,x; for(int i=0;i<n;i++){cin>>x; s+=x;} cout<<s<<"\\n"; }\n'
CODE_P6_WA=$'#include <bits/stdc++.h>\nusing namespace std;\nint main(){ long long n,m; if(!(cin>>n>>m)) return 0; cout << (n*m)%1000000007LL << "\\n"; }\n'
CODE_P6_AC=$'#include <bits/stdc++.h>\nusing namespace std;\nconst long long MOD=1000000007;\nint main(){ int n,m; if(!(cin>>n>>m)) return 0; vector<long long> dp(m,1); for(int i=1;i<n;i++){ for(int j=1;j<m;j++){ dp[j]=(dp[j]+dp[j-1])%MOD; }} cout<<dp[m-1]<<"\\n"; }\n'
CODE_CE=$'#include <bits/stdc++.h>\nusing namespace std;\nint main(){ this_does_not_compile }\n'
CODE_RE=$'#include <bits/stdc++.h>\nusing namespace std;\nint main(){ vector<int> v; cout<<v.at(1)<<"\\n"; }\n'

post_submit_code "Team A P1 AC" "$TEAM_A_ID" 1 "$CODE_P1_AC"
post_submit_code "Team B P2 WA (uses + instead of *)" "$TEAM_B_ID" 2 "$CODE_P2_WA"
post_submit_code "Team B P2 AC retry" "$TEAM_B_ID" 2 "$CODE_P2_AC"
post_submit_code "Team C P7 AC (bracket balance)" "$TEAM_C_ID" 7 "$CODE_P7_AC"
post_submit_code "Team C P8 AC (prime queries)" "$TEAM_C_ID" 8 "$CODE_P8_AC"
post_submit_code "Team A P3 AC (large sum)" "$TEAM_A_ID" 3 "$CODE_P3_AC"
post_submit_code "Team A P6 WA (wrong formula n*m)" "$TEAM_A_ID" 6 "$CODE_P6_WA"
post_submit_code "Team A P6 AC retry" "$TEAM_A_ID" 6 "$CODE_P6_AC"
post_submit_code "Team B P4 CE (compile error case)" "$TEAM_B_ID" 4 "$CODE_CE"
post_submit_code "Team C P1 RE (runtime error case)" "$TEAM_C_ID" 1 "$CODE_RE"

echo "[5/7] Wait for workers to finish current batch"
for _ in $(seq 1 30); do
  pending="$(sqlite3 "$DB_PATH" "select count(*) from submissions where status!='Done';")"
  if [[ "$pending" == "0" ]]; then
    break
  fi
  sleep 1
done

echo "[6/7] Scoreboard snapshot from API"
curl -s "$HOST/scoreboard"
echo

echo "[7/7] DB snapshots"
echo "Latest submissions:"
sqlite3 "$DB_PATH" "select id,team_id,problem_id,status,verdict,submitted_at from submissions order by id desc limit 25;"
echo
echo "team_problem_stats:"
sqlite3 "$DB_PATH" "select team_id,problem_id,wrong_attempt_count,solved,first_ac_time,penalty from team_problem_stats order by team_id,problem_id;"
echo
echo "score_boards:"
sqlite3 "$DB_PATH" "select team_id,total_solved,total_penalty,last_ac_time,rank from score_boards order by total_solved desc,total_penalty asc;"
echo
echo "Scenario teams:"
echo "- $TEAM_A (id=$TEAM_A_ID)"
echo "- $TEAM_B (id=$TEAM_B_ID)"
echo "- $TEAM_C (id=$TEAM_C_ID)"
echo
echo "Replay complete. Host: $HOST"
