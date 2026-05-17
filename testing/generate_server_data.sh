#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVER_DATA_DIR="$ROOT_DIR/server_data"
CONTEST_DIR="$SERVER_DATA_DIR/contest"

echo "[1/5] Resetting contest fixture directory..."
rm -rf "$CONTEST_DIR"
mkdir -p "$CONTEST_DIR/problems"

echo "[2/5] Writing contest manifest..."
cat > "$CONTEST_DIR/manifest.json" <<'JSON'
{
  "contest_id": "campuscode_stress_2026_01",
  "name": "CampusCode Stress Fixture Contest",
  "duration_minutes": 180,
  "problems": [
    { "id": "1", "title": "Simple Addition", "path": "problems/1" },
    { "id": "2", "title": "Multiplication", "path": "problems/2" },
    { "id": "3", "title": "Array Sum", "path": "problems/3" },
    { "id": "4", "title": "Maximum Subarray", "path": "problems/4" },
    { "id": "5", "title": "Pair Difference", "path": "problems/5" },
    { "id": "6", "title": "Grid Paths", "path": "problems/6" },
    { "id": "7", "title": "Bracket Balance", "path": "problems/7" },
    { "id": "8", "title": "Prime Queries", "path": "problems/8" }
  ]
}
JSON

write_problem() {
  local pid="$1"
  local title="$2"
  local description="$3"
  local input="$4"
  local output="$5"
  local constraints="$6"
  local samples="$7"

  local pdir="$CONTEST_DIR/problems/$pid"
  mkdir -p "$pdir/tests"

  cat > "$pdir/meta.json" <<'JSON'
{
  "time_limit_ms": 1000,
  "memory_limit_mb": 256
}
JSON

  cat > "$pdir/statement.md" <<EOF
# Problem $pid: $title

## Description
$description

## Input
$input

## Output
$output

## Constraints
$constraints

## Examples
$samples
EOF
}

