# concurrent-imapsync

This tool executes mutiple imapsync instaces in its own go-routines.

```go
// Default values
imapsync --source accounts.txt --threads 4
```

```
// accounts.txt
source1.imap.com|source1@user.com|source1$password|target1.imap.com|target1@user.com|target1$password
source2.imap.com|source2@user.com|source2$password|target2.imap.com|target2@user.com|target2$password
...
```
