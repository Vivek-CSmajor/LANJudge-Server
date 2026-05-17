#include <bits/stdc++.h>
using namespace std;
bool isPrime(long long x){ if(x<2) return false; if(x%2==0) return x==2; for(long long i=3;i*i<=x;i+=2) if(x%i==0) return false; return true; }
int main(){ ios::sync_with_stdio(false); cin.tie(nullptr); int q; if(!(cin>>q)) return 0; while(q--){ long long x; cin>>x; cout<<(isPrime(x)?"YES":"NO")<<"\n"; } }