echo "[3/5] Writing problem statements and metadata..."
write_problem "1" "Simple Addition" \
"Given two integers a and b, print a + b." \
"Two space-separated integers a and b." \
"Print one integer: the sum." \
"-10^9 <= a, b <= 10^9" \
"### Example 1
Input:
\`\`\`
2 3
\`\`\`
Output:
\`\`\`
5
\`\`\`"

write_problem "2" "Multiplication" \
"Given two integers a and b, print a * b." \
"Two space-separated integers a and b." \
"Print one integer: the product." \
"-10^9 <= a, b <= 10^9" \
"### Example 1
Input:
\`\`\`
7 8
\`\`\`
Output:
\`\`\`
56
\`\`\`"

write_problem "3" "Array Sum" \
"Given an array of N integers, print the sum of all values." \
"First line: N. Second line: N space-separated integers." \
"Print one integer: total sum." \
"1 <= N <= 200000, |a_i| <= 10^9" \
"### Example 1
Input:
\`\`\`
5
1 2 3 4 5
\`\`\`
Output:
\`\`\`
15
\`\`\`"

write_problem "4" "Maximum Subarray" \
"Given N integers, find the maximum possible contiguous subarray sum." \
"First line N, second line contains N integers." \
"Print one integer: maximum subarray sum." \
"1 <= N <= 200000, |a_i| <= 10^9" \
"### Example 1
Input:
\`\`\`
8
-2 -3 4 -1 -2 1 5 -3
\`\`\`
Output:
\`\`\`
7
\`\`\`"

write_problem "5" "Pair Difference" \
"Given a sorted array and integer K, count pairs (i,j) where j > i and a[j] - a[i] = K." \
"First line: N K. Second line: sorted array of N integers." \
"Print one integer: count of valid pairs." \
"1 <= N <= 200000, 0 <= K <= 10^9" \
"### Example 1
Input:
\`\`\`
5 2
1 3 5 7 9
\`\`\`
Output:
\`\`\`
4
\`\`\`"

write_problem "6" "Grid Paths" \
"Given n and m, count ways to reach (n,m) from (1,1) moving only right or down. Output modulo 1e9+7." \
"One line with n and m." \
"Print one integer: number of paths modulo 1000000007." \
"1 <= n, m <= 2000" \
"### Example 1
Input:
\`\`\`
3 3
\`\`\`
Output:
\`\`\`
6
\`\`\`"

write_problem "7" "Bracket Balance" \
"Given a string with only '(' and ')', determine if it is balanced." \
"One line with bracket string s." \
"Print YES if balanced, otherwise NO." \
"1 <= |s| <= 200000" \
"### Example 1
Input:
\`\`\`
(()())
\`\`\`
Output:
\`\`\`
YES
\`\`\`"

write_problem "8" "Prime Queries" \
"For each query x, answer if x is prime." \
"First line Q, followed by Q lines each containing x." \
"For each query print YES or NO." \
"1 <= Q <= 100000, 1 <= x <= 10^7" \
"### Example 1
Input:
\`\`\`
3
2
15
17
\`\`\`
Output:
\`\`\`
YES
NO
YES
\`\`\`"

echo "[4/5] Writing testcases..."

# Problem 1 tests
cat > "$CONTEST_DIR/problems/1/tests/1.in" <<'EOF'
2 3
EOF
cat > "$CONTEST_DIR/problems/1/tests/1.out" <<'EOF'
5
EOF
cat > "$CONTEST_DIR/problems/1/tests/2.in" <<'EOF'
-100 250
EOF
cat > "$CONTEST_DIR/problems/1/tests/2.out" <<'EOF'
150
EOF
cat > "$CONTEST_DIR/problems/1/tests/3.in" <<'EOF'
0 0
EOF
cat > "$CONTEST_DIR/problems/1/tests/3.out" <<'EOF'
0
EOF
cat > "$CONTEST_DIR/problems/1/tests/4.in" <<'EOF'
999999999 1
EOF
cat > "$CONTEST_DIR/problems/1/tests/4.out" <<'EOF'
1000000000
EOF

# Problem 2 tests
cat > "$CONTEST_DIR/problems/2/tests/1.in" <<'EOF'
5 3
EOF
cat > "$CONTEST_DIR/problems/2/tests/1.out" <<'EOF'
15
EOF
cat > "$CONTEST_DIR/problems/2/tests/2.in" <<'EOF'
-7 8
EOF
cat > "$CONTEST_DIR/problems/2/tests/2.out" <<'EOF'
-56
EOF
cat > "$CONTEST_DIR/problems/2/tests/3.in" <<'EOF'
0 123456
EOF
cat > "$CONTEST_DIR/problems/2/tests/3.out" <<'EOF'
0
EOF
cat > "$CONTEST_DIR/problems/2/tests/4.in" <<'EOF'
10000 10000
EOF
cat > "$CONTEST_DIR/problems/2/tests/4.out" <<'EOF'
100000000
EOF

# Problem 3 tests
cat > "$CONTEST_DIR/problems/3/tests/1.in" <<'EOF'
5
1 2 3 4 5
EOF
cat > "$CONTEST_DIR/problems/3/tests/1.out" <<'EOF'
15
EOF
cat > "$CONTEST_DIR/problems/3/tests/2.in" <<'EOF'
4
-1 -2 -3 -4
EOF
cat > "$CONTEST_DIR/problems/3/tests/2.out" <<'EOF'
-10
EOF
cat > "$CONTEST_DIR/problems/3/tests/3.in" <<'EOF'
6
1000000000 1000000000 1000000000 1000000000 1000000000 1000000000
EOF
cat > "$CONTEST_DIR/problems/3/tests/3.out" <<'EOF'
6000000000
EOF
cat > "$CONTEST_DIR/problems/3/tests/4.in" <<'EOF'
1
42
EOF
cat > "$CONTEST_DIR/problems/3/tests/4.out" <<'EOF'
42
EOF

# Problem 4 tests
cat > "$CONTEST_DIR/problems/4/tests/1.in" <<'EOF'
8
-2 -3 4 -1 -2 1 5 -3
EOF
cat > "$CONTEST_DIR/problems/4/tests/1.out" <<'EOF'
7
EOF
cat > "$CONTEST_DIR/problems/4/tests/2.in" <<'EOF'
5
-1 -2 -3 -4 -5
EOF
cat > "$CONTEST_DIR/problems/4/tests/2.out" <<'EOF'
-1
EOF
cat > "$CONTEST_DIR/problems/4/tests/3.in" <<'EOF'
6
1 2 3 4 5 6
EOF
cat > "$CONTEST_DIR/problems/4/tests/3.out" <<'EOF'
21
EOF
cat > "$CONTEST_DIR/problems/4/tests/4.in" <<'EOF'
7
5 -1 5 -1 5 -1 5
EOF
cat > "$CONTEST_DIR/problems/4/tests/4.out" <<'EOF'
17
EOF

# Problem 5 tests
cat > "$CONTEST_DIR/problems/5/tests/1.in" <<'EOF'
5 2
1 3 5 7 9
EOF
cat > "$CONTEST_DIR/problems/5/tests/1.out" <<'EOF'
4
EOF
cat > "$CONTEST_DIR/problems/5/tests/2.in" <<'EOF'
6 0
1 1 1 2 2 3
EOF
cat > "$CONTEST_DIR/problems/5/tests/2.out" <<'EOF'
4
EOF
cat > "$CONTEST_DIR/problems/5/tests/3.in" <<'EOF'
7 3
1 2 4 7 10 13 16
EOF
cat > "$CONTEST_DIR/problems/5/tests/3.out" <<'EOF'
5
EOF
cat > "$CONTEST_DIR/problems/5/tests/4.in" <<'EOF'
5 100
1 2 3 4 5
EOF
cat > "$CONTEST_DIR/problems/5/tests/4.out" <<'EOF'
0
EOF

# Problem 6 tests
cat > "$CONTEST_DIR/problems/6/tests/1.in" <<'EOF'
2 2
EOF
cat > "$CONTEST_DIR/problems/6/tests/1.out" <<'EOF'
2
EOF
cat > "$CONTEST_DIR/problems/6/tests/2.in" <<'EOF'
3 3
EOF
cat > "$CONTEST_DIR/problems/6/tests/2.out" <<'EOF'
6
EOF
cat > "$CONTEST_DIR/problems/6/tests/3.in" <<'EOF'
3 7
EOF
cat > "$CONTEST_DIR/problems/6/tests/3.out" <<'EOF'
28
EOF
cat > "$CONTEST_DIR/problems/6/tests/4.in" <<'EOF'
10 10
EOF
cat > "$CONTEST_DIR/problems/6/tests/4.out" <<'EOF'
48620
EOF

# Problem 7 tests
cat > "$CONTEST_DIR/problems/7/tests/1.in" <<'EOF'
(()())
EOF
cat > "$CONTEST_DIR/problems/7/tests/1.out" <<'EOF'
YES
EOF
cat > "$CONTEST_DIR/problems/7/tests/2.in" <<'EOF'
())(
EOF
cat > "$CONTEST_DIR/problems/7/tests/2.out" <<'EOF'
NO
EOF
cat > "$CONTEST_DIR/problems/7/tests/3.in" <<'EOF'
(((((())))))
EOF
cat > "$CONTEST_DIR/problems/7/tests/3.out" <<'EOF'
YES
EOF
cat > "$CONTEST_DIR/problems/7/tests/4.in" <<'EOF'
)))(((
EOF
cat > "$CONTEST_DIR/problems/7/tests/4.out" <<'EOF'
NO
EOF

# Problem 8 tests
cat > "$CONTEST_DIR/problems/8/tests/1.in" <<'EOF'
5
2
3
4
5
17
EOF
cat > "$CONTEST_DIR/problems/8/tests/1.out" <<'EOF'
YES
YES
NO
YES
YES
EOF
cat > "$CONTEST_DIR/problems/8/tests/2.in" <<'EOF'
4
1
15
97
100
EOF
cat > "$CONTEST_DIR/problems/8/tests/2.out" <<'EOF'
NO
NO
YES
NO
EOF
cat > "$CONTEST_DIR/problems/8/tests/3.in" <<'EOF'
3
999983
999984
999985
EOF
cat > "$CONTEST_DIR/problems/8/tests/3.out" <<'EOF'
YES
NO
NO
EOF
cat > "$CONTEST_DIR/problems/8/tests/4.in" <<'EOF'
2
2147483647
2147483646
EOF
cat > "$CONTEST_DIR/problems/8/tests/4.out" <<'EOF'
YES
NO
EOF

echo "[5/5] Done."
echo "Generated fixture tree under: $SERVER_DATA_DIR"
