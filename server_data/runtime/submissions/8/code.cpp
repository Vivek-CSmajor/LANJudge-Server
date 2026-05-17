#include <bits/stdc++.h>
using namespace std;
const long long MOD=1000000007;
int main(){ int n,m; if(!(cin>>n>>m)) return 0; vector<long long> dp(m,1); for(int i=1;i<n;i++){ for(int j=1;j<m;j++){ dp[j]=(dp[j]+dp[j-1])%MOD; }} cout<<dp[m-1]<<"\n"; }
