# OweNone

Single repository containing both backend and frontend apps.

## Repository Layout

```text
/
	backend/
	frontend/
```

## Git Remote via SSH

Use one SSH remote for the whole repo (both `backend/` and `frontend/` are pushed together).

1. Generate a key (if needed):

```powershell
ssh-keygen -t ed25519 -C "your_email@example.com"
```

2. Start the SSH agent and add your key:

```powershell
Get-Service ssh-agent | Set-Service -StartupType Automatic
Start-Service ssh-agent
ssh-add $env:USERPROFILE\.ssh\id_ed25519
```

3. Add your public key to GitHub:

```powershell
Get-Content $env:USERPROFILE\.ssh\id_ed25519.pub
```

Copy the output and add it at GitHub -> Settings -> SSH and GPG keys.

4. Set remote to SSH and push:

```powershell
git remote add origin git@github.com:<your-user>/<your-repo>.git
git branch -M main
git add .
git commit -m "Initial monorepo structure"
git push -u origin main
```

5. Verify SSH auth:

```powershell
ssh -T git@github.com
```