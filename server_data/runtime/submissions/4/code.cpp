#include <bits/stdc++.h>
using namespace std;
int main(){ string s; if(!(cin>>s)) return 0; int bal=0; for(char c:s){ if(c=='(') bal++; else bal--; if(bal<0){ cout<<"NO\n"; return 0; } } cout<<(bal==0?"YES":"NO")<<"\n"; }
