# Runbook

## Convert trial user to paid user
```bash
sqlite3 jobseek.db "UPDATE users SET paid = 1 WHERE email = 'user@example.com';"
```

## Upgrade user to Pro plan
```bash
sqlite3 jobseek.db "UPDATE users SET subscription_plan = 'pro', paid = 1 WHERE email = 'user@example.com';"
```
